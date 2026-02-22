package utils

import (
	"github.com/google/uuid"
)

// UUID 生成一个随机的UUID
func UUID() string {
	return uuid.New().String()
}

// UUIDV4 生成一个版本4的UUID
func UUIDV4() string {
	return uuid.New().String()
}

// UUIDFromString 从字符串解析UUID
func UUIDFromString(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// IsValidUUID 检查字符串是否是有效的UUID
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
