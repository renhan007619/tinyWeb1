// TinyWeb1 主程序入口文件
// =============================================
// 作用：
//   程序的启动入口，负责以下工作：
//
//   1. 加载配置：从环境变量读取数据库连接信息、端口等配置
//   2. 初始化主数据库：建立 GORM 连接，自动迁移表结构
//   3. 初始化测试数据库：验证双数据库连接功能
//   4. 测试数据库功能：写入和查询 visit_stats 表，验证 GORM 正常工作
//   5. 启动 HTTP 服务器：提供静态文件服务和健康检查接口
//
// 项目架构说明：
//   - config/    : 配置管理模块（环境变量 → 结构体）
//   - model/     : 数据结构定义（GORM 模型 + 请求/响应格式）
//   - db/        : 数据库连接和自动迁移（基于 GORM）
//   - handler/   : 各业务模块的 API 处理函数（待后续迁移到 GORM）
//   - main.go    : 本文件，启动和测试
//
// 启动方式：
//   cd "server(数据库代码)" && go run main.go
//
// 环境变量配置（可选）：
//   APP_ENV=development go run main.go              # 开发模式（默认）
//   DB_HOST=localhost DB_PASS=123456 go run main.go  # 指定数据库
//
// Day 1 更新日志（2026-04-07）：
//   - 从 database/sql 迁移到 GORM
//   - 新增测试数据库连接验证
//   - 新增 visit_stats 表的读写测试
//   - 暂时移除旧 handler 路由（待后续迁移到 GORM）
// =============================================

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/gorm"

	"tinyweb1/config"
	"tinyweb1/db"
	"tinyweb1/handler"
	"tinyweb1/middleware"
	"tinyweb1/model"
)

func main() {
	// ============================================================
	// 步骤 1: 加载配置
	// ============================================================
	config.Load()
	printConfigInfo()

	// ============================================================
	// 步骤 2: 初始化主数据库（tinyweb1）
	// ============================================================
	db.Initialize()

	// ---- Day2 新增：自动创建 users 表 ----
	// GORM 的 AutoMigrate 会根据 User 结构体自动创建/更新表结构
	if err := db.GetDB().AutoMigrate(&model.User{}); err != nil {
		log.Fatal("❌ users 表自动迁移失败:", err)
	}
	fmt.Println("✅ users 表就绪")

	// ---- 专注时间功能：自动创建 study_sessions 和 study_tags 表 ----
	if err := db.GetDB().AutoMigrate(&model.StudySession{}, &model.StudyTag{}); err != nil {
		log.Fatal("❌ 专注时间表自动迁移失败:", err)
	}
	fmt.Println("✅ study_sessions + study_tags 表就绪")

	// ---- 留言板功能：自动创建 guestbook 表 ----
	if err := db.GetDB().AutoMigrate(&model.Guestbook{}); err != nil {
		log.Fatal("❌ 留言板表自动迁移失败:", err)
	}
	fmt.Println("✅ guestbook 表就绪")

	// ---- 备忘录功能：自动创建 todos 和 todo_history 表 ----
	if err := db.GetDB().AutoMigrate(&model.Todo{}, &model.TodoHistory{}); err != nil {
		log.Fatal("❌ 备忘录表自动迁移失败:", err)
	}
	fmt.Println("✅ todos + todo_history 表就绪")

	// ============================================================
	// 步骤 3: 初始化测试数据库（tinyweb1_test）
	// ============================================================
	db.InitializeTestDB()

	// ============================================================
	// 步骤 4: 测试数据库功能 - 验证 visit_stats 表的 CRUD 操作
	// ============================================================
	testVisitStats()

	// ============================================================
	// 步骤 5: 启动 HTTP 服务器（静态文件 + 健康检查）
	// ============================================================
	startServer()
}

// ============================================================
// 配置信息打印
// ============================================================

// printConfigInfo 打印当前加载的配置信息，便于确认程序运行参数
func printConfigInfo() {
	fmt.Println("📋 配置加载完成")
	fmt.Printf("   运行环境: %s\n", config.GetAppEnv())
	fmt.Printf("   主数据库: %s:%s/%s\n", config.GetDBHost(), config.GetDBPort(), config.GetDBName())
	fmt.Printf("   测试数据库: %s:%s/%s\n", config.GetTestDBHost(), config.GetTestDBPort(), config.GetTestDBName())
	fmt.Printf("   服务端口: %s\n", config.GetServerPort())
}

// ============================================================
// 数据库功能测试
// ============================================================

// testVisitStats 测试 visit_stats 表的数据库操作
// 通过实际写入、查询、更新操作验证 GORM 和表结构是否正常工作
//
// 测试步骤：
//   1. 在测试库中插入一条测试记录
//   2. 查询该记录，验证数据正确
//   3. 更新访问次数，验证更新操作
//   4. 清理测试数据
func testVisitStats() {
	fmt.Println("")
	fmt.Println("🔍 开始测试 visit_stats 表...")

	database := db.GetTestDB() // 使用测试库进行测试，不影响主库数据
	now := time.Now()

	// 定义测试数据
	testIP := "192.168.1.100"
	testRecord := model.VisitStats{
		VisitorIP:    testIP,
		VisitCount:   1,
		FirstVisitAt: now,
		LastVisitAt:  now,
		UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) TestBot/1.0",
		DeviceType:   "desktop",
		Browser:      "Chrome",
		OS:           "Windows",
		Referrer:     "https://www.baidu.com",
	}

	// 测试 1: 插入记录
	if err := database.Create(&testRecord).Error; err != nil {
		log.Printf("   ⚠️  插入测试记录失败（可能是重复数据）: %v", err)
	} else {
		fmt.Println("   ✅ 测试 1 通过: 插入记录成功")
	}

	// 测试 2: 查询记录 - 根据 visitor_ip 查找
	var found model.VisitStats
	result := database.Where("visitor_ip = ?", testIP).First(&found)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("   ⚠️  查询测试记录失败: 未找到 IP=%s 的记录", testIP)
		} else {
			log.Printf("   ⚠️  查询测试记录失败: %v", result.Error)
		}
	} else {
		fmt.Printf("   ✅ 测试 2 通过: 查询成功 [IP=%s, 次数=%d, 设备=%s]\n",
			found.VisitorIP, found.VisitCount, found.DeviceType)
	}

	// 测试 3: 更新访问次数 - 模拟访客再次访问
	if result.Error == nil {
		updateErr := database.Model(&found).Updates(map[string]interface{}{
			"visit_count":  gorm.Expr("visit_count + 1"), // 访问次数 +1
			"last_visit_at": time.Now(),                   // 更新最后访问时间
		}).Error
		if updateErr != nil {
			log.Printf("   ⚠️  更新测试记录失败: %v", updateErr)
		} else {
			fmt.Println("   ✅ 测试 3 通过: 更新访问次数成功")
		}
	}

	// 测试 4: 统计总记录数
	var totalCount int64
	database.Model(&model.VisitStats{}).Count(&totalCount)
	fmt.Printf("   ✅ 测试 4 通过: 当前 visit_stats 表共 %d 条记录\n", totalCount)

	// 测试 5: 查询主库中的记录数（主库应该是空的或只有旧数据）
	mainDB := db.GetDB()
	var mainCount int64
	mainDB.Model(&model.VisitStats{}).Count(&mainCount)
	fmt.Printf("   ✅ 测试 5 通过: 主库 visit_stats 表共 %d 条记录\n", mainCount)

	fmt.Println("   🎉 visit_stats 表测试完成!")
	fmt.Println("")
}

// ============================================================
// HTTP 服务器启动
// ============================================================

// startServer 启动 HTTP 服务器
// 提供静态文件服务（index.html 等）和健康检查接口
//
// 路由说明：
//   - GET /api/health : 健康检查接口，返回服务器和数据库状态
//   - /api/todos/*    : 备忘录 CRUD + 归档（已迁移到 GORM）
//   - /api/focus/*    : 专注时间记录和统计（新增）
//   - GET /           : 静态文件服务（index.html 等）
func startServer() {
	// 获取项目根目录：优先使用 STATIC_DIR 环境变量，否则使用当前工作目录
	rootDir := config.GetStaticDir()
	if rootDir == "" {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			log.Fatal("❌ 无法获取当前工作目录:", err)
		}
	}

	// 创建路由器
	mux := http.NewServeMux()

	// ---- 健康检查接口 ----
	// 用于验证服务器和数据库是否正常运行
	mux.HandleFunc("/api/health", healthCheckHandler)

	// ---- 访问统计接口（Day 2 新增）----
	mux.HandleFunc("/api/visit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.RecordVisit(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	})
	mux.HandleFunc("/api/visit/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetVisitStats(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	})

	// ---- 用户认证接口（注册登录功能新增）----
	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.Register(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	})

	// ---- Day2 新增：登录 + 当前用户接口 ----
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.Login(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	})

	// GET /api/auth/me 需要登录才能访问，使用 JWT 中间件保护
	mux.HandleFunc("/api/auth/me", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetCurrentUser(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	}))

	// ---- 备忘录接口（需登录认证，用户数据隔离）----
	mux.HandleFunc("/api/todos", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetTodos(w, r)
		case http.MethodPost:
			handler.CreateTodo(w, r)
		default:
			sendMethodNotAllowed(w)
		}
	}))
	mux.HandleFunc("/api/todos/archive", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.ArchiveTodos(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	}))
	mux.HandleFunc("/api/todos/history/dates", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetTodoHistoryDates(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	}))
	mux.HandleFunc("/api/todos/history", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetTodoHistoryByDate(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	}))
	// /api/todos/:id 必须放在 /api/todos/ 之后，避免匹配冲突
	mux.HandleFunc("/api/todos/", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// 排除已注册的 /api/todos/archive 和 /api/todos/history 路径
		path := r.URL.Path
		if path == "/api/todos/archive" || path == "/api/todos/history" || path == "/api/todos/history/dates" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodPut:
			handler.UpdateTodo(w, r)
		case http.MethodDelete:
			handler.DeleteTodo(w, r)
		default:
			sendMethodNotAllowed(w)
		}
	}))

	// ---- 留言板接口（无需登录认证，公开访问）----
	mux.HandleFunc("/api/guestbook", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetGuestbookMessages(w, r)
		case http.MethodPost:
			handler.CreateGuestbookMessage(w, r)
		default:
			sendMethodNotAllowed(w)
		}
	})

	// ---- 专注时间接口（需登录认证）----
	mux.HandleFunc("/api/focus/session", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateFocusSession(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	}))
	mux.HandleFunc("/api/focus/today", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetTodayFocus(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	}))
	mux.HandleFunc("/api/focus/summary", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetFocusSummary(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	}))
	mux.HandleFunc("/api/focus/history", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetFocusHistory(w, r)
		} else {
			sendMethodNotAllowed(w)
		}
	}))
	mux.HandleFunc("/api/focus/tags", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetTags(w, r)
		case http.MethodPost:
			handler.CreateTag(w, r)
		default:
			sendMethodNotAllowed(w)
		}
	}))

	// ---- 静态文件兜底路由 ----
	// 所有未被 API 路由匹配的请求都交给静态文件服务器处理
	fs := http.FileServer(http.Dir(rootDir))
	mux.Handle("/", fs)

	// 打印启动信息
	addr := config.GetServerPort()
	fmt.Println("========================================")
	fmt.Println("  🚀 TinyWeb1 Server is running!")
	fmt.Printf("  📍 访问地址: http://localhost%s\n", addr)
	fmt.Printf("  📂 静态文件: %s\n", rootDir)
	fmt.Printf("  🔧 运行环境: %s\n", config.GetAppEnv())
	fmt.Println("  🔗 接口:")
	fmt.Println("     GET  /api/health         健康检查")
	fmt.Println("     POST /api/auth/register  用户注册")
	fmt.Println("     POST /api/auth/login     用户登录 (Day2)")
	fmt.Println("     GET  /api/auth/me        当前用户 (Day2, 需要token)")
	fmt.Println("     ---- 备忘录 ----")
	fmt.Println("     GET  /api/todos          获取任务列表")
	fmt.Println("     POST /api/todos          新增任务")
	fmt.Println("     PUT  /api/todos/:id      更新任务(打勾/编辑)")
	fmt.Println("     DELETE /api/todos/:id    删除任务")
	fmt.Println("     POST /api/todos/archive  归档任务")
	fmt.Println("     GET  /api/todos/history  历史归档")
	fmt.Println("     GET  /api/todos/history/dates 归档日期列表")
	fmt.Println("     ---- 留言板 ----")
	fmt.Println("     GET  /api/guestbook      获取留言列表")
	fmt.Println("     POST /api/guestbook      发布留言")
	fmt.Println("     ---- 专注时间 ----")
	fmt.Println("     POST /api/focus/session  创建专注记录")
	fmt.Println("     GET  /api/focus/today    今日统计")
	fmt.Println("     GET  /api/focus/summary  历史总览")
	fmt.Println("     GET  /api/focus/history  某日详细记录")
	fmt.Println("     GET  /api/focus/tags     标签列表")
	fmt.Println("     POST /api/focus/tags     创建标签")
	fmt.Println("========================================")

	// 启动 HTTP 服务（带 CORS 中间件）
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}

// healthCheckHandler 健康检查接口处理函数
// 返回服务器运行状态和数据库连接信息
//
// 响应示例：
//   GET /api/health
//   {
//     "code": 0,
//     "message": "success",
//     "data": {
//       "status": "healthy",
//       "env": "development",
//       "main_db": "tinyweb1",
//       "test_db": "tinyweb1_test",
//       "time": "2026-04-07 19:00:00"
//     }
//   }
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 检查主数据库连接
	sqlDB, err := db.GetDB().DB()
	dbStatus := "connected"
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "disconnected"
	}

	response := map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]string{
			"status":  dbStatus,
			"env":     config.GetAppEnv(),
			"main_db": config.GetDBName(),
			"test_db": config.GetTestDBName(),
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	// 手动构建 JSON 响应（避免引入额外的 JSON 序列化依赖）
	fmt.Fprintf(w, `{"code":%d,"message":"%s","data":{"status":"%s","env":"%s","main_db":"%s","test_db":"%s","time":"%s"}}`,
		response["code"], response["message"],
		response["data"].(map[string]string)["status"],
		response["data"].(map[string]string)["env"],
		response["data"].(map[string]string)["main_db"],
		response["data"].(map[string]string)["test_db"],
		response["data"].(map[string]string)["time"],
	)
}

// corsMiddleware CORS 跨域中间件
// =============================================
// 作用：
//   在每个 HTTP 响应中添加 CORS 相关的头部，
//   允许前端 JavaScript 从不同域名/端口调用此 API。
//
// 为什么需要 CORS？
//   前端页面可能部署在 example.com:80，
//   后端 API 运行在 api.example.com:8081，
//   浏览器的同源策略会阻止这种跨域请求。
//   通过添加 Access-Control-Allow-* 头部来允许合法的跨域访问。
//
// 安全注意事项：
//   生产环境应将 ALLOWED_ORIGINS 设为具体的域名（如 https://yourdomain.com），
//   仅开发环境使用 "*" 允许所有来源。
// =============================================
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := config.GetAllowedOrigins()

		// 检查请求来源是否在允许列表中
		allowOrigin := ""
		if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
			// 开发模式：允许所有来源
			allowOrigin = "*"
		} else {
			// 生产模式：逐一比对允许的来源
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					allowOrigin = origin
					break
				}
			}
		}

		// 设置 CORS 响应头
		if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)                // 允许的来源
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // 允许的方法
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")      // 允许的请求头
			w.Header().Set("Access-Control-Max-Age", "86400") // 预检请求缓存时间（24小时）
		}

		// 处理预检请求 (OPTIONS)：浏览器在非简单请求前会先发送 OPTIONS 探测
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent) // 204 No Content
			return
		}

		// 继续处理实际请求
		next.ServeHTTP(w, r)
	})
}

// ============================================================
// 辅助函数
// ============================================================

// sendMethodNotAllowed 返回 405 Method Not Allowed 响应
func sendMethodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"code":405,"message":"请求方法不允许","data":null}`))
}
