package utils

import (
	"strings"
	"unicode"
)

// StringUtil 字符串工具类

// Trim 去除字符串两端的空白字符
func Trim(s string) string {
	return strings.TrimSpace(s)
}

// ToLower 将字符串转换为小写
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToUpper 将字符串转换为大写
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty 检查字符串是否不为空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// Contains 检查字符串是否包含指定的子串
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// StartsWith 检查字符串是否以指定的前缀开始
func StartsWith(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// EndsWith 检查字符串是否以指定的后缀结束
func EndsWith(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Split 分割字符串
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// Join 连接字符串
func Join(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// Replace 替换字符串中的子串
func Replace(s, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

// ReplaceAll 替换字符串中的所有子串
func ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// Substring 截取字符串
func Substring(s string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start > end {
		return ""
	}
	return s[start:end]
}

// Length 获取字符串的长度
func Length(s string) int {
	return len(s)
}

// RuneLength 获取字符串的字符长度（考虑Unicode）
func RuneLength(s string) int {
	return len([]rune(s))
}

// IsDigit 检查字符串是否只包含数字
func IsDigit(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsLetter 检查字符串是否只包含字母
func IsLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// IsAlphanumeric 检查字符串是否只包含字母和数字
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
