//go:build ignore

// Package handler 提供备忘录（Todo）相关的 HTTP API handlers
// =============================================
// 作用：
//   实现 RESTful API 接口，处理前端发来的备忘录 CRUD 请求。
//   包括：新增、查询、更新、删除待办任务，以及历史归档功能。
//
// API 接口列表：
//   GET    /api/todos?category=life          → 获取指定分类的所有待办任务
//   POST   /api/todos                        → 新增一条待办任务
//   PUT    /api/todos/:id                    → 更新指定任务（修改内容或状态）
//   DELETE /api/todos/:id                    → 删除指定任务
//   POST   /api/todos/archive                → 将当天所有任务归档到历史表
//   GET    /api/todos/history?date=2026-04-05 → 获取指定日期的历史归档
//   GET    /api/todos/history/dates          → 获取所有有归档数据的日期列表
//
// 数据流向：
//   前端 fetch() → Go HTTP handler → SQL 查询 → MySQL → JSON 响应 → 前端处理
//
// 安全措施：
//   - 所有 SQL 操作使用 PrepareStatement 防止 SQL 注入
//   - 输入数据长度校验
//   - 统一错误处理和响应格式
// =============================================

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tinyweb1/db"
	"tinyweb1/model"
)

// ============================================================
// 工具函数
// ============================================================

// sendJSON 发送 JSON 响应的辅助函数
// 将任意数据序列化为 JSON 并写入 http.ResponseWriter
// 同时设置 Content-Type: application/json 头部
func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// ============================================================
// GET /api/todos?category=xxx - 获取待办任务列表
// ============================================================

// GetTodos 获取指定用户和分类的待办任务列表
// Query 参数：
//   - category: 必填，任务分类（life/study/important）
//
// 返回示例：
//   {"code":0,"message":"success","data":[{"id":1,"category":"life","text":"买菜","done":false,...}]}
func GetTodos(w http.ResponseWriter, r *http.Request) {
	// 从 URL query 中获取 category 参数
	category := r.URL.Query().Get("category")
	if category == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "缺少 category 参数"))
		return
	}

	// 校验 category 取值范围
	validCategories := map[string]bool{"life": true, "study": true, "important": true}
	if !validCategories[category] {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "无效的 category 值，应为 life/study/important"))
		return
	}

	database := db.GetDB()
	// 使用参数化查询获取该分类下按排序的所有任务（按 sort_order 升序）
	rows, err := database.Query(
		"SELECT id, user_id, category, text, done, sort_order, created_at, updated_at FROM todos WHERE user_id = ? AND category = ? ORDER BY sort_order ASC, id ASC",
		"default", category,
	)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询失败"))
		return
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		var t model.Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Category, &t.Text, &t.Done, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt); err != nil {
			sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "数据解析失败"))
			return
		}
		todos = append(todos, t)
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(todos))
}

// ============================================================
// POST /api/todos - 新增待办任务
// ============================================================

// CreateTodo 创建一条新的待办任务
// Request Body (JSON)：
//   { "category": "life", "text": "买菜" }
//
// 成功响应：
//   {"code":0,"message":"success","data":{"id":5,"category":"life","text":"买菜",...}}
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req model.TodoCreateRequest
	// 解析请求体的 JSON 数据
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

	// 输入校验
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "任务内容不能为空"))
		return
	}
	if len(req.Text) > 200 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "任务内容不能超过200个字符"))
		return
	}
	validCategories := map[string]bool{"life": true, "study": true, "important": true}
	if !validCategories[req.Category] {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "无效的分类"))
		return
	}

	database := db.GetDB()
	// 插入新任务，sort_order 设为当前时间戳保证新任务排在后面
	result, err := database.Exec(
		"INSERT INTO todos (user_id, category, text, done, sort_order) VALUES (?, ?, ?, 0, ?)",
		"default", req.Category, req.Text, time.Now().Unix(),
	)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "创建失败"))
		return
	}

	// 获取新插入记录的自增 ID
	lastID, _ := result.LastInsertId()

	// 查询刚插入的完整记录用于返回
	var todo model.Todo
	err = database.QueryRow(
		"SELECT id, user_id, category, text, done, sort_order, created_at, updated_at FROM todos WHERE id = ?",
		lastID,
	).Scan(&todo.ID, &todo.UserID, &todo.Category, &todo.Text, &todo.Done, &todo.SortOrder, &todo.CreatedAt, &todo.UpdatedAt)

	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "创建成功但读取失败"))
		return
	}

	sendJSON(w, http.StatusCreated, model.SuccessResponse(todo))
}

// ============================================================
// PUT /api/todos/:id - 更新待办任务
// ============================================================

// UpdateTodo 更新指定的待办任务
// URL 参数：
//   - id: 任务 ID（路径参数）
//
// Request Body (JSON)：
//   { "text": "新内容" }  或  { "done": true }  或两者都有
//
// 支持部分更新：只更新请求中提供的字段
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	// 从 URL 路径中提取任务 ID
	id, err := extractTodoID(r)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "无效的任务 ID"))
		return
	}

	var req model.TodoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

	database := db.GetDB()

	// 动态构建 UPDATE 语句（根据请求中有哪些字段决定更新哪些列）
	if req.Text != nil {
		trimmed := strings.TrimSpace(*req.Text)
		if trimmed == "" {
			sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "任务内容不能为空"))
			return
		}
		if len(trimmed) > 200 {
			sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "任务内容不能超过200个字符"))
			return
		}
		_, err = database.Exec("UPDATE todos SET text = ? WHERE id = ? AND user_id = ?", trimmed, id, "default")
	} else if req.Done != nil {
		_, err = database.Exec("UPDATE todos SET done = ? WHERE id = ? AND user_id = ?", *req.Done, id, "default")
	} else {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "没有可更新的字段"))
		return
	}

	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "更新失败"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(nil))
}

// ============================================================
// DELETE /api/todos/:id - 删除待办任务
// ============================================================

// DeleteTodo 删除指定的待办任务
// URL 参数：
//   - id: 任务 ID（路径参数）
//
// 注意：软删除场景可改为 UPDATE SET deleted=1，当前实现物理删除
func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := extractTodoID(r)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "无效的任务 ID"))
		return
	}

	database := db.GetDB()
	result, err := database.Exec("DELETE FROM todos WHERE id = ? AND user_id = ?", id, "default")
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "删除失败"))
		return
	}

	// 检查是否真的删除了行（可能 ID 不存在）
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendJSON(w, http.StatusNotFound, model.ErrorResponse(404, "任务不存在"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(nil))
}

// ============================================================
// POST /api/todos/archive - 归档当天的任务
// ============================================================

// ArchiveTodos 将当前用户当天的所有待办任务复制到历史归档表中，然后清空原表中的这些记录
// 触发时机：前端每日首次加载时调用
//
// 流程：
//   1. 查询 todos 表中该用户的所有任务
//   2. 批量插入到 todo_history 表（archive_date 为今天）
//   3. 删除 todos 表中已归档的记录
//
// 返回：归档的任务数量
func ArchiveTodos(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	today := time.Now().Format("2006-01-02")

	// 开启事务，保证归档操作的原子性（要么全部成功，要么全部回滚）
	tx, err := database.Begin()
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "开启事务失败"))
		return
	}

	// 步骤1: 查询所有待归档的任务
	rows, err := tx.Query(
		"SELECT id, category, text, done FROM todos WHERE user_id = ? ORDER BY sort_order ASC",
		"default",
	)
	if err != nil {
		tx.Rollback()
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询待归档任务失败"))
		return
	}
	defer rows.Close()

	// 收集所有待归档的任务
	type taskToArchive struct {
		ID       int
		Category string
		Text     string
		Done     bool
	}
	var tasks []taskToArchive
	for rows.Next() {
		var t taskToArchive
		if err := rows.Scan(&t.ID, &t.Category, &t.Text, &t.Done); err != nil {
			tx.Rollback()
			sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "解析任务失败"))
			return
		}
		tasks = append(tasks, t)
	}

	if len(tasks) == 0 {
		tx.Rollback() // 无需归档
		sendJSON(w, http.StatusOK, model.SuccessResponse(map[string]int{"archived_count": 0}))
		return
	}

	// 步骤2: 批量插入到历史表
	stmt, err := tx.Prepare("INSERT INTO todo_history (user_id, archive_date, category, text, done) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "准备归档语句失败"))
		return
	}
	defer stmt.Close()

	for _, t := range tasks {
		if _, err := stmt.Exec("default", today, t.Category, t.Text, t.Done); err != nil {
			tx.Rollback()
			sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "归档写入失败"))
			return
		}
	}

	// 步骤3: 删除已归档的任务
	_, err = tx.Exec("DELETE FROM todos WHERE user_id = ?", "default")
	if err != nil {
		tx.Rollback()
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "清理原任务失败"))
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "提交归档事务失败"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(map[string]int{"archived_count": len(tasks)}))
}

// ============================================================
// GET /api/todos/history?date=2026-04-05 - 获取指定日期的历史归档
// ============================================================

// GetTodoHistoryByDate 获取指定日期的历史归档记录
// Query 参数：
//   - date: 归档日期（YYYY-MM-DD 格式），不传则返回今天的
//
// 返回数据按 category 分组，兼容前端的展示格式
func GetTodoHistoryByDate(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02") // 默认今天
	}

	database := db.GetDB()
	rows, err := database.Query(
		"SELECT category, text, done FROM todo_history WHERE user_id = ? AND archive_date = ? ORDER BY category",
		"default", date,
	)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询历史失败"))
		return
	}
	defer rows.Close()

	// 按 category 分组收集数据（兼容前端原有格式）
	result := make(map[string][]model.TodoItem)
	for rows.Next() {
		var category, text string
		var done bool
		if err := rows.Scan(&category, &text, &done); err != nil {
			continue
		}
		result[category] = append(result[category], model.TodoItem{Text: text, Done: done})
	}

	// 包装成带日期的结构体返回
	response := model.TodoHistoryByDate{
		Date:  date,
		Todos: result,
	}
	sendJSON(w, http.StatusOK, model.SuccessResponse(response))
}

// ============================================================
// GET /api/todos/history/dates - 获取有归档数据的日期列表
// ============================================================

// GetTodoHistoryDates 获取所有存在归档记录的日期列表
// 用于前端显示日历上的标记点
//
// 返回示例：
//   ["2026-04-01","2026-04-03","2026-04-05"]
func GetTodoHistoryDates(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	rows, err := database.Query(
		"SELECT DISTINCT archive_date FROM todo_history WHERE user_id = ? ORDER BY archive_date DESC",
		"default",
	)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询归档日期失败"))
		return
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err != nil {
			continue
		}
		dates = append(dates, date)
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(dates))
}

// ============================================================
// 内部工具函数
// ============================================================

// extractTodoID 从 URL 路径中提取任务 ID
// 路径格式：/api/todos/123
// 返回解析后的整型 ID 和可能的错误
func extractTodoID(r *http.Request) (int, error) {
	path := r.URL.Path
	// 截取最后一个 / 后面的部分作为 ID
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return 0, fmt.Errorf("invalid path format")
	}
	idStr := parts[len(parts)-1]
	return strconv.Atoi(idStr)
}
