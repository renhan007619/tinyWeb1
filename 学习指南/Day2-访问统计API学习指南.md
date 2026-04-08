# Day 2 学习指南 — 访问统计 API 开发

> **完成日期**：2026-04-08
> **目标**：从零实现访问统计的 RESTful API，掌握 GORM CRUD 操作和 Upsert 模式

---

## 📋 今天完成了什么

| # | 任务 | 涉及文件 | 状态 |
|---|------|---------|------|
| 1 | 定义 VisitRecord / VisitStatsResponse 数据模型 | `model/model.go` | ✅ |
| 2 | 创建 `handler/helpers.go` 共享工具函数（sendJSON、trimString、parseIntQueryParam） | `handler/helpers.go` | ✅ |
| 3 | 实现 POST /api/visit（记录访问 + Upsert 逻辑） | `handler/visit.go` RecordVisit() | ✅ |
| 4 | 实现 GET /api/visit/stats（统计汇总接口） | `handler/visit.go` GetVisitStats() | ✅ |
| 5 | 在 main.go 中注册路由 + 编译测试通过 | `main.go` | ✅ |

---

## 🎯 最终效果

```bash
# 记录一次首次访问
curl -X POST http://localhost:8081/api/visit \
  -H "Content-Type: application/json" \
  -d '{"visitor_ip":"10.0.0.1","device_type":"desktop","browser":"Chrome"}'
返回 → {"code":0,"message":"success","data":{"is_first_visit":true,"visit_count":1}}

# 同一 IP 再次访问（Upsert 更新）
curl -X POST http://localhost:8081/api/visit \
  -H "Content-Type: application/json" \
  -d '{"visitor_ip":"10.0.0.1","device_type":"mobile"}'
返回 → {"code":0,"message":"success","data":{"is_first_visit":false,"visit_count":2}}

# 获取统计汇总
curl http://localhost:8081/api/visit/stats
返回 → {"code":0,"message":"success","data":{"total_visits":2,"unique_visitors":1,"last_visit_at":"2026-04-08 18:54:56"}}
```

---

## 📚 核心知识点

### 1️⃣ 什么是 Upsert？（今日核心概念）

**Upsert = Update + Insert**，即"有则更新，无则插入"。

```go
// 传统做法（需要两条 SQL）
// ① SELECT * FROM visit_stats WHERE visitor_ip = '10.0.0.1'
// ② 如果有结果 → UPDATE SET visit_count = visit_count + 1
//    如果没结果 → INSERT INTO visit_stats (ip, count) VALUES ('10.0.0.1', 1)

// GORM 做法（代码层面实现 Upsert）
result := database.Where("visitor_ip = ?", ip).First(&existing)
if result.Error == gorm.ErrRecordNotFound {
    // 没找到 → Create（插入新记录）
    database.Create(&newRecord)
} else {
    // 找到了 → Updates（更新已有记录）
    database.Model(&existing).Updates(map[string]interface{}{
        "visit_count": gorm.Expr("visit_count + 1"), // SQL 表达式，原子操作
        "last_visit_at": time.Now(),
    })
}
```

**对应代码位置**：`handler/visit.go` → `RecordVisit()` 函数（第 55-135 行）

---

### 2️⃣ GORM 基础 CRUD 操作速查

| 操作 | GORM 写法 | 对应 SQL | 本项目使用位置 |
|------|----------|---------|--------------|
| **Create 插入** | `db.Create(&record)` | `INSERT INTO ...` | RecordVisit 首次访问 |
| **First 查询一条** | `db.Where("ip=?",ip).First(&obj)` | `SELECT ... WHERE ... LIMIT 1` | RecordVisit 查找已有记录 |
| **Updates 更新** | `db.Model(&obj).Updates(map{...})` | `UPDATE ... SET ...` | RecordVisit 更新计数 |
| **Select 聚合** | `db.Model(&M{}).Select("SUM(x)").Scan(&v)` | `SELECT SUM(x) FROM ...` | GetVisitStats 统计汇总 |
| **Count 计数** | `db.Model(&M{}).Select("COUNT(*)").Scan(&v)` | `SELECT COUNT(*) FROM ...` | GetVisitStats 独立访客数 |

---

### 3️⃣ GORM.Expr — 在 GORM 中使用原生 SQL 表达式

```go
// ❌ 错误写法：这样会把 visit_count 设为固定值 2
database.Model(&found).Updates(map[string]interface{}{
    "visit_count": found.VisitCount + 1, // Go 层面计算，并发不安全！
})

// ✅ 正确写法：使用 GORM.Expr 执行 SQL 原子操作
database.Model(&found).Updates(map[string]interface{}{
    "visit_count": gorm.Expr("visit_count + 1"), // MySQL 层面执行 visit_count = visit_count + 1
})
```

**为什么重要？**
- 并发安全：两个请求同时访问时，不会丢失计数
- 原子操作：MySQL 保证 `visit_count + 1` 是原子执行的
- 性能更好：不需要先读再写，一条 SQL 搞定

**对应代码位置**：`handler/visit.go` 第 112-118 行

---

### 4️⃣ HTTP 方法路由分发

```go
// main.go 中的路由注册方式：同一个 URL 根据方法分发到不同 handler
mux.HandleFunc("/api/visit", func(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        handler.RecordVisit(w, r)      // POST → 记录访问
    } else {
        sendMethodNotAllowed(w)         // 其他方法 → 返回 405
    }
})

mux.HandleFunc("/api/visit/stats", func(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        handler.GetVisitStats(w, r)     // GET → 获取统计
    } else {
        sendMethodNotAllowed(w)
    }
})
```

**RESTful 设计规范**：
- `POST /api/visit` — 创建资源（记录一次访问）
- `GET /api/visit/stats` — 读取聚合数据（不是单个资源，用复数名词）

---

### 5️⃣ JSON 请求体解析与校验模式

```go
// Step 1: 解析 JSON 请求体到结构体
var req model.VisitRecord
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    sendJSON(w, 400, model.ErrorResponse(400, "请求数据格式错误"))
    return
}

// Step 2: 必填字段校验
req.VisitorIP = trimString(req.VisitorIP)
if req.VisitorIP == "" {
    sendJSON(w, 400, model.ErrorResponse(400, "缺少 visitor_ip 参数"))
    return
}
```

**这个模式的优点**：
- 快速失败：数据不对立即返回错误，不会继续执行有问题的逻辑
- 统一响应格式：所有错误都用 `{code, message, data}` 格式
- 安全：防止空值或恶意数据进入数据库

---

## 📖 项目文件变更说明

### 新增文件

```
server(数据库代码)/handler/
├── visit.go       ← 🆕 访问统计 API（Day 2 核心）
├── helpers.go     ← 🆕 共享工具函数（sendJSON/trimString/parseIntQueryParam）
├── todo.go        ← ⛔ 已禁用（待 GORM 迁移）
├── guestbook.go   ← ⛔ 已禁用（待 GORM 迁移）
└── setting.go     ← ⛔ 已禁用（待 GORM 迁移）
```

### 为什么旧 handler 被禁用？

```
问题：旧 handler 使用 database/sql 接口：
  db.Query(...)           ← *sql.DB 的方法
  db.QueryRow(...)        ← *sql.DB 的方法
  db.Exec(...)            ← *sql.DB 的方法（返回两个值）

但现在 db.GetDB() 返回的是 *gorm.DB：
  gorm.DB 没有 Query / QueryRow 方法
  gorm.Exec 只返回一个值

解决：在旧文件头部加了 //go:build ignore
计划：Day 3-4 将它们迁移为 GORM 版本后重新启用
```

### 修改文件

| 文件 | 改动内容 |
|------|---------|
| `model/model.go` | 新增 VisitRecord、VisitStatsResponse 结构体 |
| `main.go` | import handler 包，注册 `/api/visit` 和 `/api/visit/stats` 路由，添加 sendMethodNotAllowed 辅助函数 |

---

## 🔍 关键代码解读

### RecordVisit 完整流程图

```
POST /api/visit
    │
    ▼
┌─────────────────────┐
│ 1. 解析 JSON 请求体  │  json.NewDecoder(r.Body).Decode(&req)
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ 2. 校验 visitor_ip  │  必填，不能为空
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ 3. SELECT 查询 IP   │  db.Where("ip=?").First(&existing)
└──────────┬──────────┘
           │
     ┌─────┴─────┐
     ▼           ▼
  找不到         找到了
     │             │
     ▼             ▼
  CREATE       Updates
  (插入新记录)  (count+1, update time)
     │             │
     └─────┬───────┘
           ▼
  返回 { is_first_visit, visit_count }
```

### GetVisitStats 三条 SQL

```sql
-- ① 总访问次数（所有 IP 的 count 累加和）
SELECT COALESCE(SUM(visit_count), 0) FROM visit_stats WHERE deleted_at IS NULL;

-- ② 独立访客数（不同 IP 的数量）
SELECT COUNT(*) FROM visit_stats WHERE deleted_at IS NULL;

-- ③ 最后访问时间
SELECT MAX(last_visit_at) FROM visit_stats WHERE deleted_at IS NULL;
```

> **注意**：GORM 默认开启软删除，所有查询自动带 `WHERE deleted_at IS NULL`。

---

## 💡 今日收获 vs Day 1 对比

| 维度 | Day 1 | Day 2 |
|------|-------|-------|
| **重点** | 项目搭建、环境配置、理解架构 | 实际业务 API 开发 |
| **GORM 用法** | AutoMigrate 自动建表 | CRUD 增删改查 + 聚合查询 |
| **HTTP 处理** | 一个简单的 health check handler | 完整的 RESTful API（POST + GET） |
| **核心模式** | 初始化流程、中间件 | Upsert 模式、JSON 解析校验、路由分发 |
| **代码量** | ~320 行 main.go | ~180 行 visit.go + helpers.go |

---

## 🗺️ 下一步学习方向（Day 3 预告）

完成 Day 2 后，你已经掌握了：
- ✅ GORM 全套 CRUD 操作
- ✅ Upsert 模式（有则更新，无则插入）
- ✅ RESTful API 设计与实现
- ✅ JSON 请求解析与统一响应格式
- ✅ HTTP 方法的路由分发

**Day 3 建议**：
1. 将 `handler/todo.go`（备忘录功能）从 database/sql 迁移到 GORM
2. 学习 GORM 事务处理（替代手动 tx.Begin/Commit）
3. 了解 GORM 预加载（Preload）和关联查询

---

*本指南由 CodeBuddy AI 生成，如有疑问随时提问！*
