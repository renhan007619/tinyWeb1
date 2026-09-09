// Package handler 每日完成度自评（Daily Review）相关 HTTP API handlers
// =============================================
// 作用：
//   让用户按天对"任务完成度"自评（低/中/高），与专注时长解耦，
//   用于识别"用时少但任务完成多"的高效日：时长短 + 完成度高 = 效率高。
//
// 存储设计（按用户要求，不加新表）：
//   完成度直接记录在 study_sessions.completion_level 列（low/medium/high，NULL=未评）。
//   study_sessions 是"会话"级（一天可能有多条记录），而完成度是"天"级属性，因此：
//   - 写入：评价某天时把该用户当天【所有】会话行同步更新为同一评级
//   - 读取：按天聚合时取非空评级 MAX(completion_level)（NULL 会被忽略，
//     所以当天评价后新插入的会话行不会破坏该天的评级）
//
// 评价窗口规则（北京时间，与服务器系统时区无关）：
//   - 当天 18:00（含）之后：可评价/修改"今天"
//   - 次日全天：可评价/修改"昨天"（补评 + 修正头天晚上的判断）
//   - D+2 起：该日期锁定，UI 不再提供入口，仅能直接改数据库
//
// API 接口列表：
//   GET  /api/focus/review/editable → 评价弹窗初始状态（今天/昨天的日期、评级、是否可评）
//   POST /api/focus/review          → 提交/修改某天评级（服务端权威校验评价窗口）
//   GET  /api/focus/reviews         → 指定日期范围内的全部评级（柱状图柱顶标注用）
//
// 数据流向：
//   前端 fetch() → Go HTTP handler → GORM 查询 → MySQL(study_sessions) → JSON 响应
// =============================================

package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tinyweb1/db"
	"tinyweb1/model"
)

// 完成度三档取值（与前端展示 低/中/高 对应）
const (
	reviewLevelLow    = "low"
	reviewLevelMedium = "medium"
	reviewLevelHigh   = "high"

	// reviewOpenHour 当天可评价的起始小时（北京时间）
	// 18:00 之前当天还在进行中，无法评判"今天整天的任务完成度"
	reviewOpenHour = 18
)

// cstLocation 北京时间时区
// 优先使用系统时区库加载 Asia/Shanghai（含夏令时历史的正确规则），
// 失败则退化为固定 UTC+8，保证 18:00 判定与日期归属不依赖服务器系统时区
var cstLocation = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		return loc
	}
	return time.FixedZone("CST", 8*3600)
}()

// beijingNow 返回当前北京时间
// 评价窗口、日期归属统一以它为准（历史 study_sessions 的日期不迁移，保持原样）
func beijingNow() time.Time {
	return time.Now().In(cstLocation)
}

// validReviewLevel 校验评级取值是否合法
func validReviewLevel(level string) bool {
	return level == reviewLevelLow || level == reviewLevelMedium || level == reviewLevelHigh
}

// normalizeDate 将日期字符串统一截取为 YYYY-MM-DD 前 10 位
// （兼容不同驱动/序列化下 DATE 可能带时间部分的情况）
func normalizeDate(d string) string {
	if len(d) >= 10 {
		return d[:10]
	}
	return d
}

// reviewWindow 检查某天（YYYY-MM-DD）当前是否处于可评价窗口
// 返回 (是否可评, 不可评原因)。字符串日期可直接字典序比较（格式固定）。
func reviewWindow(date string) (bool, string) {
	now := beijingNow()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	switch {
	case date == yesterday:
		// 昨天全天可补评/修改
		return true, ""
	case date == today && now.Hour() >= reviewOpenHour:
		// 今天 18:00 后开始可评
		return true, ""
	case date == today:
		return false, fmt.Sprintf("每天 %d:00 后才能评价当天，次日全天仍可补评或修改", reviewOpenHour)
	case date > today:
		return false, "不能评价未来的日期"
	default:
		return false, "该日期已过评价期限（仅支持评价今天与昨天）"
	}
}

// loadDayReviewLevel 查询某一天当前的完成度评级
// 返回该天所有会话行中非空的评级（MAX 忽略 NULL）；当天无记录或全未评时返回空串
func loadDayReviewLevel(userID uint, date string) (string, error) {
	var level sql.NullString
	err := db.GetDB().Raw(
		"SELECT MAX(completion_level) FROM study_sessions WHERE user_id = ? AND date = ? AND deleted_at IS NULL",
		userID, date,
	).Scan(&level).Error
	if err != nil {
		return "", err
	}
	return level.String, nil // NULL → ""
}

// ============================================================
// GET /api/focus/review/editable - 评价弹窗初始状态
// ============================================================

// GetReviewEditable 返回评价弹窗需要的初始状态：
// 今天/昨天的日期字符串、当前评级（空=未评）、是否可评
// 前端据此渲染：可评的日期可点选三档，18:00 前的"今天"置灰并给出提示
func GetReviewEditable(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	now := beijingNow()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	todayLevel, err := loadDayReviewLevel(userID, today)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询完成度评价失败"))
		return
	}
	yesterdayLevel, err := loadDayReviewLevel(userID, yesterday)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询完成度评价失败"))
		return
	}

	todayCan, todayHint := reviewWindow(today)

	response := model.ReviewEditableResponse{
		ServerTime: now.Format("2006-01-02 15:04"),
		Today: model.ReviewItem{
			Date:    today,
			Level:   todayLevel,
			CanEdit: todayCan,
			Hint:    todayHint,
		},
		Yesterday: model.ReviewItem{
			Date:    yesterday,
			Level:   yesterdayLevel,
			CanEdit: true, // 昨天全天可评
		},
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(response))
}

// ============================================================
// POST /api/focus/review - 提交/修改某天评级
// ============================================================

// SaveReview 提交或修改某天的完成度评级
// 把该用户当天【所有】专注记录行的 completion_level 同步更新为目标评级
// Request Body (JSON)：{ "date": "2026-09-10", "level": "high" }
//
// 校验规则（服务端权威判定，不信任前端）：
//   - date 必须为 YYYY-MM-DD
//   - level 必须为 low / medium / high
//   - date 必须处于可评价窗口：今天(≥18:00 北京) 或 昨天，其余一律 403 拒绝
//   - 当天必须有专注记录（否则没有行可写评级，返回 400）
func SaveReview(w http.ResponseWriter, r *http.Request) {
	var req model.SaveReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求数据格式错误"))
		return
	}

	req.Date = normalizeDate(req.Date)
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "日期格式错误，请使用YYYY-MM-DD格式"))
		return
	}
	if !validReviewLevel(req.Level) {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "评级必须为 low/medium/high"))
		return
	}

	// 评价窗口校验（权威判定）
	if ok, hint := reviewWindow(req.Date); !ok {
		sendJSON(w, http.StatusForbidden, model.ErrorResponse(403, hint))
		return
	}

	database := db.GetDB()
	userID := getUserID(r)

	// 当天所有专注记录行同步更新为同一评级
	// （day-level 语义：完成度属于"这一天"，不属于某一条会话）
	result := database.Model(&model.StudySession{}).
		Where("user_id = ? AND date = ?", userID, req.Date).
		Update("completion_level", req.Level)
	if result.Error != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "保存评价失败"))
		return
	}
	if result.RowsAffected == 0 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "该日期没有专注记录，无法评价完成度"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(map[string]string{
		"date":  req.Date,
		"level": req.Level,
	}))
}

// ============================================================
// GET /api/focus/reviews - 指定日期范围内的评级列表
// ============================================================

// GetReviewsRange 返回指定日期范围内"已评价日期"的评级（升序）
// Query 参数：
//   - start: 开始日期（YYYY-MM-DD）
//   - end: 结束日期（YYYY-MM-DD）
//
// 按天聚合取非空评级：SELECT date, MAX(completion_level) ... GROUP BY date
// （同一天多条会话行共享同一评级，MAX 只用于剔除 NULL）
// 前端拿到该周评级后，在柱状图的每根柱子上标注 未/低/中/高
func GetReviewsRange(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")
	if startDate == "" || endDate == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "缺少 start/end 日期参数"))
		return
	}
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "开始日期格式错误，请使用YYYY-MM-DD格式"))
		return
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "结束日期格式错误，请使用YYYY-MM-DD格式"))
		return
	}

	// 按天聚合：同天全部行评级一致，MAX(completion_level) 只为剔除 NULL
	rows, err := db.GetDB().Model(&model.StudySession{}).
		Select("date, MAX(completion_level) AS level").
		Where("user_id = ? AND date >= ? AND date <= ? AND completion_level IS NOT NULL",
			getUserID(r), startDate, endDate).
		Group("date").
		Order("date ASC").
		Rows()
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "查询完成度评价失败"))
		return
	}
	defer rows.Close()

	items := make([]model.ReviewItem, 0)
	for rows.Next() {
		var item model.ReviewItem
		var level sql.NullString
		if err := rows.Scan(&item.Date, &level); err != nil {
			continue
		}
		item.Level = level.String
		items = append(items, item)
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(model.ReviewRangeResponse{Items: items}))
}
