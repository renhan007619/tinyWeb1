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
//   前端 fetch() → Go HTTP handler → GORM 查询 → MySQL → JSON 响应 → 前端处理
//
// 迁移说明（2026-04-15）：
//   - 从 database/sql 迁移到 GORM
//   - 移除 //go:build ignore 标签，重新启用此文件
//   - 所有数据库操作改用 GORM 风格
// =============================================

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"tinyweb1/db"
	"tinyweb1/model"
)

// todoUser 备忘录功能的默认用户标识
// 当前固定为 "default"，预留多用户扩展
const todoUser = "default"

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
	category := r.URL.Query().Get("category")
	if category == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "缺少 category 参数"))
		return
	}

	validCategories := map[string]bool{"life": true, "study": true, "important": true}
	if !validCategories[category] {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "无效的 category 值，应为 life/study/important"))
		return
	}

	database := db.GetDB()
	var todos []model.Todo
	if err := database.Where("user_id = ? AND category = ?", todoUser, category).
		Order("sort_order ASC, id ASC").Find(&todos).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询失败"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(todos))
}

// ============================================================
// POST /api/todos - 新增待办任务
// ============================================================

// CreateTodo 创建一条新的待办任务
// Request Body (JSON)：
//   { "category": "life", "text": "买菜" }
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req model.TodoCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

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
	todo := model.Todo{
		UserID:    todoUser,
		Category:  req.Category,
		Text:      req.Text,
		Done:      false,
		SortOrder: int(time.Now().Unix()),
	}

	if err := database.Create(&todo).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "创建失败"))
		return
	}

	sendJSON(w, http.StatusCreated, model.SuccessResponse(todo))
}

// ============================================================
// PUT /api/todos/:id - 更新待办任务
// ============================================================

// UpdateTodo 更新指定的待办任务
// 支持部分更新：只更新请求中提供的字段
// Request Body (JSON)：
//   { "text": "新内容" }  或  { "done": true }  或两者都有
func UpdateTodo(w http.ResponseWriter, r *http.Request) {
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

	// 先查询记录是否存在
	var todo model.Todo
	if err := database.Where("id = ? AND user_id = ?", id, todoUser).First(&todo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			sendJSON(w, http.StatusNotFound, model.ErrorResponse(404, "任务不存在"))
		} else {
			sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询失败"))
		}
		return
	}

	// 动态构建更新字段
	updates := make(map[string]interface{})
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
		updates["text"] = trimmed
	}
	if req.Done != nil {
		updates["done"] = *req.Done
	}

	if len(updates) == 0 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "没有可更新的字段"))
		return
	}

	if err := database.Model(&todo).Updates(updates).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "更新失败"))
		return
	}

	// 重新查询返回最新数据
	database.Where("id = ?", id).First(&todo)
	sendJSON(w, http.StatusOK, model.SuccessResponse(todo))
}

// ============================================================
// DELETE /api/todos/:id - 删除待办任务
// ============================================================

// DeleteTodo 删除指定的待办任务
func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := extractTodoID(r)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "无效的任务 ID"))
		return
	}

	database := db.GetDB()
	result := database.Where("id = ? AND user_id = ?", id, todoUser).Delete(&model.Todo{})
	if result.Error != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "删除失败"))
		return
	}
	if result.RowsAffected == 0 {
		sendJSON(w, http.StatusNotFound, model.ErrorResponse(404, "任务不存在"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(nil))
}

// ============================================================
// POST /api/todos/archive - 归档当天的任务
// ============================================================

// ArchiveTodos 将当前用户的所有待办任务复制到历史归档表中，然后清空原表中的这些记录
// 触发时机：前端每日首次加载时调用
func ArchiveTodos(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	today := time.Now().Format("2006-01-02")

	// 查询所有待归档的任务
	var tasks []model.Todo
	if err := database.Where("user_id = ?", todoUser).Order("sort_order ASC").Find(&tasks).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询待归档任务失败"))
		return
	}

	if len(tasks) == 0 {
		sendJSON(w, http.StatusOK, model.SuccessResponse(map[string]int{"archived_count": 0}))
		return
	}

	// 开启事务
	tx := database.Begin()
	if tx.Error != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "开启事务失败"))
		return
	}

	// 批量插入到历史表
	for _, t := range tasks {
		history := model.TodoHistory{
			UserID:      t.UserID,
			ArchiveDate: today,
			Category:    t.Category,
			Text:        t.Text,
			Done:        t.Done,
		}
		if err := tx.Create(&history).Error; err != nil {
			tx.Rollback()
			sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "归档写入失败"))
			return
		}
	}

	// 删除已归档的任务
	if err := tx.Where("user_id = ?", todoUser).Delete(&model.Todo{}).Error; err != nil {
		tx.Rollback()
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "清理原任务失败"))
		return
	}

	if err := tx.Commit().Error; err != nil {
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
func GetTodoHistoryByDate(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	database := db.GetDB()
	var histories []model.TodoHistory
	if err := database.Where("user_id = ? AND archive_date = ?", todoUser, date).
		Order("category ASC").Find(&histories).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询历史失败"))
		return
	}

	// 按 category 分组收集数据（兼容前端原有格式）
	result := make(map[string][]model.TodoItem)
	for _, h := range histories {
		result[h.Category] = append(result[h.Category], model.TodoItem{Text: h.Text, Done: h.Done})
	}

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
// 返回示例：["2026-04-01","2026-04-03","2026-04-05"]
func GetTodoHistoryDates(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()

	var dates []string
	rows, err := database.Model(&model.TodoHistory{}).
		Select("DISTINCT archive_date").
		Where("user_id = ?", todoUser).
		Order("archive_date DESC").
		Rows()
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询归档日期失败"))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			continue
		}
		dates = append(dates, d)
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(dates))
}

// ============================================================
// 内部工具函数
// ============================================================

// extractTodoID 从 URL 路径中提取任务 ID
// 路径格式：/api/todos/123
func extractTodoID(r *http.Request) (int, error) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return 0, fmt.Errorf("invalid path format")
	}
	idStr := parts[len(parts)-1]
	return strconv.Atoi(idStr)
}
