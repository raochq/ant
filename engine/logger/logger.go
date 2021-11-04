package logger

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	//"github.com/lestrrat/go-file-rotatelogs"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/mattn/go-colorable"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var (
	gLogger = New(logrus.DebugLevel, true)
)

type FireFn func(ent *logrus.Entry) error

func (f FireFn) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
}
func (f FireFn) Fire(ent *logrus.Entry) error {
	return f(ent)
}

func caller() string {
	_, file, line, ok := runtime.Caller(11)
	if ok {
		return fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}
	return ""
}
func New(level logrus.Level, fileFlag bool) *logrus.Logger {
	l := logrus.New()

	if fileFlag {
		l.AddHook(FireFn(func(ent *logrus.Entry) error {
			ent.Data["file"] = caller()
			return nil
		}))
	}
	l.SetLevel(level)
	l.SetOutput(colorable.NewColorableStdout())
	l.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
		//FullTimestamp:   true,
		//TimestampFormat: "15:04:05.000",
	})
	return l
}

func setOutputFile(l *logrus.Logger, filename string) {
	if filename == "" {
		return
	}
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)

	if !filepath.IsAbs(filename) {
		filename, _ = filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), filename))
	}
	dir := filepath.Dir(filename)
	if _, e := os.Stat(dir); e != nil {
		if os.IsNotExist(e) {
			if e := os.MkdirAll(dir, 0777); e != nil {
				logrus.Fatalf("create log dir failed %v", e)
			}
		}
	}
	out, err := rotatelogs.New(
		filename+".%Y%m%d_%H"+ext,
		rotatelogs.WithLinkName(filename+ext), // 生成软链，指向最新日志文件
		//rotatelogs.WithMaxAge(28*24*time.Hour), // 文件最大保存时间
		rotatelogs.WithMaxAge(-1),              // 保存文件个数
		rotatelogs.WithRotationCount(10),       // 保存文件个数
		rotatelogs.WithRotationTime(time.Hour), // 日志切割时间间隔
	)
	logrus.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	ext = ".error" + ext
	errOut, _ := rotatelogs.New(
		filename+".%Y%m%d"+ext,
		rotatelogs.WithLinkName(filename+ext), // 生成软链，指向最新日志文件
	)

	fileHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: out, // 为不同级别设置不同的输出目的
		logrus.InfoLevel:  out,
		logrus.WarnLevel:  out,
		logrus.ErrorLevel: out,
		logrus.FatalLevel: out,
		logrus.PanicLevel: out,
	}, &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05.0000",
	})

	l.AddHook(fileHook)
	errHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.ErrorLevel: errOut,
		logrus.FatalLevel: errOut,
		logrus.PanicLevel: errOut,
	}, &logrus.TextFormatter{
		FullTimestamp: true,
	})
	l.AddHook(errHook)
}

func SetLogLevel(level string) {
	lv, err := logrus.ParseLevel(level)
	if err == nil {
		gLogger.SetLevel(lv)
	}
}
func SetOutputFile(filename string) {
	setOutputFile(gLogger, filename)
}

// Debug log debug protocol with cyan color.
func Debug(format string, v ...interface{}) {
	gLogger.Debugf(format, v...)
}

// Info log normal protocol.
func Info(format string, v ...interface{}) {
	gLogger.Infof(format, v...)
}

// Warn log error protocol
func Warn(format string, v ...interface{}) {
	gLogger.Warnf(format, v...)
}

// Error log error protocol with red color.
func Error(format string, v ...interface{}) {
	gLogger.Errorf(format, v...)
}

// Fatal log error protocol
func Fatal(format string, v ...interface{}) {
	gLogger.Fatalf(format, v...)
}

// Panic log error protocol
func Panic(format string, v ...interface{}) {
	gLogger.Panicf(format, v...)
}
