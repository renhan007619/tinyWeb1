// Package handler 提供管理员查询访客列表功能
// =============================================
// 新增接口：GET /api/admin/visitors
// 作用：查询所有独立访客的详细信息
// 需要管理员权限
// =============================================

package handler

import (
	"net/http"

	"tinyweb1/db"
	"tinyweb1/middleware"
	"tinyweb1/model"
)

// GetAllVisitors 获取所有访客列表（管理员接口）
// URL: GET /api/admin/visitors
// 权限：仅管理员可访问
//
// 权限验证流程：
//  1. 从请求中获取用户角色（通过JWT token）
//  2. 验证角色是否为 "admin"
//  3. 非管理员返回 403 Forbidden
//
// 成功响应示例：
//
//	{
//	  "code": 0,
//	  "message": "success",
//	  "data": {
//	    "total": 53,
//	    "visitors": [
//	      {
//	        "visitor_ip": "192.168.1.100",
//	        "visit_count": 5,
//	        "device_type": "desktop",
//	        "browser": "Chrome",
//	        "os": "Windows",
//	        "last_visit_at": "2026-05-09 11:24:40"
//	      }
//	    ]
//	  }
//	}
func GetAllVisitors(w http.ResponseWriter, r *http.Request) {
	// ===== 权限检查开始 =====
	// 从 context 获取用户角色（由 AuthMiddleware 注入）
	role, ok := middleware.GetRole(r.Context())
	if !ok {
		sendJSON(w, http.StatusUnauthorized, model.ErrorResponse(401, "未登录"))
		return
	}

	// 检查是否为管理员
	if role != "admin" {
		sendJSON(w, http.StatusForbidden, model.ErrorResponse(403, "权限不足，需要管理员权限"))
		return
	}
	// ===== 权限检查结束 =====

	database := db.GetDB()

	var visitors []model.VisitStats
	result := database.Order("visit_count DESC").Find(&visitors)
	if result.Error != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询访客列表失败"))
		return
	}

	// 格式化响应数据
	type VisitorInfo struct {
		VisitorIP   string `json:"visitor_ip"`
		VisitCount  int    `json:"visit_count"`
		DeviceType  string `json:"device_type"`
		Browser     string `json:"browser"`
		OS          string `json:"os"`
		LastVisitAt string `json:"last_visit_at"`
	}

	var visitorList []VisitorInfo
	for _, v := range visitors {
		visitorList = append(visitorList, VisitorInfo{
			VisitorIP:   v.VisitorIP,
			VisitCount:  v.VisitCount,
			DeviceType:  v.DeviceType,
			Browser:     v.Browser,
			OS:          v.OS,
			LastVisitAt: v.LastVisitAt.Format("2006-01-02 15:04:05"),
		})
	}

	responseData := map[string]interface{}{
		"total":    len(visitorList),
		"visitors": visitorList,
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(responseData))
}
