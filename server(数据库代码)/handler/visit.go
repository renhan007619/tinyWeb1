// Package handler 提供访问统计（Visit Stats）相关的 HTTP API handlers
// =============================================
// 作用：
//   实现访客访问记录和统计的 RESTful API，用于追踪网站访问情况。
//
// API 接口列表：
//   POST /api/visit        → 记录一次访问（首次插入 / 已存在则更新次数）
//   GET  /api/visit/stats  → 获取访问统计汇总
//
// 核心逻辑（Upsert 模式）：
//   - 收到请求后根据 visitor_ip 查询数据库
//   - 如果该 IP 首次访问 → INSERT 新记录
//   - 如果该 IP 已存在 → UPDATE visit_count +1, 更新 last_visit_at
//   - 同时更新设备信息（浏览器、OS 等）
// =============================================

package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"tinyweb1/db"
	"tinyweb1/model"

	"gorm.io/gorm"
)

// ============================================================
// 工具函数：从 HTTP 请求中获取客户端真实 IP
// ============================================================

// getClientIP 按优先级从请求中提取客户端 IP：
//   1. X-Forwarded-For（经过反向代理时，取第一个）
//   2. X-Real-IP（Nginx 等设置）
//   3. RemoteAddr（直连地址，去掉端口号）
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	addr := r.RemoteAddr
	if lastColon := strings.LastIndex(addr, ":"); lastColon != -1 {
		return addr[:lastColon]
	}
	return addr
}

// ============================================================
// POST /api/visit - 记录一次访问
// ============================================================

// RecordVisit 记录一次页面访问
// 使用 GORM 的 Upsert 语义（Clause + OnConflict）实现"有则更新，无则插入"
//
// Request Body (JSON)：
//   {
//     "visitor_ip": "192.168.1.100",     // 必填，访客 IP
//     "user_agent": "Mozilla/5.0 ...",   // 可选，浏览器 UA
//     "device_type": "desktop",          // 可选，设备类型
//     "browser": "Chrome",               // 可选，浏览器名称
//     "os": "Windows",                   // 可选，操作系统
//     "referrer": "https://google.com"    // 可选，来源页面
//   }
//
// 成功响应（200 OK）：
//   {"code":0,"message":"success","data":{"is_first_visit":false,"visit_count":3}}
//
// 业务流程：
//   1. 解析 JSON 请求体，校验必填字段 visitor_ip
//   2. 在 visit_stats 表中查找该 IP 的现有记录
//   3. 如果找到 → 更新 visit_count+1、last_visit_at、扩展字段
//   4. 如果没找到 → 插入新记录，visit_count=1
func RecordVisit(w http.ResponseWriter, r *http.Request) {
	var req model.VisitRecord

	// 解析请求体 JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

	// 校验必填字段：visitor_ip
	req.VisitorIP = trimString(req.VisitorIP)
	if req.VisitorIP == "" {
		// 前端无法获取真实 IP，由后端从 HTTP 请求头自动提取
		req.VisitorIP = getClientIP(r)
	}
	if req.VisitorIP == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "无法获取访客 IP 地址"))
		return
	}

	database := db.GetDB()
	now := time.Now()
	isFirstVisit := false

	// ---- GORM Upsert 操作 ----
	// 尝试根据 visitor_ip 查找已有记录
	var existing model.VisitStats
	result := database.Where("visitor_ip = ?", req.VisitorIP).First(&existing)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 该 IP 是首次访问 → 创建新记录
			isFirstVisit = true
			newRecord := model.VisitStats{
				VisitorIP:    req.VisitorIP,
				VisitCount:   1,
				FirstVisitAt: now,
				LastVisitAt:  now,
				UserAgent:    req.UserAgent,
				DeviceType:   req.DeviceType,
				Browser:      req.Browser,
				OS:           req.OS,
				Referrer:     req.Referrer,
			}
			if err := database.Create(&newRecord).Error; err != nil {
				sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "创建访问记录失败: "+err.Error()))
				return
			}
		} else {
			// 数据库查询错误
			sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询访问记录失败"))
			return
		}
	} else {
		// 该 IP 已访问过 → 更新记录（访问次数 +1，更新最后访问时间等）
		updateErr := database.Model(&existing).Updates(map[string]interface{}{
			"visit_count":  gorm.Expr("visit_count + 1"),
			"last_visit_at": now,
			"user_agent":    req.UserAgent,
			"device_type":   req.DeviceType,
			"browser":       req.Browser,
			"os":            req.OS,
			"referrer":      req.Referrer,
		}).Error

		if updateErr != nil {
			sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "更新访问记录失败"))
			return
		}
	}

	// 返回响应
	responseData := map[string]interface{}{
		"is_first_visit": isFirstVisit,
		"visit_count":    existing.VisitCount + 1, // 更新后的计数
	}
	if isFirstVisit {
		responseData["visit_count"] = 1 // 首次访问，当前计数为 1
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(responseData))
}

// ============================================================
// GET /api/visit/stats - 获取访问统计汇总
// ============================================================

// GetVisitStats 获取网站的访问统计数据
// 无需参数，直接查询 visit_stats 表计算汇总指标
//
// 成功响应示例：
//   {
//     "code": 0,
//     "message": "success",
//     "data": {
//       "total_visits": 150,
//       "unique_visitors": 42,
//       "last_visit_at": "2026-04-08 18:50:00"
//     }
//   }
//
// 统计说明：
//   - total_visits: 所有访客的 visit_count 累加和（总点击量）
//   - unique_visitors: 不同 visitor_ip 的数量（独立访客数）
//   - last_visit_at: 最近一次访问的时间戳
func GetVisitStats(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()

	var stats model.VisitStatsResponse

	// 统计 1：总访问次数 = SUM(所有记录的 visit_count)
	database.Model(&model.VisitStats{}).Select("COALESCE(SUM(visit_count), 0)").Scan(&stats.TotalVisits)

	// 统计 2：独立访客数 = COUNT(DISTINCT visitor_ip)
	database.Model(&model.VisitStats{}).Select("COUNT(*)").Scan(&stats.UniqueVisitors)

	// 统计 3：最后访问时间 = MAX(last_visit_at)
	var lastVisit *time.Time
	database.Model(&model.VisitStats{}).Select("MAX(last_visit_at)").Scan(&lastVisit)
	if lastVisit != nil {
		formatted := lastVisit.Format("2006-01-02 15:04:05")
		stats.LastVisitAt = &formatted
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(stats))
}
