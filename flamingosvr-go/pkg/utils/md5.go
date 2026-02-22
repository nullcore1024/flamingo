package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 计算字符串的MD5哈希值
func MD5(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

// MD5Bytes 计算字节数组的MD5哈希值
func MD5Bytes(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// MD5File 计算文件的MD5哈希值
// 注意：这个函数需要传入文件内容的字节数组
func MD5File(fileContent []byte) string {
	return MD5Bytes(fileContent)
}
