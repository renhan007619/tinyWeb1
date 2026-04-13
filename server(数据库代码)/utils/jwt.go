// Package utils 提供 JWT token 生成和验证工具
// =============================================
// 作用：
//   使用 golang-jwt/jwt/v5 库进行 JWT（JSON Web Token）的生成和验证。
//   JWT 是一种无状态的认证机制，用户登录后获得 token，
//   之后每次请求都携带此 token 来证明身份。
//
// JWT 结构（三部分用 . 分隔）：
//   Header.Payload.Signature
//   - Header: 算法和类型 {"alg":"HS256","typ":"JWT"}
//   - Payload: 数据 {"user_id":1,"username":"zhangsan","exp":1234567890}
//   - Signature: 签名，防篡改
//
// 使用方式：
//   token, _ := utils.GenerateToken(1, "zhangsan", "user")  // 登录时生成
//   claims, _ := utils.ValidateToken(token)                   // 验证时解析
// =============================================

package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JwtSecretKey 是签名密钥，生产环境应该从环境变量读取
// 注意：这个密钥必须保密，泄露后攻击者可以伪造任意 token
var JwtSecretKey = []byte("tinyweb1-secret-key-2026")

// CustomClaims 自定义 JWT 载荷（Payload）
// 存储我们需要从 token 中提取的用户信息
type CustomClaims struct {
	UserID   uint   `json:"user_id"`   // 用户 ID
	Username string `json:"username"`  // 用户名
	Role     string `json:"role"`      // 用户角色：admin / user
	jwt.RegisteredClaims               // JWT 标准声明（包含 exp、iat 等）
}

// GenerateToken 生成 JWT token
// 参数：userID 用户ID, username 用户名, role 角色
// 返回：token字符串, 错误信息
//
// token 有效期设置为 24 小时（生产环境可配置）
func GenerateToken(userID uint, username, role string) (string, error) {
	// 设置 token 过期时间为当前时间 + 24小时
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建自定义载荷
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			// 签发时间（何时生成）
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// 生效时间（立即生效）
			NotBefore: jwt.NewNumericDate(time.Now()),
			// 签发者（可选，标识谁签发的）
			Issuer: "tinyweb1",
		},
	}

	// 使用 HS256 算法创建 token 对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 用密钥签名并返回完整 token 字符串
	tokenString, err := token.SignedString(JwtSecretKey)
	return tokenString, err
}

// ValidateToken 验证并解析 JWT token
// 参数：tokenString 客户端传来的 token
// 返回：自定义载荷（包含用户信息）, 错误信息
//
// 验证流程：
//   1. 检查签名是否正确（防止篡改）
//   2. 检查是否过期
//   3. 解析出用户信息
func ValidateToken(tokenString string) (*CustomClaims, error) {
	// 解析 token，同时验证签名
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法是否为 HS256（防止算法混淆攻击）
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// 返回签名密钥用于验证
		return JwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// 类型断言，提取自定义载荷
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
