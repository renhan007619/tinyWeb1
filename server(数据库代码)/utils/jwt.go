// Package utils 提供 JWT token 生成和验证工具
// =============================================
// 作用：
//   使用 golang-jwt/jwt/v5 库进行 JWT（JSON Web Token）的生成和验证。
//   JWT 是一种无状态的认证机制，用户登录后获得 token，
//   之后每次请求都携带此 token 来证明身份。
//
// 双Token机制（2026-06-05更新）：
//   - Access Token: 2小时有效期，用于日常API请求
//   - Refresh Token: 7天有效期，用于刷新Access Token
//   优点：Access Token过期时间短更安全，Refresh Token保证7天免登录
//
// JWT 结构（三部分用 . 分隔）：
//   Header.Payload.Signature
//   - Header: 算法和类型 {"alg":"HS256","typ":"JWT"}
//   - Payload: 数据 {"user_id":1,"username":"zhangsan","exp":1234567890}
//   - Signature: 签名，防篡改
//
// 使用方式：
//   tokens, _ := utils.GenerateTokenPair(1, "zhangsan", "user")  // 登录时生成双Token
//   claims, _ := utils.ValidateAccessToken(token)                // 验证Access Token
//   claims, _ := utils.ValidateRefreshToken(token)               // 验证Refresh Token
// =============================================

package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JwtSecretKey 是签名密钥，从环境变量读取
// 注意：这个密钥必须保密，泄露后攻击者可以伪造任意 token
// 开发环境默认使用 dev-secret，生产环境必须设置 JWT_SECRET 环境变量
var JwtSecretKey []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// 开发环境默认值，生产环境必须设置环境变量
		secret = "dev-secret-key-do-not-use-in-production"
	}
	JwtSecretKey = []byte(secret)
}

// TokenPair 双Token结构
type TokenPair struct {
	AccessToken  string `json:"access_token"`  // 访问令牌（2小时）
	RefreshToken string `json:"refresh_token"` // 刷新令牌（7天）
	ExpiresIn    int64  `json:"expires_in"`    // Access Token有效期（秒）
}

// CustomClaims 自定义 JWT 载荷（Payload）
// 存储我们需要从 token 中提取的用户信息
type CustomClaims struct {
	UserID               uint   `json:"user_id"`    // 用户 ID
	Username             string `json:"username"`   // 用户名
	Role                 string `json:"role"`       // 用户角色：admin / user
	TokenType            string `json:"token_type"` // token类型：access/refresh
	jwt.RegisteredClaims        // JWT 标准声明（包含 exp、iat 等）
}

// GenerateTokenPair 生成双Token（Access Token + Refresh Token）
// 参数：userID 用户ID, username 用户名, role 角色
// 返回：TokenPair结构体, 错误信息
//
// Access Token: 2小时有效期，用于日常API请求
// Refresh Token: 7天有效期，用于刷新获取新的Access Token
func GenerateTokenPair(userID uint, username, role string) (*TokenPair, error) {
	now := time.Now()

	// 1. 生成 Access Token（2小时有效期）
	accessClaims := CustomClaims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "tinyweb1",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(JwtSecretKey)
	if err != nil {
		return nil, err
	}

	// 2. 生成 Refresh Token（7天有效期）
	refreshClaims := CustomClaims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)), // 7天
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "tinyweb1",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(JwtSecretKey)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    7200, // 2小时 = 7200秒
	}, nil
}

// ValidateAccessToken 验证 Access Token
// 用于验证日常API请求的token
func ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	claims, err := validateToken(tokenString)
	if err != nil {
		return nil, err
	}
	// 检查token类型
	if claims.TokenType != "access" {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}

// ValidateRefreshToken 验证 Refresh Token
// 用于验证刷新请求的token
func ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
	claims, err := validateToken(tokenString)
	if err != nil {
		return nil, err
	}
	// 检查token类型
	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}

// validateToken 内部通用的token验证函数
func validateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return JwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// =============================================
// 兼容旧版本的单Token接口（已废弃，保留用于兼容）
// =============================================

// GenerateToken 生成单JWT token（兼容旧版本）
// 已废弃，建议使用 GenerateTokenPair
func GenerateToken(userID uint, username, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tinyweb1",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecretKey)
}

// ValidateToken 验证单JWT token（兼容旧版本）
// 已废弃，建议使用 ValidateAccessToken
func ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return JwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
