// Package utils 提供密码加密和验证工具
// =============================================
// 作用：
//   使用 bcrypt 算法对用户密码进行哈希加密和验证。
//   bcrypt 是目前最推荐的密码哈希算法，优点：
//   - 自带盐值（salt），同样的密码每次加密结果不同
//   - 可调成本因子（cost），抵抗暴力破解
//   - 慢哈希算法，专门为密码设计
//
// 为什么不用 MD5/SHA256？
//   MD5/SHA256 是快速哈希，攻击者可以每秒尝试几十亿次
//   bcrypt 每次验证约 100ms，暴力破解成本极高
//
// 使用方式：
//   hash, _ := utils.HashPassword("mypassword")     // 注册时加密
//   ok := utils.CheckPassword("mypassword", hash)   // 登录时验证
// =============================================

package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对明文密码进行 bcrypt 加密
// 参数：password 明文密码
// 返回：hash 加密后的哈希字符串（60字符），err 错误信息
//
// bcrypt.DefaultCost = 10，表示 2^10 = 1024 轮迭代
// 每次加密约 100ms，安全性和性能的平衡点
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证明文密码是否匹配哈希值
// 参数：password 用户输入的明文密码，hash 数据库中存储的哈希值
// 返回：true 密码正确，false 密码错误
//
// 工作原理：
//   bcrypt 从 hash 中提取盐值和成本因子，
//   用同样的参数对输入密码进行哈希，
//   然后对比两个哈希值是否一致
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
