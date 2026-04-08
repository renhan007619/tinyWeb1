//go:build ignore

// Package handler 提供用户设置（Settings）相关的 HTTP API handler
// =============================================
// 作用：
//   处理用户个性化设置的读写请求，当前仅支持主题偏好（light/dark）的存储。
//   将原本保存在 localStorage 中的主题设置迁移到服务端数据库，
//   实现多设备间的设置同步。
//
// API 接口列表：
//   GET  /api/settings/theme  → 获取当前用户的主题偏好
//   PUT  /api/settings/theme  → 更新主题偏好
//
// 设计说明：
//   - 使用 user_id 作为主键（当前固定为 "default"），每个用户一行记录
//   - 更新操作使用 UPSERT 语义（INSERT ... ON DUPLICATE KEY UPDATE）
//   - 如果用户从未设置过主题，首次读取返回默认值 "light"
// =============================================

package handler

import (
	"encoding/json"
	"net/http"

	"tinyweb1/db"
	"tinyweb1/model"
)

// ============================================================
// GET /api/settings/theme - 获取主题偏好
// ============================================================

// GetTheme 获取当前用户的主题偏好设置
// 无需请求参数，从 Authorization 或默认 user_id 识别用户
//
// 成功响应示例：
//   {"code":0,"message":"success","data":{"user_id":"default","theme":"dark","updated_at":"..."}}
//
// 特殊情况：
//   - 如果用户从未设置过主题，返回默认值 "light"
//   - 数据库查询失败时返回错误信息
func GetTheme(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()

	var theme string
	// 查询该用户的主题设置
	err := database.QueryRow("SELECT theme FROM settings WHERE user_id = ?", "default").Scan(&theme)

	if err != nil {
		// 如果没有找到记录（sql.ErrNoRows），返回默认主题 light
		// 其他错误则返回 500
		if err.Error() == "sql: no rows in result set" {
			theme = "light" // 默认亮色主题
		} else {
			sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询主题设置失败"))
			return
		}
	}

	setting := model.Setting{
		UserID: "default",
		Theme:  theme,
	}
	sendJSON(w, http.StatusOK, model.SuccessResponse(setting))
}

// ============================================================
// PUT /api/settings/theme - 更新主题偏好
// ============================================================

// UpdateTheme 更新当前用户的主题偏好设置
// Request Body (JSON)：
//   { "theme": "dark" }
//
// 有效值：
//   - "light": 亮色主题（白底黑字）
//   - "dark": 暗色主题（深色背景，护眼模式）
//
// 成功响应：
//   {"code":0,"message":"success","data":null}
//
// 实现细节：
//   使用 MySQL 的 ON DUPLICATE KEY UPDATE 实现 Upsert：
//   - 如果记录不存在 → INSERT 新记录
//   - 如果记录已存在 → UPDATE 已有记录的 theme 字段
func UpdateTheme(w http.ResponseWriter, r *http.Request) {
	var req model.ThemeUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

	// 校验主题值的有效性
	req.Theme = trimString(req.Theme)
	if req.Theme != "light" && req.Theme != "dark" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "无效的主题值，应为 light 或 dark"))
		return
	}

	database := db.GetDB()

	// 使用 UPSERT 语义：有则更新，无则插入
	_, err := database.Exec(`
		INSERT INTO settings (user_id, theme) VALUES (?, ?)
		ON DUPLICATE KEY UPDATE theme = ?, updated_at = CURRENT_TIMESTAMP
	`, "default", req.Theme, req.Theme)

	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "保存主题设置失败"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(nil))
}

// ============================================================
// 内部工具函数
// ============================================================

// trimString 去除字符串首尾空格的辅助函数
func trimString(s string) string {
	// Go 的 strings.TrimSpace 实际实现
	result := make([]byte, len(s))
	i := 0
	started := false
	for j := 0; j < len(s); j++ {
		c := s[j]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if started {
				result[i] = c
				i++
			}
		} else {
			started = true
			result[i] = c
			i++
		}
	}
	// 去除末尾空格
	for i > 0 && (result[i-1] == ' ' || result[i-1] == '\t' || result[i-1] == '\n' || result[i-1] == '\r') {
		i--
	}
	return string(result[:i])
}
