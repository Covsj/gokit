package iutil

import (
	"math/rand"
	"strconv"
	"time"
)

var GlobalRandom = rand.New(rand.NewSource(time.Now().Unix()))

func GenerateRandomStr(length int, charset string) string {
	if charset == "" {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}
	b := make([]byte, length)
	for i := range b {
		randIndex := GlobalRandom.Intn(len(charset))
		b[i] = charset[randIndex]
	}
	return string(b)
}

func EduMailId() string {
	// 生成一个大的随机整数并转为 36 进制字符串
	s := strconv.FormatInt(GlobalRandom.Int63(), 36)
	// 等价于 substring(8)，长度不足时返回空串
	if len(s) <= 8 {
		return ""
	}
	return s[8:]
}
