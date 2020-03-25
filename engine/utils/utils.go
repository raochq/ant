package utils

import (
	"github.com/raochq/ant/engine/logger"
	"os"
	"path/filepath"
	"runtime/debug"
)

// 获取程序路径
func GetAppPath() string {
	fp, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "."
	}
	return filepath.Dir(fp)

}

// 判断文件或文件是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 打印panic堆栈
func PrintPanicStack() {
	if err := recover(); err != nil {
		buf := debug.Stack()
		logger.Error("panic: %v\n%s", err, buf)
	}
}

func IsInSlice(value interface{}, arr ...interface{}) bool {
	for _, val := range arr {
		if val == value {
			return true
		}
	}
	return false
}
