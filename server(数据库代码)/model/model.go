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
// 用户认证相关模型（注册登录功能新增）
// ============================================================

// User 用户结构体
// 对应数据库 users 表，存储用户账号信息
//
// 设计思路：
//   - Username 设为唯一索引，不允许重复
//   - PasswordHash 存储 bcrypt 加密后的密码哈希，永远不存明文
//   - Role 区分角色：admin（管理员）/ user（普通用户），默认 user
//   - json:"-" 标签表示 PasswordHash 不会序列化到 JSON 返回前端（安全）
//
// 数据库表结构（由 GORM AutoMigrate 自动创建）：
//
//	| 列名          | 类型          | 说明                    |
//	|---------------|---------------|------------------------|
//	| id            | bigint unsigned| 自增主键                |
//	| created_at    | datetime(3)   | 创建时间                |
//	| updated_at    | datetime(3)   | 更新时间                |
//	| deleted_at    | datetime(3)   | 软删除时间              |
//	| username      | varchar(50)   | 用户名（唯一索引）       |
//	| password_hash | varchar(255)  | bcrypt密码哈希           |
//	| role          | varchar(20)   | 角色：admin/user         |
type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"` // 用户名（唯一）
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`                   // 密码哈希（不返回前端）
	Role         string `gorm:"type:varchar(20);default:user;not null" json:"role"`    // 角色：admin / user
}

// TableName 指定 User 对应的数据库表名
func (User) TableName() string {
	return "users"
}

// RegisterRequest 注册请求体
// 前端 POST /api/auth/register 时提交的 JSON 数据，接受前端发来的注册请求数据
type RegisterRequest struct {
	Username string `json:"username" binding:"required"` // 用户名（必填，3-50字符）
	Password string `json:"password" binding:"required"` // 密码（必填，6位以上）
}

// LoginRequest 登录请求体
// 前端 POST /api/auth/login 时提交的 JSON 数据
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// LoginResponse 登录成功响应数据（双Token版本）
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`  // Access Token（2小时有效）
	RefreshToken string   `json:"refresh_token"` // Refresh Token（7天有效）
	ExpiresIn    int64    `json:"expires_in"`    // Access Token有效期（秒）
	User         UserInfo `json:"user"`          // 用户信息
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // Refresh Token
}

// TokenPairResponse 双Token响应
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`  // 新的Access Token
	RefreshToken string `json:"refresh_token"` // 新的Refresh Token
	ExpiresIn    int64  `json:"expires_in"`    // Access Token有效期（秒）
}

// UserInfo 用户公开信息（不含密码）
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

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
//
//	| 列名           | 类型          | 说明                    |
//	|----------------|---------------|------------------------|
//	| id             | bigint unsigned| 自增主键（gorm.Model自带）|
//	| created_at     | datetime(3)   | 记录创建时间             |
//	| updated_at     | datetime(3)   | 记录更新时间             |
//	| deleted_at     | datetime(3)   | 软删除时间（NULL=未删除）  |
//	| visitor_ip     | varchar(45)   | 访客IP（唯一索引）        |
//	| visit_count    | int           | 访问次数（默认1）         |
//	| first_visit_at | datetime(3)   | 首次访问时间             |
//	| last_visit_at  | datetime(3)   | 最后访问时间             |
//	| user_agent     | varchar(500)  | 原始 User-Agent（备用）   |
//	| device_type    | varchar(20)   | 设备类型                 |
//	| browser        | varchar(50)   | 浏览器名称               |
//	| os             | varchar(50)   | 操作系统                 |
//	| referrer       | varchar(500)  | 来源页面                 |
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
	UserAgent  string `gorm:"type:varchar(500)" json:"user_agent"` // 原始 User-Agent 字符串（保留原始数据，方便调试）
	DeviceType string `gorm:"type:varchar(20)" json:"device_type"` // 设备类型：mobile（手机）/ desktop（电脑）/ tablet（平板）
	Browser    string `gorm:"type:varchar(50)" json:"browser"`     // 浏览器名称：Chrome / Safari / Firefox / Edge 等
	OS         string `gorm:"type:varchar(50)" json:"os"`          // 操作系统：Windows / macOS / Linux / Android / iOS
	Referrer   string `gorm:"type:varchar(500)" json:"referrer"`   // 访问来源页面 URL（如搜索引擎、直接访问等）
}

// TableName 指定 VisitStats 对应的数据库表名
// GORM 默认会将结	构体名转为蛇形复数（visit_stats），
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
	VisitorIP  string `json:"visitor_ip"`  // 访客 IP 地址
	UserAgent  string `json:"user_agent"`  // 浏览器 User-Agent
	DeviceType string `json:"device_type"` // 设备类型：mobile/desktop/tablet
	Browser    string `json:"browser"`     // 浏览器名称
	OS         string `json:"os"`          // 操作系统
	Referrer   string `json:"referrer"`    // 来源页面
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
	UserID    uint      `json:"user_id"`    // 关联用户ID（登录后从JWT token获取）
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
	UserID      uint   `json:"user_id"`      // 关联用户ID
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
	ID        int       `json:"id"`         // 留言唯一标识（自增主键）
	Nickname  string    `json:"nickname"`   // 留言者昵称（可选，为空时显示"匿名访客"）
	Content   string    `json:"content"`    // 留言内容（最长500字符）
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
// HTML 页面管理相关模型（管理员共享文件库功能）
// ============================================================

// Page HTML 页面元数据表
// 对应数据库 pages 表，存储管理员上传的 HTML 文件信息
//
// 设计思路：
//   - 文件实际存储在磁盘 uploads/pages/ 目录，数据库存元数据
//   - Slug 用于 URL 访问（如 /pages/my-page），只允许小写字母、数字、横线
//   - UploadBy 冗余存储用户名，方便前端显示而不需要 JOIN 查询
//   - 多个管理员可以上传/查看/删除所有页面（共享文件库）
//
// 数据库表结构：
//
//	| 列名       | 类型          | 说明                        |
//	|------------|---------------|----------------------------|
//	| id         | bigint unsigned| 自增主键                    |
//	| created_at | datetime(3)   | 创建时间                    |
//	| updated_at | datetime(3)   | 更新时间                    |
//	| deleted_at | datetime(3)   | 软删除时间                  |
//	| title      | varchar(100)  | 页面标题                    |
//	| slug       | varchar(50)   | URL标识（唯一索引）          |
//	| file_name  | varchar(100)  | 磁盘文件名                  |
//	| size       | bigint        | 文件大小(bytes)             |
//	| upload_by  | varchar(50)   | 上传者用户名                |
type Page struct {
	gorm.Model
	Title    string `gorm:"type:varchar(100);not null" json:"title"`           // 页面显示标题
	Slug     string `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug"` // URL标识（如 "my-page"）
	FileName string `gorm:"type:varchar(100);not null" json:"file_name"`       // 磁盘存储文件名
	Size     int64  `gorm:"not null" json:"size"`                              // 文件大小（字节）
	UploadBy string `gorm:"type:varchar(50);not null" json:"upload_by"`        // 上传者用户名
}

// TableName 指定 Page 对应的数据库表名
func (Page) TableName() string {
	return "pages"
}

// PageResponse 页面列表响应格式
// 前端 GET /api/admin/pages 的返回数据项
type PageResponse struct {
	ID        uint      `json:"id"`         // 页面ID
	Title     string    `json:"title"`      // 页面标题
	Slug      string    `json:"slug"`       // URL标识
	FileName  string    `json:"file_name"`  // 文件名
	Size      int64     `json:"size"`       // 文件大小（字节）
	SizeHuman string    `json:"size_human"` // 格式化的大小（如 "1.5 KB"）
	UploadBy  string    `json:"upload_by"`  // 上传者
	CreatedAt time.Time `json:"created_at"` // 上传时间
}

// ============================================================
// 专注时间相关模型（Focus Time 功能新增）
// ============================================================

// StudySession 一次专注/学习时段记录
// 对应数据库 study_sessions 表，记录用户每次专注的详细信息
//
// 设计思路：
//   - 以 user_id + date 为核心查询维度（按用户、按日期聚合统计）
//   - Duration 存储秒数（前端转换为"X小时Y分钟"展示）
//   - Tag 标识本次专注做了什么（如"Go语言开发"、"算法练习"）
//   - TagColor 前端用于区分不同标签的显示颜色
//   - Date 单独存日期字符串，避免每次查询都要从 started_at 提取日期
//
// 数据库表结构（由 GORM AutoMigrate 自动创建）：
//
//	| 列名       | 类型          | 说明                        |
//	|------------|---------------|----------------------------|
//	| id         | bigint unsigned| 自增主键                    |
//	| created_at | datetime(3)   | 创建时间                    |
//	| updated_at | datetime(3)   | 更新时间                    |
//	| deleted_at | datetime(3)   | 软删除时间                  |
//	| user_id    | bigint unsigned| 关联用户ID（索引）           |
//	| duration   | int           | 专注秒数（如1500=25分钟）    |
//	| date       | date          | 日期 YYYY-MM-DD（索引）      |
//	| started_at | datetime(3)   | 开始专注的时间               |
//	| tag        | varchar(50)   | 标签名（如"Go语言开发"）      |
//	| tag_color  | varchar(7)    | 标签颜色（如"#FF6B6B"）      |
//	| completion_level | varchar(10)| 当天完成度 low/medium/high（按天同步，NULL=未评）|
type StudySession struct {
	gorm.Model
	UserID           uint      `gorm:"index;not null" json:"user_id"`                      // 关联用户ID
	Duration         int       `gorm:"not null" json:"duration"`                           // 本次专注秒数
	Date             string    `gorm:"type:date;index;not null" json:"date"`               // 日期 YYYY-MM-DD
	StartedAt        time.Time `gorm:"not null" json:"started_at"`                         // 开始时间
	Tag              string    `gorm:"type:varchar(50);index;not null" json:"tag"`         // 标签名
	TagColor         string    `gorm:"type:varchar(7);default:'#6C5CE7'" json:"tag_color"` // 标签颜色
	CompletionLevel  string    `gorm:"type:varchar(10);index;default:NULL" json:"completion_level,omitempty"` // 当日完成度（任务完成度评价功能新增；NULL=未评，default:NULL 保证新记录写入 NULL 而非空串）
}

// TableName 指定 StudySession 对应的数据库表名
func (StudySession) TableName() string {
	return "study_sessions"
}

// ============================================================
// 专注碎片相关模型（Focus Fragment 功能新增）
// ============================================================

// FocusFragment 专注碎片临时存储结构体
// 对应数据库 focus_fragments 表，存储用户中断的专注时间碎片
//
// 设计思路：
//   - 用户暂停专注时可以选择将已专注时间存入碎片银行（原额存入，不折扣）
//   - 碎片次日 0 点自动过期清零，制造紧迫感
//   - 用户可以随时将碎片原额提现成专注记录（存多少提多少，无惩罚机制）
//   - 小于 3 分钟的专注不能被保存为碎片
//
// 数据库表结构（由 GORM AutoMigrate 自动创建）：
//
//	| 列名       | 类型          | 说明                        |
//	|------------|---------------|----------------------------|
//	| id         | bigint unsigned| 自增主键                    |
//	| created_at | datetime(3)   | 创建时间                    |
//	| updated_at | datetime(3)   | 更新时间                    |
//	| deleted_at | datetime(3)   | 软删除时间                  |
//	| user_id    | bigint unsigned| 关联用户ID                   |
//	| duration   | int           | 碎片秒数（实际专注时间）      |
//	| date       | date          | 所属日期 YYYY-MM-DD          |
//	| tag        | varchar(50)   | 标签名                       |
//	| tag_color  | varchar(7)    | 标签颜色                     |
type FocusFragment struct {
	gorm.Model
	UserID   uint   `gorm:"index;not null" json:"user_id"`                      // 关联用户ID
	Duration int    `gorm:"not null" json:"duration"`                           // 碎片秒数
	Date     string `gorm:"type:date;index;not null" json:"date"`               // 所属日期
	Tag      string `gorm:"type:varchar(50);not null" json:"tag"`               // 标签名
	TagColor string `gorm:"type:varchar(7);default:'#6C5CE7'" json:"tag_color"` // 标签颜色
}

// TableName 指定 FocusFragment 对应的数据库表名
func (FocusFragment) TableName() string {
	return "focus_fragments"
}

// ---- 每日完成度 API 请求/响应结构体（任务完成度评价功能新增）----
// 说明：完成度以"天"为单位记录在 study_sessions.completion_level 列上
// （评价时按天同步更新当天所有会话行，聚合时按天取非空评级），不再单独建表。

// SaveReviewRequest 提交/修改完成度评价的请求体
// 前端 POST /api/focus/review 时提交的 JSON 数据
type SaveReviewRequest struct {
	Date  string `json:"date" binding:"required"`  // 被评价的日期 YYYY-MM-DD
	Level string `json:"level" binding:"required"` // low / medium / high
}

// ReviewItem 单个日期的可评价状态
type ReviewItem struct {
	Date    string `json:"date"`     // 日期 YYYY-MM-DD
	Level   string `json:"level"`    // 当前评级：low/medium/high，空串=未评
	CanEdit bool   `json:"can_edit"` // 当前是否允许评价/修改
	Hint    string `json:"hint"`     // 不可评价时的原因提示（可为空）
}

// ReviewEditableResponse 评价弹窗初始数据
// GET /api/focus/review/editable 返回的数据
type ReviewEditableResponse struct {
	ServerTime string     `json:"server_time"` // 服务器北京时间（用于前端提示）
	Today      ReviewItem `json:"today"`       // 今天（18:00 后可评）
	Yesterday  ReviewItem `json:"yesterday"`   // 昨天（全天可评）
}

// ReviewRangeResponse 指定日期范围内的完成度评级列表
// GET /api/focus/reviews?start=..&end=.. 返回的数据（柱状图柱顶标注用）
// 只包含已评价的日期，未评价的日期由前端根据"是否有专注记录"决定是否标注"未"
type ReviewRangeResponse struct {
	Items []ReviewItem `json:"items"`
}

// ---- 专注碎片 API 请求/响应结构体 ----

// SaveFragmentRequest 保存专注碎片的请求体
// 前端 POST /api/focus/fragment 时提交的 JSON 数据
type SaveFragmentRequest struct {
	Duration int    `json:"duration" binding:"required,min=180,max=14400"` // 碎片秒数（最少3分钟，最多4小时）
	Tag      string `json:"tag"`                                           // 标签名
	TagColor string `json:"tag_color"`                                     // 标签颜色
}

// FragmentItem 单个碎片信息
type FragmentItem struct {
	ID       uint      `json:"id"`         // 碎片ID
	Duration int       `json:"duration"`   // 碎片秒数
	Minutes  int       `json:"minutes"`    // 碎片分钟数（方便前端展示）
	Tag      string    `json:"tag"`        // 标签名
	TagColor string    `json:"tag_color"`  // 标签颜色
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// FragmentListResponse 碎片列表响应
type FragmentListResponse struct {
	Fragments    []FragmentItem `json:"fragments"`       // 碎片列表
	TotalSeconds int64          `json:"total_seconds"`   // 总秒数
	TotalMinutes int            `json:"total_minutes"`   // 总分钟数
	CashoutMins  int            `json:"cashout_mins"`    // 可提现分钟数（原额提现，等于总分钟数）
}

// CashoutFragmentsResponse 提现碎片响应
type CashoutFragmentsResponse struct {
	TotalMinutes int64 `json:"total_minutes"`  // 提现总分钟数（原额计入专注时间）
	SessionCount int   `json:"session_count"`  // 生成的专注记录条数（每条碎片一条，保留各自标签）
	ClearedCount int   `json:"cleared_count"`  // 清空的碎片数量
}

// StudyTag 用户自定义的专注标签模板
// 对应数据库 study_tags 表，存储用户创建/使用的标签
//
// 设计思路：
//   - uniqueIndex:user_tag 确保同一个用户不会有同名标签
//   - Color 用于前端显示标签时的颜色标识
//
// 数据库表结构：
//
//	| 列名       | 类型          | 说明                        |
//	|------------|---------------|----------------------------|
//	| id         | bigint unsigned| 自增主键                    |
//	| created_at | datetime(3)   | 创建时间                    |
//	| updated_at | datetime(3)   | 更新时间                    |
//	| deleted_at | datetime(3)   | 软删除时间                  |
//	| user_id    | bigint unsigned| 关联用户ID                   |
//	| name       | varchar(50)   | 标签名（唯一索引）           |
//	| color      | varchar(7)    | 标签颜色                    |
type StudyTag struct {
	gorm.Model
	UserID uint   `gorm:"uniqueIndex:user_tag;not null" json:"user_id"`      // 关联用户ID
	Name   string `gorm:"uniqueIndex:user_tag;size:50;not null" json:"name"` // 标签名
	Color  string `gorm:"type:varchar(7);default:'#6C5CE7'" json:"color"`    // 显示颜色
}

// TableName 指定 StudyTag 对应的数据库表名
func (StudyTag) TableName() string {
	return "study_tags"
}

// StudyTagItem 标签列表响应项
// GET /api/focus/tags 返回的数据，显式带小写 id 字段。
// StudyTag 内嵌 gorm.Model，其 ID 字段默认序列化为大写 "ID"，前端无法用 t.id 读取，
// 因此对外返回时用本结构体转换。
type StudyTagItem struct {
	ID    uint   `json:"id"`    // 标签ID
	Name  string `json:"name"`  // 标签名
	Color string `json:"color"` // 标签颜色
}

// ---- 专注时间 API 请求/响应结构体 ----

// CreateFocusSessionRequest 创建专注记录的请求体
// 前端 POST /api/focus/session 时提交的 JSON 数据
type CreateFocusSessionRequest struct {
	Duration int    `json:"duration" binding:"required,min=60,max=14400"` // 专注秒数（最少1分钟，最多4小时）
	Tag      string `json:"tag"`                                          // 标签名
	TagColor string `json:"tag_color"`                                    // 标签颜色
}

// CreateTagRequest 创建标签的请求体
// 前端 POST /api/focus/tags 时提交的 JSON 数据
type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"` // 标签名（必填）
	Color string `json:"color"`                   // 标签颜色（可选，默认#6C5CE7）
}

// TodayFocusResponse 今日学习统计响应
// GET /api/focus/today 返回的数据
type TodayFocusResponse struct {
	TotalSeconds   int64        `json:"total_seconds"`   // 今日总专注秒数
	TotalFormatted string       `json:"total_formatted"` // 格式化的总时长（如"3小时25分钟"）
	SessionCount   int64        `json:"session_count"`   // 今日专注次数
	ByTag          []TagSummary `json:"by_tag"`          // 按标签分组的统计
}

// TagSummary 单个标签的汇总信息
type TagSummary struct {
	Tag        string  `json:"tag"`        // 标签名
	Seconds    int64   `json:"seconds"`    // 该标签总秒数
	Percentage float64 `json:"percentage"` // 占比（百分比，如 39.0 表示 39%）
	Color      string  `json:"color"`      // 标签颜色
}

// FocusSummaryResponse 历史总览响应
// GET /api/focus/summary 返回的数据
type FocusSummaryResponse struct {
	TotalSeconds   int64           `json:"total_seconds"`   // 历史总专注秒数
	TotalFormatted string          `json:"total_formatted"` // 格式化的总时长
	TotalSessions  int64           `json:"total_sessions"`  // 总专注次数
	DailyStats     []DailyStatItem `json:"daily_stats"`     // 每日统计列表
}

// DailyStatItem 单日统计项
type DailyStatItem struct {
	Date           string `json:"date"`            // 日期 YYYY-MM-DD
	TotalSeconds   int64  `json:"total_seconds"`   // 当日总秒数
	TotalFormatted string `json:"total_formatted"` // 格式化时长
	SessionCount   int64  `json:"session_count"`   // 当日专注次数
}

// FocusHistoryResponse 某日详细记录响应
// GET /api/focus/history?date=2026-04-15 返回的数据
type FocusHistoryResponse struct {
	Date     string               `json:"date"`     // 日期
	Sessions []StudySessionDetail `json:"sessions"` // 该日所有专注记录
}

// StudySessionDetail 单次专注记录的详细信息
type StudySessionDetail struct {
	ID        uint   `json:"id"`         // 记录ID
	Duration  int    `json:"duration"`   // 专注秒数
	Tag       string `json:"tag"`        // 标签名
	TagColor  string `json:"tag_color"`  // 标签颜色
	StartedAt string `json:"started_at"` // 开始时间（格式化后的字符串）
}

// ============================================================
// API 统一响应模型
// ============================================================

// APIResponse 统一的 API 响应格式
// 所有 API 接口都使用此结构返回数据，便于前端统一处理
// 成功时 code=0，失败时 code>0 并附带错误信息
type APIResponse struct {
	Code    int         `json:"code"`           // 状态码：0=成功, 其他=错误码
	Message string      `json:"message"`        // 响应消息：成功时为 "success"，失败时为错误描述
	Data    interface{} `json:"data,omitempty"` // 响应数据（可选，查询接口有值）
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

// ============================================================
// 图库相关模型（Gallery 功能新增 - Day 1）
// ============================================================

// GalleryImage 图片结构体
// 对应数据库 gallery_images 表，存储用户上传的图片元数据
//
// 设计思路：
//   - 文件实际存储在磁盘 uploads/gallery/user_{id}/{date}/ 目录
//   - FilePath 存储相对路径，便于迁移和备份
//   - UploadDate 单独存日期字符串，用于按日期分组查询
//   - IsDeleted + DeletedAt 实现回收站功能（软删除）
//   - OCRText 存储识别的文字内容，支持文字搜索图片
//   - Width/Height 记录图片尺寸，用于前端布局计算
//
// 数据库表结构（由 GORM AutoMigrate 自动创建）：
//
//	| 列名          | 类型          | 说明                        |
//	|---------------|---------------|----------------------------|
//	| id            | bigint unsigned| 自增主键                    |
//	| created_at    | datetime(3)   | 创建时间                    |
//	| updated_at    | datetime(3)   | 更新时间                    |
//	| deleted_at    | datetime(3)   | 软删除时间                  |
//	| user_id       | bigint unsigned| 关联用户ID（索引）           |
//	| file_path     | varchar(255)  | 文件相对路径                |
//	| file_name     | varchar(255)  | 原始文件名                  |
//	| file_size     | bigint        | 文件大小(bytes)             |
//	| mime_type     | varchar(50)   | 文件类型(image/jpeg等)      |
//	| upload_date   | date          | 上传日期 YYYY-MM-DD（索引）  |
//	| album_id      | bigint unsigned| 所属专辑ID（可为空）        |
//	| description   | varchar(500)  | 图片描述                    |
//	| tags          | varchar(255)  | 标签，逗号分隔              |
//	| is_deleted    | tinyint(1)    | 是否删除（0=否, 1=是）      |
//	| deleted_at    | datetime(3)   | 删除时间（用于回收站）       |
//	| is_favorite   | tinyint(1)    | 是否收藏                    |
//	| width         | int           | 图片宽度                    |
//	| height        | int           | 图片高度                    |
//	| ocr_text      | text          | OCR识别的文字内容           |
type GalleryImage struct {
	gorm.Model
	UserID      uint       `gorm:"index;not null" json:"user_id"`               // 关联用户ID
	FilePath    string     `gorm:"type:varchar(255);not null" json:"file_path"` // 文件相对路径
	FileName    string     `gorm:"type:varchar(255);not null" json:"file_name"` // 原始文件名
	FileSize    int64      `gorm:"not null" json:"file_size"`                   // 文件大小（字节）
	MimeType    string     `gorm:"type:varchar(50)" json:"mime_type"`           // 文件类型
	UploadDate  string     `gorm:"type:date;index;not null" json:"upload_date"` // 上传日期 YYYY-MM-DD
	AlbumID     *uint      `gorm:"index" json:"album_id"`                       // 所属专辑ID（可为空）
	Description string     `gorm:"type:varchar(500)" json:"description"`        // 图片描述
	Tags        string     `gorm:"type:varchar(255)" json:"tags"`               // 标签，逗号分隔
	IsDeleted   bool       `gorm:"default:false;index" json:"is_deleted"`       // 软删除标记
	DeletedTime *time.Time `json:"deleted_time"`                                // 删除时间（用于回收站30天清理）
	IsFavorite  bool       `gorm:"default:false;index" json:"is_favorite"`      // 收藏标记
	Width       int        `json:"width"`                                       // 图片宽度
	Height      int        `json:"height"`                                      // 图片高度
	OCRText     string     `gorm:"type:text" json:"ocr_text"`                   // OCR识别的文字内容
}

// TableName 指定 GalleryImage 对应的数据库表名
func (GalleryImage) TableName() string {
	return "gallery_images"
}

// GalleryAlbum 图片专辑结构体
// 对应数据库 gallery_albums 表，存储用户创建的图片专辑
//
// 设计思路：
//   - 每个用户可以创建多个专辑来分类管理图片
//   - CoverImageID 指向专辑封面图片，可为空（为空时显示默认封面）
//   - SortOrder 控制专辑的显示顺序
//   - 删除专辑时，其中的图片变为"未分类"状态（AlbumID设为NULL）
//
// 数据库表结构（由 GORM AutoMigrate 自动创建）：
//
//	| 列名           | 类型          | 说明                        |
//	|----------------|---------------|----------------------------|
//	| id             | bigint unsigned| 自增主键                    |
//	| created_at     | datetime(3)   | 创建时间                    |
//	| updated_at     | datetime(3)   | 更新时间                    |
//	| deleted_at     | datetime(3)   | 软删除时间                  |
//	| user_id        | bigint unsigned| 关联用户ID（索引）           |
//	| name           | varchar(100)  | 专辑名称                    |
//	| description    | varchar(500)  | 专辑描述                    |
//	| cover_image_id | bigint unsigned| 封面图片ID（可为空）         |
//	| sort_order     | int           | 排序序号（默认0）            |
type GalleryAlbum struct {
	gorm.Model
	UserID       uint   `gorm:"index;not null" json:"user_id"`          // 关联用户ID
	Name         string `gorm:"type:varchar(100);not null" json:"name"` // 专辑名称
	Description  string `gorm:"type:varchar(500)" json:"description"`   // 专辑描述
	CoverImageID *uint  `json:"cover_image_id"`                         // 封面图片ID（可为空）
	SortOrder    int    `gorm:"default:0" json:"sort_order"`            // 排序序号
}

// TableName 指定 GalleryAlbum 对应的数据库表名
func (GalleryAlbum) TableName() string {
	return "gallery_albums"
}

// ---- 图库 API 请求/响应结构体 ----

// UploadImageRequest 图片上传请求（multipart/form-data）
// 前端 POST /api/gallery/upload 时提交的表单数据
type UploadImageRequest struct {
	Description string `form:"description"` // 可选，图片描述
	AlbumID     uint   `form:"album_id"`    // 可选，指定专辑ID
	Tags        string `form:"tags"`        // 可选，标签（逗号分隔）
}

// UploadImageResponse 图片上传成功响应
type UploadImageResponse struct {
	ID         uint   `json:"id"`          // 图片ID
	FilePath   string `json:"file_path"`   // 文件路径
	UploadDate string `json:"upload_date"` // 上传日期
}

// ImageListQuery 图片列表查询参数
// GET /api/gallery/images 的查询参数
type ImageListQuery struct {
	Date     string `form:"date"`      // 可选，按日期筛选 YYYY-MM-DD
	AlbumID  uint   `form:"album_id"`  // 可选，按专辑筛选
	Tag      string `form:"tag"`       // 可选，按标签筛选
	Favorite bool   `form:"favorite"`  // 可选，只看收藏
	Page     int    `form:"page"`      // 页码，默认1
	PageSize int    `form:"page_size"` // 每页数量，默认50
}

// ImageListResponse 图片列表响应
type ImageListResponse struct {
	List       []GalleryImageItem `json:"list"`        // 图片列表
	Total      int64              `json:"total"`       // 总数
	Page       int                `json:"page"`        // 当前页
	PageSize   int                `json:"page_size"`   // 每页数量
	TotalPages int                `json:"total_pages"` // 总页数
}

// GalleryImageItem 图片列表项（简化版，不含OCR等敏感信息）
type GalleryImageItem struct {
	ID          uint   `json:"id"`          // 图片ID
	FilePath    string `json:"file_path"`   // 文件路径
	ThumbPath   string `json:"thumb_path"`  // 缩略图路径
	FileName    string `json:"file_name"`   // 原始文件名
	FileSize    int64  `json:"file_size"`   // 文件大小
	UploadDate  string `json:"upload_date"` // 上传日期
	AlbumID     *uint  `json:"album_id"`    // 专辑ID
	Description string `json:"description"` // 描述
	Tags        string `json:"tags"`        // 标签
	IsFavorite  bool   `json:"is_favorite"` // 是否收藏
	Width       int    `json:"width"`       // 宽度
	Height      int    `json:"height"`      // 高度
	CreatedAt   string `json:"created_at"`  // 创建时间
}

// CreateAlbumRequest 创建专辑请求
// POST /api/gallery/albums
type CreateAlbumRequest struct {
	Name        string `json:"name" binding:"required"` // 专辑名称（必填）
	Description string `json:"description"`             // 专辑描述（可选）
}

// UpdateAlbumRequest 更新专辑请求
// PUT /api/gallery/albums/:id
type UpdateAlbumRequest struct {
	Name         string `json:"name"`           // 专辑名称（可选）
	Description  string `json:"description"`    // 专辑描述（可选）
	CoverImageID *uint  `json:"cover_image_id"` // 封面图片ID（可选）
}

// AlbumResponse 专辑响应数据
type AlbumResponse struct {
	ID           uint   `json:"id"`             // 专辑ID
	Name         string `json:"name"`           // 专辑名称
	Description  string `json:"description"`    // 专辑描述
	CoverImageID *uint  `json:"cover_image_id"` // 封面图片ID
	ImageCount   int64  `json:"image_count"`    // 图片数量
	SortOrder    int    `json:"sort_order"`     // 排序
	CreatedAt    string `json:"created_at"`     // 创建时间
}

// BatchDeleteRequest 批量删除请求
// POST /api/gallery/batch-delete
type BatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"` // 要删除的图片ID列表
}

// BatchMoveRequest 批量移动到专辑请求
// POST /api/gallery/batch-move
type BatchMoveRequest struct {
	IDs     []uint `json:"ids" binding:"required,min=1"` // 要移动的图片ID列表
	AlbumID *uint  `json:"album_id"`                     // 目标专辑ID（nil表示移出专辑）
}

// UpdateImageRequest 更新图片信息请求
// PUT /api/gallery/images/:id
type UpdateImageRequest struct {
	Description *string `json:"description"` // 描述（可选）
	Tags        *string `json:"tags"`        // 标签（可选）
	AlbumID     *uint   `json:"album_id"`    // 专辑ID（可选）
	IsFavorite  *bool   `json:"is_favorite"` // 收藏状态（可选）
}

// SearchImagesRequest 搜索图片请求
// GET /api/gallery/search?q=关键词
type SearchImagesRequest struct {
	Query    string `form:"q" binding:"required"` // 搜索关键词
	Page     int    `form:"page"`                 // 页码
	PageSize int    `form:"page_size"`            // 每页数量
}

// RecycleBinQuery 回收站查询参数
type RecycleBinQuery struct {
	Page     int `form:"page"`      // 页码
	PageSize int `form:"page_size"` // 每页数量
}

// DateGroupItem 按日期分组的图片数据
// 用于前端日期视图展示
type DateGroupItem struct {
	Date   string             `json:"date"`   // 日期 YYYY-MM-DD
	Count  int64              `json:"count"`  // 该日期图片数量
	Images []GalleryImageItem `json:"images"` // 图片列表
}

// DateGroupResponse 日期分组响应
type DateGroupResponse struct {
	Groups     []DateGroupItem `json:"groups"`      // 日期分组列表
	Total      int64           `json:"total"`       // 总图片数
	TotalDates int             `json:"total_dates"` // 总日期数
}
