package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// 获取程序名称
func GetAppName() string {
	path := filepath.Dir(os.Args[0])
	appname := strings.Trim(os.Args[0], path)
	appname = strings.Trim(appname, filepath.Ext(appname))

	return appname

}

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
		return errors.Is(err, fs.ErrExist)
	}
	return true
}

// 打印panic堆栈
func PrintPanicStack() {
	if err := recover(); err != nil {
		buf := debug.Stack()
		slog.Error(fmt.Sprintf("panic:%s", buf), "error", err)
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
