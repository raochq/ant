package logger

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	"path/filepath"
)

type LogLevel int8

const (
	LogLevel_None LogLevel = iota
	LogLevel_Fatal
	LogLevel_Error
	LogLevel_Warn
	LogLevel_Info
	LogLevel_Debug
)

type PrefixStyle int8

const (
	PrefixStyle_None PrefixStyle = iota
	PrefixStyle_Normal
	PrefixStyle_Color
)

var (
	gLogger      = New(LogLevel_Info, os.Stdout, PrefixStyle_Normal, log.LstdFlags|log.Lshortfile)
	colorPrefixs = []string{
		"",
		"\033[0;33mFATAL:\033[0m ",
		"\033[0;31mERROR:\033[0m ",
		"\033[0;35mWARN:\033[0m ",
		"\033[0;32mINFO:\033[0m ",
		"\033[0;36mDEBUG:\033[0m ",
	}
	normalPrefixs = []string{
		"",
		"FATAL: ",
		"ERROR: ",
		"WARN: ",
		"INFO: ",
		"DEBUG: ",
	}
)

type logger struct {
	level                         LogLevel
	out                           io.Writer
	debug, info, warn, err, fatal *log.Logger
}

func New(level LogLevel, out io.Writer, style PrefixStyle, flag int) *logger {
	var prefix []string
	switch style {
	case PrefixStyle_Normal:
		prefix = normalPrefixs
	case PrefixStyle_Color:
		prefix = colorPrefixs
	default:
		prefix = make([]string, 6)
	}

	l := &logger{
		level: level,
		out:   out,
		debug: log.New(out, prefix[LogLevel_Debug], flag),
		info:  log.New(out, prefix[LogLevel_Info], flag),
		warn:  log.New(out, prefix[LogLevel_Warn], flag),
		err:   log.New(out, prefix[LogLevel_Error], flag),
		fatal: log.New(out, prefix[LogLevel_Fatal], flag),
	}
	return l
}
func (l *logger) SetLogLevel(level string) {
	lv := LogLevel_Info
	switch level {
	case "debug":
		lv = LogLevel_Debug
	case "info":
		lv = LogLevel_Info
	case "warn":
		lv = LogLevel_Warn
	case "error":
		lv = LogLevel_Error
	case "fatal":
		lv = LogLevel_Fatal
	default:
		lv = LogLevel_Info
	}
	l.level = lv
}

func SetLogLevel(level string) {
	gLogger.SetLogLevel(level)
}

func (l *logger) SetPrefixStyle(style PrefixStyle) {
	var prefix []string
	switch style {
	case PrefixStyle_Normal:
		prefix = normalPrefixs
	case PrefixStyle_Color:
		prefix = colorPrefixs
	default:
		prefix = make([]string, 6)
	}
	l.debug.SetPrefix(prefix[LogLevel_Debug])
	l.info.SetPrefix(prefix[LogLevel_Info])
	l.warn.SetPrefix(prefix[LogLevel_Warn])
	l.err.SetPrefix(prefix[LogLevel_Error])
	l.fatal.SetPrefix(prefix[LogLevel_Fatal])
}
func SetPrefixStyle(style PrefixStyle) {
	gLogger.SetPrefixStyle(style)
}

func (l *logger) SetOutPut(out io.Writer) {
	l.out = out
	l.debug.SetOutput(out)
	l.info.SetOutput(out)
	l.warn.SetOutput(out)
	l.err.SetOutput(out)
	l.fatal.SetOutput(out)
}
func SetOutput(out io.Writer) {
	gLogger.SetOutPut(out)
}
func SetOutputFile(filename string) {
	if !filepath.IsAbs(filename) {
		filename, _ = filepath.Abs(filepath.Dir(os.Args[0]) + "/" + filename)
	}
	l := gLogger
	if filename == "" {
		if val, ok := l.out.(*lumberjack.Logger); ok {
			val.Close()
		}
		l.SetOutPut(os.Stdout)
	} else {
		if val, ok := l.out.(*lumberjack.Logger); ok {
			val.Filename = filename
		} else {
			out := &lumberjack.Logger{
				Filename:   filename,
				MaxSize:    10, // megabytes
				MaxBackups: 10,
				MaxAge:     28, //days
			}
			l.SetOutPut(out)
		}
	}
}

// Debug log debug protocol with cyan color.
func Debug(format string, v ...interface{}) {
	gLogger.debug.Output(2, fmt.Sprintf(format, v...))
}

// Info log normal protocol.
func Info(format string, v ...interface{}) {
	gLogger.info.Output(2, fmt.Sprintf(format, v...))
}

// Warn log error protocol
func Warn(format string, v ...interface{}) {
	gLogger.warn.Output(2, fmt.Sprintf(format, v...))
}

// Error log error protocol with red color.
func Error(format string, v ...interface{}) {
	gLogger.err.Output(2, fmt.Sprintf(format, v...))
}

// Fatal log error protocol
func Fatal(format string, v ...interface{}) {
	gLogger.fatal.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}
