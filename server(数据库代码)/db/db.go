// Package db 提供数据库连接管理和初始化功能
// =============================================
// 作用：
//   1. 使用 GORM 建立和管理与 MySQL 的双数据库连接（主库 + 测试库）
//   2. 应用启动时通过 AutoMigrate 自动创建/更新数据表（无需手写建表 SQL）
//   3. 提供全局的数据库访问入口
//
// Day 1 更新日志（2026-04-07）：
//   - 从 database/sql 迁移到 GORM ORM 框架
//   - 支持双数据库连接：mainDB（主库 tinyweb1）和 testDB（测试库 tinyweb1_test）
//   - 使用 GORM AutoMigrate 替代手动 CREATE TABLE 语句
//   - 新增 visit_stats 表的自动迁移
//   - 保留 GetDB() 返回 *gorm.DB，供所有 handler 使用
//
// GORM 相比 database/sql 的优势：
//   - 自动建表：结构体定义即可，不用写 SQL
//   - CRUD 简化：db.Create(&obj)、db.Save(&obj)、db.Delete(&obj)
//   - 类型安全：查询结果直接映射到 Go 结构体
//   - 软删除：gorm.Model 自带 DeletedAt 字段
//   - 连接池：GORM 底层仍使用 database/sql 的连接池
//
// 使用方式：
//   在 main.go 中调用 db.Initialize() 初始化数据库连接，
//   在各 handler 文件中调用 db.GetDB() 获取 GORM 数据库实例。
//
// 连接池配置：
//   - 最大打开连接数：10（同时最多10个活跃数据库连接）
//   - 最大空闲连接数：5（保持5个空闲连接以备复用）
//   - 连接最大生命周期：30分钟（防止长时间占用连接）
//
// 自动迁移的表（AutoMigrate）：
//   - visit_stats: 访问统计表（Day 1 新增，GORM 管理）
// =============================================

package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tinyweb1/config"
	"tinyweb1/model"
)

// mainDB 全局主数据库 GORM 实例（包内私有）
// 对应 tinyweb1 数据库，用于生产环境
var mainDB *gorm.DB

// testDB 全局测试数据库 GORM 实例（包内私有）
// 对应 tinyweb1_test 数据库，用于开发和测试
var testDB *gorm.DB

// Initialize 初始化主数据库连接并执行自动迁移
// 此函数应在程序启动时（main.go 中）调用一次
//
// 执行步骤：
//   1. 从 config 获取主库 DSN 连接字符串
//   2. 建立 GORM 连接并配置连接池参数
//   3. 测试连接是否正常
//   4. 使用 AutoMigrate 自动创建/更新数据表
func Initialize() {
	mainDB = connectDB(
		config.GetDSN(),
		config.GetDBHost(),
		config.GetDBPort(),
		config.GetDBName(),
		"主库",
	)

	// 主库自动迁移：根据结构体定义自动创建/更新表结构
	// AutoMigrate 只会添加缺失的列和索引，不会删除已有列，安全无侵入
	autoMigrateMainDB(mainDB)

	fmt.Println("✅ 主数据库初始化完成!")
}

// InitializeTestDB 初始化测试数据库连接并执行自动迁移
// 与 Initialize() 类似，但连接的是测试库（tinyweb1_test）
// 用于开发和测试阶段验证数据库功能
func InitializeTestDB() {
	testDB = connectDB(
		config.GetTestDSN(),
		config.GetTestDBHost(),
		config.GetTestDBPort(),
		config.GetTestDBName(),
		"测试库",
	)

	// 测试库也执行同样的自动迁移，保持表结构一致
	autoMigrateMainDB(testDB)

	fmt.Println("✅ 测试数据库初始化完成!")
}

// ============================================================
// 数据库连接相关
// ============================================================

// connectDB 通用的数据库连接函数
// 根据 DSN 连接字符串建立 GORM 数据库连接，配置连接池并验证连通性
//
// 参数：
//   - dsn: MySQL 数据源名称（格式：用户:密码@tcp(地址:端口)/数据库?参数）
//   - host: MySQL 服务器地址（用于日志输出）
//   - port: MySQL 端口（用于日志输出）
//   - name: 数据库名称（用于日志输出）
//   - label: 数据库标签，如 "主库" 或 "测试库"（用于日志输出）
//
// 返回：
//   - *gorm.DB: GORM 数据库实例
//   - 失败时 log.Fatal 终止程序
func connectDB(dsn, host, port, name, label string) *gorm.DB {
	// 自动创建数据库（如果不存在）
	// 连接 MySQL 时不指定数据库名，执行 CREATE DATABASE IF NOT EXISTS
	createDBDSN := fmt.Sprintf("%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		config.GetDBUser()+":"+config.GetDBPass(), host, port)
	if label == "测试库" {
		createDBDSN = fmt.Sprintf("%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
			config.GetTestDBUser()+":"+config.GetTestDBPass(), host, port)
	}

	rawDB, err := sql.Open("mysql", createDBDSN)
	if err != nil {
		log.Fatalf("❌ 无法连接 MySQL 服务器 [%s:%s] 以创建%s: %v", host, port, label, err)
	}
	defer rawDB.Close()

	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", name)
	if _, err = rawDB.Exec(createSQL); err != nil {
		log.Fatalf("❌ 自动创建%s [%s] 失败: %v", label, name, err)
	}
	fmt.Printf("📦 %s [%s] 已就绪\n", label, name)

	// 设置 GORM 日志级别
	// development 模式下显示所有 SQL 日志，production 模式下只显示错误
	gormConfig := &gorm.Config{}
	if config.IsDevelopment() {
		// 开发模式：显示详细 SQL 日志，方便调试
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		// 生产模式：只记录错误日志，减少性能开销
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	// 使用 GORM 的 MySQL 驱动打开数据库连接
	// 内部仍使用 database/sql 的连接池机制
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("❌ 无法连接%s [%s:%s/%s]: %v", label, host, port, name, err)
	}

	// 获取底层的 *sql.DB 来配置连接池参数
	// GORM 基于 database/sql，所以连接池配置方式相同
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ 无法获取%s底层连接池: %v", label, err)
	}

	// 配置连接池参数
	sqlDB.SetMaxOpenConns(10)                  // 最大同时打开的连接数（防止连接数爆炸）
	sqlDB.SetMaxIdleConns(5)                   // 最大空闲连接数（保持一定数量的空闲连接复用）
	sqlDB.SetConnMaxLifetime(30 * time.Minute)  // 连接最大生存时间（防止长时间占用或过期连接）

	// 验证数据库连接是否可用（真正发起一次 TCP 连接到 MySQL）
	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("❌ %s连接测试失败: %v", label, err)
	}

	fmt.Printf("✅ %s连接成功! [%s:%s/%s]\n", label, host, port, name)

	return db
}

// ============================================================
// 获取数据库实例
// ============================================================

// GetDB 获取主数据库 GORM 实例
// 供各个 handler 使用来执行数据库操作
// 注意：必须在 Initialize() 之后调用
func GetDB() *gorm.DB {
	return mainDB
}

// GetTestDB 获取测试数据库 GORM 实例
// 仅供测试代码使用
// 注意：必须在 InitializeTestDB() 之后调用
func GetTestDB() *gorm.DB {
	return testDB
}

// ============================================================
// 自动迁移（AutoMigrate）
// ============================================================

// autoMigrateMainDB 对主库执行表结构自动迁移
//
// AutoMigrate 工作原理：
//   1. 读取 Go 结构体的字段和 gorm 标签
//   2. 检查数据库中是否存在对应的表
//   3. 如果表不存在，自动 CREATE TABLE
//   4. 如果表存在但缺少某些列，自动 ALTER TABLE ADD COLUMN
//   5. 不会删除已有的列或数据，安全无侵入
//
// 当前迁移的表：
//   - visit_stats: 访问统计表（GORM 管理）
func autoMigrateMainDB(db *gorm.DB) {
	// 将所有需要 GORM 管理的模型传入 AutoMigrate
	// GORM 会根据结构体定义自动创建对应的表
	err := db.AutoMigrate(
		&model.VisitStats{},   // 访问统计表
		&model.User{},         // 用户表（注册登录功能新增）
		&model.StudySession{}, // 专注记录表（专注时间功能新增）
		&model.StudyTag{},     // 专注标签表（专注时间功能新增）
	)
	if err != nil {
		log.Fatalf("❌ 数据库自动迁移失败: %v", err)
	}
	fmt.Println("✅ 数据库表结构迁移完成!")
}
