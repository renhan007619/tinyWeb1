// Package handler 专注时间（Focus Time）相关的 HTTP API handlers
// =============================================
// 作用：
//   实现专注时间记录和统计的 RESTful API，帮助用户追踪学习/工作时长。
//
// API 接口列表：
//   POST /api/focus/session      → 创建一条专注记录
//   GET  /api/focus/today        → 获取今日专注统计
//   GET  /api/focus/summary      → 获取历史总览（每日统计）
//   GET  /api/focus/history      → 获取某日详细记录
//   POST /api/focus/tags         → 创建自定义标签
//   GET  /api/focus/tags         → 获取标签列表
//
// 数据流向：
//   前端 fetch() → Go HTTP handler → GORM 查询 → MySQL → JSON 响应 → 前端处理
//
// 当前用户处理：
//   暂时使用固定 user_id=1（defaultUser），后续接入 JWT 后从 token 获取
// =============================================

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tinyweb1/db"
	"tinyweb1/model"
)

// defaultUser 专注时间功能的默认用户ID
// 后续接入 JWT 认证后，会从 token 中获取真实的 user_id
const defaultUser uint = 1

// formatDuration 将秒数格式化为 "X小时Y分钟" 的可读字符串
func formatDuration(seconds int64) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时", hours)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d分钟", minutes)
	}
	return fmt.Sprintf("%d秒", seconds)
}

// ============================================================
// POST /api/focus/session - 创建专注记录
// ============================================================

// CreateFocusSession 创建一条新的专注记录
// Request Body (JSON)：
//   { "duration": 1500, "tag": "Go语言开发", "tag_color": "#FF6B6B" }
//
// 成功响应：
//   {"code":0,"message":"success","data":{...}}
func CreateFocusSession(w http.ResponseWriter, r *http.Request) {
	var req model.CreateFocusSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

	// 校验专注时长
	if req.Duration < 60 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "专注时长最少1分钟"))
		return
	}
	if req.Duration > 14400 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "专注时长最多4小时"))
		return
	}

	// 设置默认标签
	if req.Tag == "" {
		req.Tag = "未分类"
	}
	if req.TagColor == "" {
		req.TagColor = "#6C5CE7"
	}

	now := time.Now()
	session := model.StudySession{
		UserID:    defaultUser,
		Duration:  req.Duration,
		Date:      now.Format("2006-01-02"),
		StartedAt: now,
		Tag:       req.Tag,
		TagColor:  req.TagColor,
	}

	database := db.GetDB()
	if err := database.Create(&session).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "创建专注记录失败"))
		return
	}

	sendJSON(w, http.StatusCreated, model.SuccessResponse(session))
}

// ============================================================
// GET /api/focus/today - 获取今日专注统计
// ============================================================

// GetTodayFocus 获取今日的专注时间统计
// 返回今日总时长、专注次数、按标签分组的统计
func GetTodayFocus(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	today := time.Now().Format("2006-01-02")

	// 查询今日所有专注记录
	var sessions []model.StudySession
	if err := database.Where("user_id = ? AND date = ?", defaultUser, today).Find(&sessions).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询今日专注数据失败"))
		return
	}

	// 计算总秒数和次数
	var totalSeconds int64
	tagMap := make(map[string]*model.TagSummary) // tag名 → 汇总

	for _, s := range sessions {
		totalSeconds += int64(s.Duration)

		if ts, ok := tagMap[s.Tag]; ok {
			ts.Seconds += int64(s.Duration)
		} else {
			tagMap[s.Tag] = &model.TagSummary{
				Tag:   s.Tag,
				Color: s.TagColor,
			}
			tagMap[s.Tag].Seconds += int64(s.Duration)
		}
	}

	// 计算百分比，组装 byTag 列表
	var byTag []model.TagSummary
	for _, ts := range tagMap {
		if totalSeconds > 0 {
			ts.Percentage = float64(ts.Seconds) / float64(totalSeconds) * 100
		}
		byTag = append(byTag, *ts)
	}

	response := model.TodayFocusResponse{
		TotalSeconds:   totalSeconds,
		TotalFormatted: formatDuration(totalSeconds),
		SessionCount:   int64(len(sessions)),
		ByTag:          byTag,
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(response))
}

// ============================================================
// GET /api/focus/summary - 获取历史总览
// ============================================================

// GetFocusSummary 获取历史专注时间总览
// Query 参数：
//   - days: 查询最近多少天的统计（默认30天）
func GetFocusSummary(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()

	// 历史总计
	var totalSeconds int64
	var totalSessions int64
	database.Model(&model.StudySession{}).Where("user_id = ?", defaultUser).
		Select("COALESCE(SUM(duration), 0)").Scan(&totalSeconds)
	database.Model(&model.StudySession{}).Where("user_id = ?", defaultUser).
		Count(&totalSessions)

	// 每日统计
	days := parseIntQueryParam(r, "days", 30)
	var dailyStats []model.DailyStatItem

	rows, err := database.Model(&model.StudySession{}).
		Select("date, SUM(duration) as total_seconds, COUNT(*) as session_count").
		Where("user_id = ? AND date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)", defaultUser, days).
		Group("date").
		Order("date DESC").
		Rows()

	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询历史统计失败"))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item model.DailyStatItem
		var secs int64
		var count int64
		if err := rows.Scan(&item.Date, &secs, &count); err != nil {
			continue
		}
		item.TotalSeconds = secs
		item.TotalFormatted = formatDuration(secs)
		item.SessionCount = count
		dailyStats = append(dailyStats, item)
	}

	response := model.FocusSummaryResponse{
		TotalSeconds:   totalSeconds,
		TotalFormatted: formatDuration(totalSeconds),
		TotalSessions:  totalSessions,
		DailyStats:     dailyStats,
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(response))
}

// ============================================================
// GET /api/focus/history - 获取某日详细记录
// ============================================================

// GetFocusHistory 获取指定日期的专注记录详情
// Query 参数：
//   - date: 日期（YYYY-MM-DD），默认今天
func GetFocusHistory(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	database := db.GetDB()
	var sessions []model.StudySession
	if err := database.Where("user_id = ? AND date = ?", defaultUser, date).
		Order("started_at ASC").Find(&sessions).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询专注记录失败"))
		return
	}

	// 转换为详情格式
	var details []model.StudySessionDetail
	for _, s := range sessions {
		details = append(details, model.StudySessionDetail{
			ID:        s.ID,
			Duration:  s.Duration,
			Tag:       s.Tag,
			TagColor:  s.TagColor,
			StartedAt: s.StartedAt.Format("15:04:05"),
		})
	}

	response := model.FocusHistoryResponse{
		Date:     date,
		Sessions: details,
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(response))
}

// ============================================================
// POST /api/focus/tags - 创建标签
// ============================================================

// CreateTag 创建一个自定义专注标签
// Request Body (JSON)：
//   { "name": "Go语言开发", "color": "#FF6B6B" }
func CreateTag(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "标签名不能为空"))
		return
	}
	if len(req.Name) > 50 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "标签名不能超过50个字符"))
		return
	}

	if req.Color == "" {
		req.Color = "#6C5CE7"
	}

	tag := model.StudyTag{
		UserID: defaultUser,
		Name:   req.Name,
		Color:  req.Color,
	}

	database := db.GetDB()
	if err := database.Create(&tag).Error; err != nil {
		// 检查是否是唯一约束冲突（同名标签已存在）
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "unique") {
			sendJSON(w, http.StatusConflict, model.ErrorResponse(409, "标签名已存在"))
			return
		}
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "创建标签失败"))
		return
	}

	sendJSON(w, http.StatusCreated, model.SuccessResponse(tag))
}

// ============================================================
// GET /api/focus/tags - 获取标签列表
// ============================================================

// GetTags 获取当前用户的所有自定义标签
func GetTags(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	var tags []model.StudyTag
	if err := database.Where("user_id = ?", defaultUser).Order("created_at ASC").Find(&tags).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询标签失败"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(tags))
}
