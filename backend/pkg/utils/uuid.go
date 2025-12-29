package utils

import "github.com/google/uuid"

// NewUUID 返回一个全局唯一的 UUID 字符串。
func NewUUID() string {
	return uuid.NewString()
}
