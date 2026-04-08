// Package handler 共享工具函数
// =============================================
// 作用：
//   提供 handler 包内通用的辅助函数，包括 JSON 响应发送、字符串处理等。
//   这些函数被多个 handler 文件共用，避免重复定义。
// =============================================

package handler

import (
	"encoding/json"
	"net/http"
)

// sendJSON 发送 JSON 响应的辅助函数
// 将任意数据序列化为 JSON 并写入 http.ResponseWriter
// 同时设置 Content-Type: application/json 头部
func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// trimString 去除字符串首尾空格的辅助函数
func trimString(s string) string {
	result := make([]byte, len(s))
	i := 0
	started := false
	for j := 0; j < len(s); j++ {
		c := s[j]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if started {
				result[i] = c
				i++
			}
		} else {
			started = true
			result[i] = c
			i++
		}
	}
	for i > 0 && (result[i-1] == ' ' || result[i-1] == '\t' || result[i-1] == '\n' || result[i-1] == '\r') {
		i--
	}
	return string(result[:i])
}

// parseIntQueryParam 从 URL Query 中解析整数参数
// 如果参数不存在或解析失败，返回默认值 defaultValue
func parseIntQueryParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	result := 0
	isNegative := false
	for i, c := range value {
		if i == 0 && c == '-' {
			isNegative = true
			continue
		}
		if c < '0' || c > '9' {
			return defaultValue
		}
		result = result*10 + int(c-'0')
	}
	if isNegative {
		result = -result
	}
	return result
}
