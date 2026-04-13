// Package config 提供配置管理功能
// =============================================
// 作用：
//   统一管理应用程序的所有配置项，包括数据库连接参数、服务器端口、CORS 策略等。
//   配置优先级：环境变量 > 默认值
//
// Day 1 更新日志（2026-04-07）：
//   - 新增 APP_ENV 环境变量，支持 development（开发）和 production（生产）两种模式
//   - 新增测试数据库配置（DB_TEST_HOST 等），支持同时连接主库和测试库
//   - 主库：tinyweb1（生产环境使用）
//   - 测试库：tinyweb1_test（开发测试使用）
//
// 使用方式：
//   在 main.go 中调用 config.Load() 加载配置，通过 config.Get*() 函数获取各项配置值。
//
// 环境变量说明：
//   APP_ENV          - 运行环境：development(默认) / production
//   DB_HOST          - 主库 MySQL 服务器地址（默认 localhost）
//   DB_PORT          - 主库 MySQL 端口（默认 3306）
//   DB_USER          - 主库 MySQL 用户名（默认 root）
//   DB_PASS          - 主库 MySQL 密码（默认空字符串）
//   DB_NAME          - 主库数据库名称（默认 tinyweb1）
//   DB_TEST_HOST     - 测试库 MySQL 服务器地址（默认 localhost）
//   DB_TEST_PORT     - 测试库 MySQL 端口（默认 3306）
//   DB_TEST_USER     - 测试库 MySQL 用户名（默认 root）
//   DB_TEST_PASS     - 测试库 MySQL 密码（默认空字符串）
//   DB_TEST_NAME     - 测试库数据库名称（默认 tinyweb1_test）
//   SERVER_PORT      - HTTP 服务端口（默认 :8080）
//   ALLOWED_ORIGINS  - 允许的跨域来源，逗号分隔（开发环境默认 *）
// =============================================

package config

import (
	"os"
	"strings"
)

// DBConfig 单个数据库的连接配置
// 将主库和测试库的配置抽象为统一结构，便于管理
type DBConfig struct {
	Host string // MySQL 服务器地址
	Port string // MySQL 端口
	User string // MySQL 用户名
	Pass string // MySQL 密码
	Name string // 数据库名称
}

// AppConfig 应用程序全局配置结构体
// 包含所有运行时需要的配置项
type AppConfig struct {
	// 运行环境
	AppEnv string // 当前运行模式：development(开发) 或 production(生产)

	// 数据库配置
	MainDB  DBConfig // 主数据库配置（生产环境使用的数据库）
	TestDB  DBConfig // 测试数据库配置（开发和测试使用的数据库）

	// 服务器配置
	ServerPort     string   // HTTP 监听端口（含冒号，如 ":8080"）
	AllowedOrigins []string // 允许的 CORS 跨域来源列表
	StaticDir      string   // 静态文件目录（为空则使用当前工作目录）
}

// appConfig 全局配置实例（包内私有，通过函数访问）
var appConfig *AppConfig

// Load 加载并初始化所有配置项
// 从环境变量读取配置，如果环境变量未设置则使用默认值
// 应在程序启动时（main.go 中）首先调用此函数
func Load() {
	appConfig = &AppConfig{
		// 运行环境配置
		AppEnv: getEnv("APP_ENV", "development"),

		// 主数据库配置（用于生产环境）
		MainDB: DBConfig{
			Host: getEnv("DB_HOST", "localhost"),
			Port: getEnv("DB_PORT", "3306"),
			User: getEnv("DB_USER", "root"),
			Pass: getEnv("DB_PASS", ""),
			Name: getEnv("DB_NAME", "tinyweb1"),
		},

		// 测试数据库配置（用于开发和测试）
		// 测试库的环境变量带有 _TEST_ 前缀，与主库区分
		TestDB: DBConfig{
			Host: getEnv("DB_TEST_HOST", "localhost"),
			Port: getEnv("DB_TEST_PORT", "3306"),
			User: getEnv("DB_TEST_USER", "root"),
			Pass: getEnv("DB_TEST_PASS", ""),
			Name: getEnv("DB_TEST_NAME", "tinyweb1_test"),
		},

		// 服务器配置
		ServerPort:     ":" + getEnv("SERVER_PORT", "8080"),
		AllowedOrigins: parseOrigins(getEnv("ALLOWED_ORIGINS", "*")),
		StaticDir:      getEnv("STATIC_DIR", ""),
	}
}

// ============================================================
// 运行环境相关
// ============================================================

// GetAppEnv 返回当前运行环境
// "development" = 开发环境（允许所有跨域、详细日志）
// "production" = 生产环境（严格跨域、精简日志）
func GetAppEnv() string {
	return appConfig.AppEnv
}

// IsDevelopment 判断当前是否为开发环境
func IsDevelopment() bool {
	return appConfig.AppEnv == "development"
}

// ============================================================
// 主数据库配置（生产库）
// ============================================================

// GetDBHost 返回主库 MySQL 服务器地址
func GetDBHost() string {
	return appConfig.MainDB.Host
}

// GetDBPort 返回主库 MySQL 端口
func GetDBPort() string {
	return appConfig.MainDB.Port
}

// GetDBUser 返回主库 MySQL 用户名
func GetDBUser() string {
	return appConfig.MainDB.User
}

// GetDBPass 返回主库 MySQL 密码
func GetDBPass() string {
	return appConfig.MainDB.Pass
}

// GetDBName 返回主库数据库名称
func GetDBName() string {
	return appConfig.MainDB.Name
}

// ============================================================
// 测试数据库配置
// ============================================================

// GetTestDBHost 返回测试库 MySQL 服务器地址
func GetTestDBHost() string {
	return appConfig.TestDB.Host
}

// GetTestDBPort 返回测试库 MySQL 端口
func GetTestDBPort() string {
	return appConfig.TestDB.Port
}

// GetTestDBUser 返回测试库 MySQL 用户名
func GetTestDBUser() string {
	return appConfig.TestDB.User
}

// GetTestDBPass 返回测试库 MySQL 密码
func GetTestDBPass() string {
	return appConfig.TestDB.Pass
}

// GetTestDBName 返回测试库数据库名称
func GetTestDBName() string {
	return appConfig.TestDB.Name
}

// ============================================================
// 服务器配置
// ============================================================

// GetServerPort 返回 HTTP 服务监听端口（格式如 ":8080"）
func GetServerPort() string {
	return appConfig.ServerPort
}

// GetAllowedOrigins 返回允许的 CORS 跨域来源列表
func GetAllowedOrigins() []string {
	return appConfig.AllowedOrigins
}

// GetStaticDir 返回静态文件目录路径
// 如果未配置，返回空字符串（由调用方决定默认行为）
func GetStaticDir() string {
	return appConfig.StaticDir
}

// ============================================================
// DSN（数据源名称）构建
// ============================================================

// GetDSN 构建并返回主库的 MySQL 数据源名称 (Data Source Name)
// 格式：用户名:密码@tcp(地址:端口)/数据库名?参数
// 用于 GORM 或 database/sql 的连接
func GetDSN() string {
	return buildDSN(appConfig.MainDB)
}

// GetTestDSN 构建并返回测试库的 MySQL 数据源名称
// 格式与 GetDSN 相同，但使用测试库的配置参数
func GetTestDSN() string {
	return buildDSN(appConfig.TestDB)
}

// buildDSN 根据 DBConfig 构建标准的 MySQL DSN 连接字符串
// 参数说明：
//   - charset=utf8mb4  : 支持完整的 Unicode 字符（包括 emoji）
//   - parseTime=True   : 将 MySQL 的 DATETIME 自动解析为 Go 的 time.Time
//   - loc=Local        : 使用服务器的本地时区
func buildDSN(db DBConfig) string {
	return db.User + ":" + db.Pass +
		"@tcp(" + db.Host + ":" + db.Port + ")/" +
		db.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// ============================================================
// 内部辅助函数
// ============================================================

// getEnv 获取环境变量值，如果未设置则返回默认值
// 辅助函数，供内部使用
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseOrigins 解析逗号分隔的允许来源字符串为字符串数组
// 特殊处理：如果输入为 "*"，返回包含单个 "*" 的切片（表示允许所有来源）
func parseOrigins(originsStr string) []string {
	if originsStr == "*" {
		return []string{"*"}
	}
	origins := strings.Split(originsStr, ",")
	// 去除每个元素的首尾空格
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}
