package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	//"github.com/lestrrat/go-file-rotatelogs"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/mattn/go-colorable"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var (
	defaultLogger   *logrus.Logger
	captureLineInfo = false
)

func init() {
	defaultLogger = New(logrus.DebugLevel, true)
	hookError(defaultLogger, os.Args[0]+".error")
}
func lineInfo() string {
	_, file, line, ok := runtime.Caller(3)
	if ok {
		idx := 0
		if idx = strings.LastIndexByte(file, '/'); idx > 0 {
			idx = strings.LastIndexByte(file[:idx], '/')
		}
		if idx > 0 {
			return fmt.Sprintf("%s:%d", file[idx+1:], line)
		} else {
			return fmt.Sprintf("%s:%d", filepath.Base(file), line)
		}
	}
	return ""
}
func defaultEntry() logrus.FieldLogger {
	if captureLineInfo {
		return defaultLogger.WithField("line", lineInfo())
	}
	return defaultLogger
}

func WithField(key string, value interface{}) *logrus.Entry {
	return defaultEntry().WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return defaultEntry().WithFields(fields)
}

func WithError(err error) *logrus.Entry {
	return defaultEntry().WithError(err)
}

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

func New(level logrus.Level, captureLine bool) *logrus.Logger {
	l := logrus.New()
	captureLineInfo = captureLine
	l.SetLevel(level)
	l.SetOutput(colorable.NewColorableStdout())
	l.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.TimeOnly,
	})

	return l
}

func hookError(l *logrus.Logger, filename string) {
	if filename == "" {
		return
	}
	out, _ := rotatelogs.New(filename)
	errHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.ErrorLevel: out,
		logrus.FatalLevel: out,
		logrus.PanicLevel: out,
	}, &logrus.JSONFormatter{
		TimestampFormat: "01-02 15:04:05",
	})
	l.AddHook(errHook)
}

func setOutputFile(l *logrus.Logger, filename string) {
	if filename == "" {
		return
	}
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)

	if !filepath.IsAbs(filename) {
		filename, _ = filepath.Abs(filename)
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
		withLinkName(filename+ext), // 生成软链，指向最新日志文件
		//rotatelogs.WithMaxAge(28*24*time.Hour), // 文件最大保存时间
		rotatelogs.WithMaxAge(-1),              // 保存文件个数
		rotatelogs.WithRotationCount(10),       // 保存文件个数
		rotatelogs.WithRotationTime(time.Hour), // 日志切割时间间隔
	)
	logrus.Errorf("config local file system logger error. %+v", errors.WithStack(err))

	l.AddHook(lfshook.NewHook(out, &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.TimeOnly,
	}))
}

func SetLogLevel(lv int32) {
	level, err := logrus.ParseLevel(logrus.Level(lv).String())
	if err == nil {
		defaultLogger.SetLevel(level)
	}
}
func SetOutputFile(filename string) {
	setOutputFile(defaultLogger, filename)
}

// Debug log debug protocol with cyan color.
func Debug(v ...interface{}) {
	defaultEntry().Debug(v...)
}
func Debugf(format string, v ...interface{}) {
	defaultEntry().Debugf(format, v...)
}

// Info log normal protocol.
func Info(v ...interface{}) {
	defaultEntry().Info(v...)
}
func Infof(format string, v ...interface{}) {
	defaultEntry().Infof(format, v...)
}

// Warn log error protocol
func Warn(v ...interface{}) {
	defaultEntry().Warn(v...)
}
func Warnf(format string, v ...interface{}) {
	defaultEntry().Warnf(format, v...)
}

// Error log error protocol with red color.
func Error(v ...interface{}) {
	defaultEntry().Error(v...)
}
func Errorf(format string, v ...interface{}) {
	defaultEntry().Errorf(format, v...)
}

// Fatal log error protocol
func Fatal(v ...interface{}) {
	defaultEntry().Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	defaultEntry().Fatalf(format, v...)
}

// Panic log error protocol
func Panic(v ...interface{}) {
	defaultEntry().Panic(v...)
}
func Panicf(format string, v ...interface{}) {
	defaultEntry().Panicf(format, v...)
}
