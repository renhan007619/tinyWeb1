module tinyweb1

go 1.25.0

// GORM - Go 语言最流行的 ORM 库，用于简化数据库操作
// 提供自动建表、CRUD 操作、关联查询、事务管理等功能
require (
	github.com/go-sql-driver/mysql v1.8.1
	gorm.io/driver/mysql v1.5.7
	gorm.io/gorm v1.25.12
)

require golang.org/x/crypto v0.50.0

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.36.0 // indirect
)
