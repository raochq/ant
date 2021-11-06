package util

import (
	"crypto/sha1"
	"fmt"
	"io"
	"time"
)

// 计算字符串哈希值
func Sha1Hash(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

// 生产user的token
func CalculateUserToken(input string) (userToken string) {
	userToken = Sha1Hash(time.Now().String() + input)
	return userToken
}
