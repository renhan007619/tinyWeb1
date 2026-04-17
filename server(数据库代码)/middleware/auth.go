// Package middleware 提供 HTTP 中间件
// =============================================
// 作用：
//   AuthMiddleware 是 JWT 认证中间件，用于保护需要登录才能访问的 API 接口。
//   工作流程：
//   1. 从请求头 Authorization 中提取 token
//   2. 验证 token 是否有效（签名、过期时间）
//   3. 如果有效，将用户信息存入 context，传递给下一个处理器
//   4. 如果无效，返回 401 未授权错误
//
// 使用方式：
//   http.HandleFunc("/api/todos", middleware.AuthMiddleware(todoHandler))
// =============================================

package middleware

import (
	"context"
	"net/http"
	"strings"

	"tinyweb1/utils"

	"github.com/golang-jwt/jwt/v5"
)

// ContextKey 用于在 context 中存储用户信息的 key 类型
type ContextKey string

const (
	// UserIDKey context 中存储用户 ID 的 key
	UserIDKey ContextKey = "user_id"
	// UsernameKey context 中存储用户名的 key
	UsernameKey ContextKey = "username"
	// RoleKey context 中存储用户角色的 key
	RoleKey ContextKey = "role"
)

// AuthMiddleware JWT 认证中间件
// 保护需要登录的接口，验证通过后将用户信息注入 request context
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 从请求头获取 Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"code":401,"message":"缺少认证token"}`, http.StatusUnauthorized)
			return
		}

		// 2. 提取 Bearer token（格式："Bearer <token>"）
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"code":401,"message":"认证格式错误"}`, http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// 3. 验证 token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			// 区分不同错误类型
			if err == jwt.ErrTokenExpired {
				http.Error(w, `{"code":401,"message":"token已过期，请重新登录"}`, http.StatusUnauthorized)
				return
			}
			http.Error(w, `{"code":401,"message":"无效的认证token"}`, http.StatusUnauthorized)
			return
		}

		// 4. 将用户信息注入 context，传递给后续处理器
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		// 5. 调用下一个处理器（使用新的 context）
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserID 从 context 中提取用户ID
func GetUserID(ctx interface{ Value(key any) any }) (uint, bool) {
	if v := ctx.Value(UserIDKey); v != nil {
		if id, ok := v.(uint); ok {
			return id, true
		}
	}
	return 0, false
}

// GetUsername 从 context 中提取用户名
func GetUsername(ctx interface{ Value(key any) any }) (string, bool) {
	if v := ctx.Value(UsernameKey); v != nil {
		if name, ok := v.(string); ok {
			return name, true
		}
	}
	return "", false
}

// GetRole 从 context 中提取用户角色
func GetRole(ctx interface{ Value(key any) any }) (string, bool) {
	if v := ctx.Value(RoleKey); v != nil {
		if role, ok := v.(string); ok {
			return role, true
		}
	}
	return "", false
}
