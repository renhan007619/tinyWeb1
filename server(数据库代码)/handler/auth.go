// Package handler 用户认证相关接口
// =============================================
// 作用：
//   处理用户注册和登录的 HTTP 请求
//   - POST /api/auth/register : 用户注册
//   - POST /api/auth/login    : 用户登录（Day2 实现）
//   - GET  /api/auth/me       : 获取当前用户信息（Day2 实现）
//
// 安全设计：
//   - 密码使用 bcrypt 加密存储，永远不存明文
//   - 用户名唯一索引，防止重复注册
//   - 密码长度最少6位，用户名最少3位
//   - 注册时默认角色为 user（普通用户）
//   - 管理员账号需要手动在数据库中修改 role 字段
// =============================================

package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"tinyweb1/db"
	"tinyweb1/model"
	"tinyweb1/utils"
)

// Register 用户注册接口
// POST /api/auth/register
//
// 请求体：
//
//	{
//	  "username": "zhangsan",
//	  "password": "123456"
//	}
//
// 成功响应：
//
//	{
//	  "code": 0,
//	  "message": "success",
//	  "data": {
//	    "id": 1,
//	    "username": "zhangsan",
//	    "role": "user"
//	  }
//	}
//
// 错误响应：
//
//	{"code": 400, "message": "用户名不能为空"}
//	{"code": 400, "message": "用户名至少3个字符"}
//	{"code": 400, "message": "密码至少6位"}
//	{"code": 409, "message": "用户名已存在"}
func Register(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求体
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "请求参数格式错误"))
		return
	}

	// 2. 参数校验
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "用户名不能为空"))
		return
	}
	if len(req.Username) < 3 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "用户名至少3个字符"))
		return
	}
	if len(req.Password) < 6 {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "密码至少6位"))
		return
	}

	// 3. 检查用户名是否已存在
	database := db.GetDB()
	var existingUser model.User
	result := database.Where("username = ?", req.Username).First(&existingUser)
	if result.Error == nil {
		// 找到了同名用户 → 用户名已被占用
		sendJSON(w, http.StatusConflict, model.ErrorResponse(409, "用户名已存在"))
		return
	}
	if result.Error != gorm.ErrRecordNotFound {
		// 不是"找不到记录"的错误，而是数据库查询出错了
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "服务器内部错误"))
		return
	}

	// 4. 对密码进行 bcrypt 加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "密码加密失败"))
		return
	}

	// 5. 创建用户记录
	user := model.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Role:         "user", // 默认普通用户
	}
	if err := database.Create(&user).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "创建用户失败"))
		return
	}

	// 6. 返回用户信息（不含密码）
	sendJSON(w, http.StatusCreated, model.SuccessResponse(model.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	}))
}
