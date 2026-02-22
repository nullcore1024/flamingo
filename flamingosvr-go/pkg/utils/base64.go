package utils

import (
	"encoding/base64"
)

// Base64Encode 对数据进行Base64编码
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode 对Base64编码的数据进行解码
func Base64Decode(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// Base64URLEncode 对数据进行URL安全的Base64编码
func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64URLDecode 对URL安全的Base64编码的数据进行解码
func Base64URLDecode(encoded string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(encoded)
}
