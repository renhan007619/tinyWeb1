// Package model 定义应用程序中使用的所有数据结构
// =============================================
// 作用：
//   定义前后端交互的数据模型和数据库表映射，包括：
//   - Todo / TodoHistory: 备忘录待办任务和历史归档
//   - Setting: 用户设置（主题偏好等）
//   - Guestbook: 留言板留言
//   - VisitStats: 访问统计（Day 1 新增，使用 GORM 管理）
//   - APIResponse: 统一的 API 响应格式
//
// Day 1 更新日志（2026-04-07）：
//   - 新增 VisitStats 模型，用于记录访客的访问信息
//   - VisitStats 使用 GORM 标签进行数据库映射
//   - 包含核心字段（IP、访问次数）和扩展字段（设备类型、浏览器等）
//
// 使用方式：
//   handler 层使用这些结构体进行 JSON 序列化/反序列化，
//   db 层使用 GORM 的 AutoMigrate 自动根据结构体创建/更新数据表。
//
// GORM 标签说明：
//   - type       : 指定 MySQL 列类型
//   - uniqueIndex: 创建唯一索引
//   - not null   : 设置 NOT NULL 约束
//   - default    : 设置默认值
//   - size       : 指定字符串长度
// =============================================

package model

import (
	"time"

	"gorm.io/gorm" // GORM ORM 库，用于数据库操作
)

// ============================================================
// 访问统计相关模型（Day 1 新增）
// ============================================================

// VisitStats 访问统计结构体
// 对应数据库 visit_stats 表，记录每个访客的访问信息
//
// 设计思路：
//   - 以访客 IP 为主标识（uniqueIndex），同一个 IP 的多次访问会累加 visit_count
//   - 保留 FirstVisitAt 和 LastVisitAt（虽然 gorm.Model 已有 CreatedAt/UpdatedAt），
//     因为这两个字段的语义更明确，专门表示"首次访问"和"最后访问"
//   - 扩展字段（设备类型、浏览器、OS、来源）用于前端统计展示
//
// 数据库表结构（由 GORM AutoMigrate 自动创建）：
//   | 列名           | 类型          | 说明                    |
//   |----------------|---------------|------------------------|
//   | id             | bigint unsigned| 自增主键（gorm.Model自带）|
//   | created_at     | datetime(3)   | 记录创建时间             |
//   | updated_at     | datetime(3)   | 记录更新时间             |
//   | deleted_at     | datetime(3)   | 软删除时间（NULL=未删除）  |
//   | visitor_ip     | varchar(45)   | 访客IP（唯一索引）        |
//   | visit_count    | int           | 访问次数（默认1）         |
//   | first_visit_at | datetime(3)   | 首次访问时间             |
//   | last_visit_at  | datetime(3)   | 最后访问时间             |
//   | user_agent     | varchar(500)  | 原始 User-Agent（备用）   |
//   | device_type    | varchar(20)   | 设备类型                 |
//   | browser        | varchar(50)   | 浏览器名称               |
//   | os             | varchar(50)   | 操作系统                 |
//   | referrer       | varchar(500)  | 来源页面                 |
type VisitStats struct {
	// gorm.Model 是 GORM 内置的基础模型，自动包含以下字段：
	//   ID        uint           `gorm:"primarykey"`  // 自增主键
	//   CreatedAt time.Time      // 记录创建时间
	//   UpdatedAt time.Time      // 记录最后更新时间
	//   DeletedAt gorm.DeletedAt `gorm:"index"`       // 软删除标记（0值表示未删除）
	gorm.Model

	// ---- 核心字段 ----
	VisitorIP    string    `gorm:"type:varchar(45);uniqueIndex;not null" json:"visitor_ip"` // 访客 IP 地址（IPv6 最长45字符，设为唯一索引防重复）
	VisitCount   int       `gorm:"default:1;not null" json:"visit_count"`                   // 累计访问次数（每次访问+1）
	FirstVisitAt time.Time `gorm:"not null" json:"first_visit_at"`                          // 该访客首次访问的时间
	LastVisitAt  time.Time `gorm:"not null" json:"last_visit_at"`                           // 该访客最近一次访问的时间

	// ---- 扩展字段（用于统计分析） ----
	UserAgent  string `gorm:"type:varchar(500)" json:"user_agent"`  // 原始 User-Agent 字符串（保留原始数据，方便调试）
	DeviceType string `gorm:"type:varchar(20)" json:"device_type"`  // 设备类型：mobile（手机）/ desktop（电脑）/ tablet（平板）
	Browser    string `gorm:"type:varchar(50)" json:"browser"`      // 浏览器名称：Chrome / Safari / Firefox / Edge 等
	OS         string `gorm:"type:varchar(50)" json:"os"`           // 操作系统：Windows / macOS / Linux / Android / iOS
	Referrer   string `gorm:"type:varchar(500)" json:"referrer"`    // 访问来源页面 URL（如搜索引擎、直接访问等）
}

// TableName 指定 VisitStats 对应的数据库表名
// GORM 默认会将结构体名转为蛇形复数（visit_stats），
// 这里显式指定以确保表名一致
func (VisitStats) TableName() string {
	return "visit_stats"
}

// ============================================================
// 访问统计 API 相关模型（Day 2 新增）
// ============================================================

// VisitRecord 访问记录请求体
// 前端 POST /api/visit 时提交的 JSON 数据
type VisitRecord struct {
	VisitorIP  string `json:"visitor_ip"`   // 访客 IP 地址
	UserAgent  string `json:"user_agent"`   // 浏览器 User-Agent
	DeviceType string `json:"device_type"`  // 设备类型：mobile/desktop/tablet
	Browser    string `json:"browser"`      // 浏览器名称
	OS         string `json:"os"`           // 操作系统
	Referrer   string `json:"referrer"`     // 来源页面
}

// VisitStatsResponse 访问统计汇总响应
// 前端 GET /api/visit/stats 的返回数据
type VisitStatsResponse struct {
	TotalVisits    int64   `json:"total_visits"`    // 总访问次数（所有 IP 累加）
	UniqueVisitors int64   `json:"unique_visitors"` // 独立访客数（不同 IP 数量）
	LastVisitAt    *string `json:"last_visit_at"`   // 最后访问时间（可为空）
}

// ============================================================
// 备忘录相关模型
// ============================================================

// Todo 待办任务结构体
// 对应数据库 todos 表，存储用户当前的待办事项
type Todo struct {
	ID        int       `json:"id"`         // 任务唯一标识（自增主键）
	UserID    string    `json:"user_id"`    // 用户标识（当前固定为 "default"，预留多用户扩展）
	Category  string    `json:"category"`   // 分类："life"(生活) / "study"(学习) / "important"(重要)
	Text      string    `json:"text"`       // 任务内容文本（最长200字符）
	Done      bool      `json:"done"`       // 是否已完成：true=完成, false=未完成
	SortOrder int       `json:"sort_order"` // 排序序号（数值越小越靠前）
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 最后更新时间
}

// TodoCreateRequest 新增任务的请求体结构
// 前端 POST /api/todos 时提交的 JSON 数据
type TodoCreateRequest struct {
	Category string `json:"category"` // 必填，分类：life/study/important
	Text     string `json:"text"`     // 必填，任务内容
}

// TodoUpdateRequest 更新任务的请求体结构
// 前端 PUT /api/todos/:id 时提交的 JSON 数据
// 字段均为可选，只更新提供的字段
type TodoUpdateRequest struct {
	Text *string `json:"text,omitempty"` // 可选，更新的任务内容
	Done *bool   `json:"done,omitempty"` // 可选，更新的完成状态
}

// TodoHistory 历史归档结构体
// 对应数据库 todo_history 表，存储已归档的过期任务
type TodoHistory struct {
	ID          int    `json:"id"`           // 记录唯一标识（自增主键）
	UserID      string `json:"user_id"`      // 用户标识
	ArchiveDate string `json:"archive_date"` // 归档日期（格式 YYYY-MM-DD）
	Category    string `json:"category"`     // 归档时的分类
	Text        string `json:"text"`         // 任务内容
	Done        bool   `json:"done"`         // 归档时的完成状态
}

// TodoHistoryByDate 按日期分组的历史归档响应结构
// 前端 GET /api/todo/history?date=2026-04-05 的返回数据
type TodoHistoryByDate struct {
	Date  string                `json:"date"`  // 归档日期
	Todos map[string][]TodoItem `json:"todos"` // 按 category 分组的任务列表
}

// TodoItem 简化的待办项（用于历史归档展示）
type TodoItem struct {
	Text string `json:"text"` // 任务内容
	Done bool   `json:"done"` // 完成状态
}

// ============================================================
// 设置相关模型
// ============================================================

// Setting 用户设置结构体
// 对应数据库 settings 表，存储用户的个性化偏好设置
type Setting struct {
	UserID    string    `json:"user_id"`    // 用户标识（主键）
	Theme     string    `json:"theme"`      // 主题偏好："light"(亮色) / "dark"(暗色)
	UpdatedAt time.Time `json:"updated_at"` // 最后更新时间
}

// ThemeUpdateRequest 主题更新的请求体结构
// 前端 PUT /api/settings/theme 时提交的 JSON 数据
type ThemeUpdateRequest struct {
	Theme string `json:"theme"` // 必填，目标主题："light" 或 "dark"
}

// ============================================================
// 留言板相关模型
// ============================================================

// Guestbook 留言板留言结构体
// 对应数据库 guestbook 表，存储访客的留言
type Guestbook struct {
	ID        int       `json:"id"`        // 留言唯一标识（自增主键）
	Nickname  string    `json:"nickname"`  // 留言者昵称（可选，为空时显示"匿名访客"）
	Content   string    `json:"content"`   // 留言内容（最长500字符）
	CreatedAt time.Time `json:"created_at"` // 发布时间
}

// GuestbookCreateRequest 发布留言的请求体结构
// 前端 POST /api/guestbook 时提交的 JSON 数据
type GuestbookCreateRequest struct {
	Nickname string `json:"nickname"` // 可选，留言者昵称
	Content  string `json:"content"`  // 必填，留言内容
}

// GuestbookListResponse 留言列表的分页响应结构
// 前端 GET /api/guestbook?page=1&size=20 的返回数据
type GuestbookListResponse struct {
	List       []Guestbook `json:"list"`        // 当前页的留言列表
	Total      int64       `json:"total"`       // 留言总数
	Page       int         `json:"page"`        // 当前页码
	Size       int         `json:"size"`        // 每页条数
	TotalPages int         `json:"total_pages"` // 总页数
}

// ============================================================
// API 统一响应模型
// ============================================================

// APIResponse 统一的 API 响应格式
// 所有 API 接口都使用此结构返回数据，便于前端统一处理
// 成功时 code=0，失败时 code>0 并附带错误信息
type APIResponse struct {
	Code    int         `json:"code"`            // 状态码：0=成功, 其他=错误码
	Message string      `json:"message"`         // 响应消息：成功时为 "success"，失败时为错误描述
	Data    interface{} `json:"data,omitempty"`  // 响应数据（可选，查询接口有值）
}

// SuccessResponse 快速创建成功响应的辅助函数
// code=0, message="success", data 为传入的数据
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponse 快速创建错误响应的辅助函数
// code>0, message 为错误描述, data 为 nil
func ErrorResponse(code int, message string) APIResponse {
	return APIResponse{
		Code:    code,
		Message: message,
	}
}
