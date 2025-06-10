package client

import (
	"math/rand"
	"strings"
	"time"
)

// 生成随机字符串
func RandomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// URL 路径拼接
func UrlJoin(baseURL string, paths ...string) string {
	// 拼接路径部分
	for _, p := range paths {
		if !strings.HasPrefix(baseURL, "/") && !strings.HasPrefix(p, "/") {
			baseURL += "/"
		}
		baseURL += p
	}
	return baseURL
}
