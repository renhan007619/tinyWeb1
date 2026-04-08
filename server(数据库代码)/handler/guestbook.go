//go:build ignore

// Package handler 提供留言板（Guestbook）相关的 HTTP API handlers
// =============================================
// 作用：
//   实现留言板功能的 REST API，支持访客发布留言和查看留言列表。
//   这是本次新增的功能模块，对应前端右侧栏底部的留言板卡片 UI。
//
// API 接口列表：
//   GET  /api/guestbook?page=1&size=20  → 获取留言列表（分页，最新在前）
//   POST /api/guestbook                 → 发布新留言
//
// 功能特性：
//   - 分页查询：支持自定义每页数量，默认20条
//   - 匿名留言：昵称为选填，不填则显示"匿名访客"
//   - 内容过滤：自动去除 HTML 标签防止 XSS 攻击
//   - 长度限制：昵称最长64字符，留言内容最长500字符
//   - 时间排序：最新发布的留言显示在最前面
//
// 安全措施：
//   - HTML 标签过滤（防止 XSS 跨站脚本攻击）
//   - 内容长度校验（防止超长数据）
//   - SQL 参数化查询（防止 SQL 注入）
// =============================================

package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"tinyweb1/db"
	"tinyweb1/model"
)

// ============================================================
// GET /api/guestbook - 获取留言列表（分页）
// ============================================================

// GetGuestbookMessages 获取留言板留言列表，支持分页
// Query 参数：
//   - page: 页码（从1开始，默认第1页）
//   - size: 每页条数（默认20条，最大100条）
//
// 成功响应示例：
//   {
//     "code": 0,
//     "message": "success",
//     "data": {
//       "list": [{"id":1,"nickname":"张三","content":"你好！","created_at":"..."}],
//       "total": 50,
//       "page": 1,
//       "size": 20,
//       "total_pages": 3
//     }
//   }
//
// 分页逻辑：
//   - 最新发布的留言在第一页最前面（ORDER BY id DESC）
//   - total_pages 通过 math.Ceil(total / size) 计算
func GetGuestbookMessages(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()

	// 解析分页参数，带默认值
	page := parseIntQueryParam(r, "page", 1)
	size := parseIntQueryParam(r, "size", 20)

	// 参数范围限制
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100 // 单次最多拉取100条
	}

	// 查询总数（用于计算总页数）
	var total int64
	err := database.QueryRow("SELECT COUNT(*) FROM guestbook").Scan(&total)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询留言总数失败"))
		return
	}

	// 计算偏移量（OFFSET）
	offset := (page - 1) * size

	// 分页查询留言列表（按 ID 降序 = 最新在前）
	rows, err := database.Query(
		"SELECT id, nickname, content, created_at FROM guestbook ORDER BY id DESC LIMIT ? OFFSET ?",
		size, offset,
	)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询留言列表失败"))
		return
	}
	defer rows.Close()

	// 遍历结果集构建列表
	var messages []model.Guestbook
	for rows.Next() {
		var m model.Guestbook
		if err := rows.Scan(&m.ID, &m.Nickname, &m.Content, &m.CreatedAt); err != nil {
			continue
		}
		messages = append(messages, m)
	}

	// 计算总页数
	totalPages := int(total) / size
	if int(total)%size > 0 {
		totalPages++
	}

	// 组装分页响应
	response := model.GuestbookListResponse{
		List:       messages,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(response))
}

// ============================================================
// POST /api/guestbook - 发布新留言
// ============================================================

// CreateGuestbookMessage 发布一条新的留言
// Request Body (JSON)：
//   {
//     "nickname": "张三",    // 可选，不填则为匿名
//     "content": "网站做得真棒！"  // 必填，留言内容
//   }
//
// 成功响应（201 Created）：
//   {"code":0,"message":"success","data":{"id":10,"nickname":"张三","content":"...","created_at":"..."}}
//
// 数据处理流程：
//   1. 解析和校验 JSON 请求数据
//   2. 过滤 HTML 标签（防 XSS）
//   3. 校验内容和长度限制
//   4. 写入数据库
//   5. 返回完整的留言记录（含自增 ID 和时间戳）
func CreateGuestbookMessage(w http.ResponseWriter, r *http.Request) {
	var req model.GuestbookCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

	// 清理和校验昵称
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" {
		req.Nickname = "" // 保持为空，前端会显示"匿名访客"
	} else if len(req.Nickname) > 64 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "昵称不能超过64个字符"))
		return
	}

	// 清理和校验留言内容
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "留言内容不能为空"))
		return
	}
	// 过滤 HTML 标签，防止 XSS 攻击
	req.Content = stripHTMLTags(req.Content)
	if len(req.Content) > 500 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "留言内容不能超过500个字符"))
		return
	}
	// 二次清理后再检查（可能过滤标签后变空）
	if req.Content == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "留言内容不能为空"))
		return
	}

	database := db.GetDB()

	// 插入新留言
	result, err := database.Exec(
		"INSERT INTO guestbook (nickname, content) VALUES (?, ?)",
		req.Nickname, req.Content,
	)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "发布留言失败"))
		return
	}

	// 获取新插入记录的 ID
	lastID, _ := result.LastInsertId()

	// 查询完整记录用于返回
	var msg model.Guestbook
	err = database.QueryRow(
		"SELECT id, nickname, content, created_at FROM guestbook WHERE id = ?",
		lastID,
	).Scan(&msg.ID, &msg.Nickname, &msg.Content, &msg.CreatedAt)

	if err != nil {
		// 插入成功但查询失败（极端情况），仍返回成功
		msg = model.Guestbook{
			ID:        int(lastID),
			Nickname:  req.Nickname,
			Content:   req.Content,
			CreatedAt: time.Now(),
		}
	}

	sendJSON(w, http.StatusCreated, model.SuccessResponse(msg))
}

// ============================================================
// 内部工具函数
// ============================================================

// parseIntQueryParam 从 URL Query 中解析整数参数
// 如果参数不存在或解析失败，返回默认值 defaultValue
func parseIntQueryParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	result := 0
	// 手动实现简单的字符串转整数
	isNegative := false
	for i, c := range value {
		if i == 0 && c == '-' {
			isNegative = true
			continue
		}
		if c < '0' || c > '9' {
			return defaultValue
		}
		result = result*10 + int(c-'0')
	}
	if isNegative {
		result = -result
	}
	return result
}

// stripHTMLTags 去除字符串中的所有 HTML 标签
// 防止 XSS（跨站脚本攻击）：用户输入 <script>alert('xss')</script> 会被过滤为 alert('xss')
// 实现原理：状态机遍历，遇到 '<' 进入跳过模式，直到 '>' 结束跳过
func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false // 标记是否正在跳过 HTML 标签内部的内容
	for _, c := range s {
		if c == '<' {
			inTag = true // 遇到 < 开始跳过
			continue
		}
		if c == '>' && inTag {
			inTag = false // 遇到 > 结束跳过
			continue
		}
		if !inTag {
			result.WriteRune(c) // 非标签内的字符保留
		}
	}
	return result.String()
}
