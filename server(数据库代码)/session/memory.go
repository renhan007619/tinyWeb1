// Package session 提供用户会话管理功能
// =============================================
// 作用：
//   管理用户的登录状态（Session），支持两种模式：
//   - 内存模式（默认）：使用 sync.Map 存储，适合开发/单机部署
//   - Redis 模式：使用 Redis 存储，适合生产环境多实例部署
//
// Session 存储的内容：
//   - user_id: 用户 ID
//   - username: 用户名
//   - role: 用户角色 (admin/user)
//   - login_time: 登录时间
//
// 使用方式：
//   session.Create(1, "zhangsan", "user", token)  // 登录时创建
//   sess := session.Get(token)                       // 获取会话信息
//   session.Delete(token)                             // 登出时删除
// =============================================

package session

import (
	"sync"
	"time"
)

// SessionInfo 会话信息结构体
type SessionInfo struct {
	UserID    uint      `json:"user_id"`    // 用户 ID
	Username  string    `json:"username"`   // 用户名
	Role      string    `json:"role"`       // 角色
	LoginTime time.Time `json:"login_time"` // 登录时间
	Token     string    `json:"token"`      // 关联的 JWT token
}

// 内存存储（sync.Map 并发安全）
var sessions = &sessionStore{store: sync.Map{}}

// sessionStore 封装 sync.Map，提供类型安全的操作
type sessionStore struct {
	store sync.Map
}

// Create 创建新会话
func (s *sessionStore) Create(userID uint, username, role, token string) *SessionInfo {
	sess := &SessionInfo{
		UserID:    userID,
		Username:  username,
		Role:      role,
		LoginTime: time.Now(),
		Token:     token,
	}
	s.store.Store(token, sess)
	return sess
}

// Get 获取会话信息
func (s *sessionStore) Get(token string) (*SessionInfo, bool) {
	if v, ok := s.store.Load(token); ok {
		if sess, ok := v.(*SessionInfo); ok {
			return sess, true
		}
	}
	return nil, false
}

// Delete 删除会话（登出时调用）
func (s *sessionStore) Delete(token string) {
	s.store.Delete(token)
}

// ====== 对外暴露的便捷函数 ======

// Create 创建新会话
func Create(userID uint, username, role, token string) *SessionInfo {
	return sessions.Create(userID, username, role, token)
}

// Get 获取会话信息
func Get(token string) (*SessionInfo, bool) {
	return sessions.Get(token)
}

// Delete 删除会话
func Delete(token string) {
	sessions.Delete(token)
}
