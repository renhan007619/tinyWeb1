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
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"tinyweb1/db"
	"tinyweb1/middleware"
	"tinyweb1/model"

	"gorm.io/gorm"
)

// getUserID 从 JWT 认证中间件注入的 context 中提取当前登录用户 ID
// 如果取不到（理论上不会，因为路由已加 AuthMiddleware），返回 0 并报错
func getUserID(r *http.Request) uint {
	id, ok := middleware.GetUserID(r.Context())
	if !ok {
		return 0 // 不应该发生，路由层已拦截未认证请求
	}
	return id
}

// calculateDiscountRate 计算碎片折扣率
// 满30分钟：100%（全额）
// 不满30分钟：按分钟数/30的比例，从60%线性增加到100%
// 例如：3分钟=60%, 15分钟=80%, 30分钟=100%
func calculateDiscountRate(totalMinutes float64) float64 {
	if totalMinutes >= 30 {
		return 1.0 // 满30分钟全额
	}
	if totalMinutes <= 3 {
		return 0.6 // 最少60%
	}
	// 线性插值：3分钟60% -> 30分钟100%
	// (分钟数 - 3) / (30 - 3) * (1.0 - 0.6) + 0.6
	return (totalMinutes-3)/27*0.4 + 0.6
}

// formatDiscount 格式化折扣显示（如 "60%", "85%", "100%"）
func formatDiscount(rate float64) string {
	percent := int(rate * 100)
	return fmt.Sprintf("%d%%", percent)
}

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
//
//	{ "duration": 1500, "tag": "Go语言开发", "tag_color": "#FF6B6B" }
//
// 成功响应：
//
//	{"code":0,"message":"success","data":{...}}
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
	} else {
		req.Tag = stripHTMLTags(req.Tag)
	}

	if req.TagColor == "" {
		req.TagColor = "#6C5CE7"
	}

	now := time.Now()
	session := model.StudySession{
		UserID:    getUserID(r),
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
	if err := database.Where("user_id = ? AND date = ?", getUserID(r), today).Find(&sessions).Error; err != nil {
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
//   - start: 开始日期（YYYY-MM-DD），与end配合使用
//   - end: 结束日期（YYYY-MM-DD），与start配合使用
//
// 注意：如果提供了start和end，则days参数被忽略
func GetFocusSummary(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()

	// 解析日期范围参数
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")
	useDateRange := startDate != "" && endDate != ""

	// 历史总计（如果是日期范围，则只统计该范围内的总计）
	var totalSeconds int64
	var totalSessions int64

	if useDateRange {
		// 验证日期格式
		if _, err := time.Parse("2006-01-02", startDate); err != nil {
			sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "开始日期格式错误，请使用YYYY-MM-DD格式"))
			return
		}
		if _, err := time.Parse("2006-01-02", endDate); err != nil {
			sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "结束日期格式错误，请使用YYYY-MM-DD格式"))
			return
		}

		// 查询指定日期范围内的总计
		database.Model(&model.StudySession{}).
			Where("user_id = ? AND date >= ? AND date <= ?", getUserID(r), startDate, endDate).
			Select("COALESCE(SUM(duration), 0)").Scan(&totalSeconds)
		database.Model(&model.StudySession{}).
			Where("user_id = ? AND date >= ? AND date <= ?", getUserID(r), startDate, endDate).
			Count(&totalSessions)
	} else {
		// 查询历史全部总计
		database.Model(&model.StudySession{}).Where("user_id = ?", getUserID(r)).
			Select("COALESCE(SUM(duration), 0)").Scan(&totalSeconds)
		database.Model(&model.StudySession{}).Where("user_id = ?", getUserID(r)).
			Count(&totalSessions)
	}

	// 每日统计
	var dailyStats []model.DailyStatItem
	var rows *sql.Rows
	var err error

	if useDateRange {
		// 使用指定日期范围查询
		rows, err = database.Model(&model.StudySession{}).
			Select("date, SUM(duration) as total_seconds, COUNT(*) as session_count").
			Where("user_id = ? AND date >= ? AND date <= ?", getUserID(r), startDate, endDate).
			Group("date").
			Order("date ASC"). // 日期范围查询时按升序排列，便于前端展示时间线
			Rows()
	} else {
		// 使用days参数查询最近N天
		days := parseIntQueryParam(r, "days", 30)
		rows, err = database.Model(&model.StudySession{}).
			Select("date, SUM(duration) as total_seconds, COUNT(*) as session_count").
			Where("user_id = ? AND date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)", getUserID(r), days).
			Group("date").
			Order("date DESC").
			Rows()
	}

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
	if err := database.Where("user_id = ? AND date = ?", getUserID(r), date).
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
//
//	{ "name": "Go语言开发", "color": "#FF6B6B" }
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
	req.Name = stripHTMLTags(req.Name)
	if len(req.Name) > 50 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "标签名不能超过50个字符"))
		return
	}

	if req.Color == "" {
		req.Color = "#6C5CE7"
	}

	tag := model.StudyTag{
		UserID: getUserID(r),
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
	if err := database.Where("user_id = ?", getUserID(r)).Order("created_at ASC").Find(&tags).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询标签失败"))
		return
	}

	// 转换为响应结构，显式带上小写 id 字段
	// （StudyTag 内嵌 gorm.Model，其 ID 默认序列化为大写 "ID"，前端无法用 t.id 读取）
	items := make([]model.StudyTagItem, 0, len(tags))
	for _, t := range tags {
		items = append(items, model.StudyTagItem{
			ID:    t.ID,
			Name:  t.Name,
			Color: t.Color,
		})
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(items))
}

// ============================================================
// DELETE /api/focus/tags/:id - 删除标签
// ============================================================

// DeleteTag 删除指定ID的标签
// 只有该标签的创建者才能删除
func DeleteTag(w http.ResponseWriter, r *http.Request) {
	// 从URL中提取标签ID
	// URL格式: /api/focus/tags/123
	path := strings.TrimPrefix(r.URL.Path, "/api/focus/tags/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 || parts[0] == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "标签ID不能为空"))
		return
	}

	tagID := parts[0]
	userID := getUserID(r)
	database := db.GetDB()

	// 查询标签是否存在且属于当前用户
	var tag model.StudyTag
	if err := database.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sendJSON(w, http.StatusNotFound, model.ErrorResponse(404, "标签不存在或无权限删除"))
			return
		}
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询标签失败"))
		return
	}

	// 删除标签（硬删除：彻底移除记录，释放 user_id+name 唯一索引，
	// 这样用户删除后可以重新创建同名标签，不会被软删除残留记录挡住）
	if err := database.Unscoped().Delete(&tag).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "删除标签失败"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(map[string]string{
		"message": "标签删除成功",
		"id":      tagID,
	}))
}

// ============================================================
// GET /api/focus/average - 获取用户平均每日专注时间
// ============================================================

// GetAverageFocus 获取用户的平均每日专注时间（分钟）
// 用于前端绘制平均线参考
func GetAverageFocus(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	userID := getUserID(r)

	// 查询用户所有专注记录的总秒数和总天数
	var totalSeconds int64
	var totalDays int64

	database.Model(&model.StudySession{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(duration), 0)").
		Scan(&totalSeconds)

	// 查询有多少天有专注记录（去重统计）
	database.Model(&model.StudySession{}).
		Where("user_id = ?", userID).
		Distinct("date").
		Count(&totalDays)

	// 计算平均值（分钟）
	var avgMinutes float64
	if totalDays > 0 {
		avgMinutes = float64(totalSeconds) / float64(totalDays) / 60.0
	} else {
		avgMinutes = 0
	}

	// 格式化显示
	var avgFormatted string
	if avgMinutes >= 60 {
		hours := int(avgMinutes / 60)
		mins := int(avgMinutes) % 60
		if mins > 0 {
			avgFormatted = fmt.Sprintf("%d小时%d分钟", hours, mins)
		} else {
			avgFormatted = fmt.Sprintf("%d小时", hours)
		}
	} else {
		avgFormatted = fmt.Sprintf("%d分钟", int(avgMinutes))
	}

	response := map[string]interface{}{
		"avg_minutes":   int(avgMinutes),
		"avg_formatted": avgFormatted,
		"total_days":    totalDays,
		"total_seconds": totalSeconds,
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(response))
}

// ============================================================
// GET /api/focus/earliest - 获取最早专注记录日期
// ============================================================

// GetEarliestFocusDate 获取当前用户最早的一条专注记录日期
// 用于前端确定可以查看的历史时间范围上限
func GetEarliestFocusDate(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	userID := getUserID(r)

	// 查询最早的专注记录日期
	var earliestDate string
	err := database.Model(&model.StudySession{}).
		Where("user_id = ?", userID).
		Select("MIN(date)").
		Row().
		Scan(&earliestDate)

	if err != nil {
		// 没有记录时返回空字符串，前端会处理为不限制
		sendJSON(w, http.StatusOK, model.SuccessResponse(map[string]string{
			"earliest_date": "",
		}))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(map[string]string{
		"earliest_date": earliestDate,
	}))
}

// ============================================================
// 专注碎片相关 API（碎片存储功能新增）
// ============================================================

// SaveFragment 保存专注碎片
// POST /api/focus/fragment
// 将中断的专注时间存入碎片银行
func SaveFragment(w http.ResponseWriter, r *http.Request) {
	var req model.SaveFragmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

	// 校验碎片时长（最少3分钟，最多4小时）
	if req.Duration < 180 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "碎片时长最少3分钟"))
		return
	}
	if req.Duration > 14400 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "碎片时长最多4小时"))
		return
	}

	// 设置默认标签
	if req.Tag == "" {
		req.Tag = "未分类"
	} else {
		req.Tag = stripHTMLTags(req.Tag)
	}
	if req.TagColor == "" {
		req.TagColor = "#6C5CE7"
	}

	now := time.Now()
	fragment := model.FocusFragment{
		UserID:   getUserID(r),
		Duration: req.Duration,
		Date:     now.Format("2006-01-02"),
		Tag:      req.Tag,
		TagColor: req.TagColor,
	}

	database := db.GetDB()
	if err := database.Create(&fragment).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "保存碎片失败"))
		return
	}

	sendJSON(w, http.StatusCreated, model.SuccessResponse(fragment))
}

// GetFragments 获取当日所有碎片
// GET /api/focus/fragments
func GetFragments(w http.ResponseWriter, r *http.Request) {
	database := db.GetDB()
	userID := getUserID(r)
	today := time.Now().Format("2006-01-02")

	var fragments []model.FocusFragment
	if err := database.Where("user_id = ? AND date = ?", userID, today).
		Order("created_at DESC").
		Find(&fragments).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询碎片失败"))
		return
	}

	// 转换为响应格式
	var items []model.FragmentItem
	var totalSeconds int64
	for _, f := range fragments {
		items = append(items, model.FragmentItem{
			ID:       f.ID,
			Duration: f.Duration,
			Minutes:  f.Duration / 60,
			Tag:      f.Tag,
			TagColor: f.TagColor,
			CreatedAt: f.CreatedAt,
		})
		totalSeconds += int64(f.Duration)
	}

	// 计算折扣后时间（按碎片大小分段折扣 + 数量惩罚）
	var discountedSeconds int64
	for _, f := range fragments {
		mins := float64(f.Duration) / 60
		var rate float64
		switch {
		case mins >= 40:
			rate = 1.0
		case mins >= 30:
			rate = 0.85
		case mins >= 20:
			rate = 0.70
		case mins >= 15:
			rate = 0.60
		case mins >= 10:
			rate = 0.50
		case mins >= 5:
			rate = 0.40
		default:
			rate = 0.30
		}
		discountedSeconds += int64(float64(f.Duration) * rate)
	}

	// 数量惩罚
	fragmentCount := len(fragments)
	var quantityPenalty float64 = 1.0
	switch {
	case fragmentCount > 5:
		quantityPenalty = 0.5
	case fragmentCount >= 4:
		quantityPenalty = 0.7
	case fragmentCount == 3:
		quantityPenalty = 0.9
	}

	discountedMins := int(float64(discountedSeconds) * quantityPenalty / 60)

	response := model.FragmentListResponse{
		Fragments:      items,
		TotalSeconds:   totalSeconds,
		TotalMinutes:   int(totalSeconds / 60),
		DiscountedMins: discountedMins,
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(response))
}

// CashoutFragments 兑现当日所有碎片
// POST /api/focus/fragments/cashout
// 将碎片按60%折扣兑换成专注记录
func CashoutFragments(w http.ResponseWriter, r *http.Request) {
	var req model.CashoutFragmentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

	database := db.GetDB()
	userID := getUserID(r)
	today := time.Now().Format("2006-01-02")

	// 查询当日所有碎片
	var fragments []model.FocusFragment
	if err := database.Where("user_id = ? AND date = ?", userID, today).
		Find(&fragments).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询碎片失败"))
		return
	}

	if len(fragments) == 0 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "没有可兑换的碎片"))
		return
	}

	// 计算碎片折扣（新算法：按碎片大小分段折扣 + 数量惩罚）
	var totalRawSeconds int64
	var discountedSeconds int64

	for _, f := range fragments {
		mins := float64(f.Duration) / 60
		totalRawSeconds += int64(f.Duration)

		// 按碎片大小计算折扣
		var rate float64
		switch {
		case mins >= 40:
			rate = 1.0 // ≥40分钟: 100%
		case mins >= 30:
			rate = 0.85 // 30-40分钟: 85%
		case mins >= 20:
			rate = 0.70 // 20-30分钟: 70%
		case mins >= 15:
			rate = 0.60 // 15-20分钟: 60%
		case mins >= 10:
			rate = 0.50 // 10-15分钟: 50%
		case mins >= 5:
			rate = 0.40 // 5-10分钟: 40%
		default:
			rate = 0.30 // 3-5分钟: 30%
		}
		discountedSeconds += int64(float64(f.Duration) * rate)
	}

	// 数量惩罚
	fragmentCount := len(fragments)
	var quantityPenalty float64 = 1.0
	switch {
	case fragmentCount > 5:
		quantityPenalty = 0.5 // >5个: 50%
	case fragmentCount >= 4:
		quantityPenalty = 0.7 // 4-5个: 70%
	case fragmentCount == 3:
		quantityPenalty = 0.9 // 3个: 90%
	}

	discountedSeconds = int64(float64(discountedSeconds) * quantityPenalty)

	// 设置标签
	tag := req.Tag
	if tag == "" {
		tag = "碎片专注 ⚡"
	} else {
		tag = stripHTMLTags(req.Tag)
	}
	tagColor := req.TagColor
	if tagColor == "" {
		tagColor = "#ff9f43" // 橙色区分
	}

	// 创建折扣后的专注记录
	session := model.StudySession{
		UserID:    userID,
		Duration:  int(discountedSeconds),
		Date:      today,
		StartedAt: time.Now(),
		Tag:       tag,
		TagColor:  tagColor,
	}

	// 使用事务确保原子性
	tx := database.Begin()

	// 创建专注记录
	if err := tx.Create(&session).Error; err != nil {
		tx.Rollback()
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "创建专注记录失败"))
		return
	}

	// 删除已兑换的碎片
	if err := tx.Where("user_id = ? AND date = ?", userID, today).
		Delete(&model.FocusFragment{}).Error; err != nil {
		tx.Rollback()
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "清空碎片失败"))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "交易提交失败"))
		return
	}

	// 计算实际折扣率
	var discountRate float64
	if totalRawSeconds > 0 {
		discountRate = float64(discountedSeconds) / float64(totalRawSeconds)
	}

	response := model.CashoutFragmentsResponse{
		RawMinutes:     totalRawSeconds / 60,
		ActualMinutes:  discountedSeconds / 60,
		Discount:       formatDiscount(discountRate),
		ClearedCount:   len(fragments),
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(response))
}

// ClearExpiredFragments 清理过期的碎片（可被定时任务调用）
// 删除非今日的碎片记录
func ClearExpiredFragments() {
	database := db.GetDB()
	today := time.Now().Format("2006-01-02")

	result := database.Where("date != ?", today).Delete(&model.FocusFragment{})
	if result.Error != nil {
		log.Printf("清理过期碎片失败: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("已清理 %d 条过期碎片", result.RowsAffected)
	}
}

// DeleteFragment 删除单个碎片
// DELETE /api/focus/fragments/:id
func DeleteFragment(w http.ResponseWriter, r *http.Request) {
	// 从URL中提取碎片ID
	// URL格式: /api/focus/fragments/123
	path := strings.TrimPrefix(r.URL.Path, "/api/focus/fragments/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 || parts[0] == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "碎片ID不能为空"))
		return
	}

	fragmentID := parts[0]
	userID := getUserID(r)
	today := time.Now().Format("2006-01-02")

	database := db.GetDB()

	// 查询碎片是否存在且属于当前用户且是今日碎片
	var fragment model.FocusFragment
	if err := database.Where("id = ? AND user_id = ? AND date = ?", fragmentID, userID, today).First(&fragment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sendJSON(w, http.StatusNotFound, model.ErrorResponse(404, "碎片不存在、无权限或已过期"))
			return
		}
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询碎片失败"))
		return
	}

	// 删除碎片
	if err := database.Delete(&fragment).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "删除碎片失败"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(map[string]string{
		"message": "碎片删除成功",
		"id":      fragmentID,
	}))
}
