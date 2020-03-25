package logger

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
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

var (
	gLogger *logger
)

type logger struct {
	level                         LogLevel
	writer                        io.Writer
	debug, info, warn, err, fatal *log.Logger
}

func New(level LogLevel, writer io.Writer) *logger {
	var out, errOut io.Writer
	if writer != nil {
		out = writer
		out = writer
	} else {
		out = os.Stdout
		errOut = os.Stderr
	}
	l := &logger{
		level:  level,
		writer: writer,
		debug:  log.New(out, "\033[0;36mDEBUG:\033[0m ", log.LstdFlags|log.Lshortfile),
		info:   log.New(out, "\033[0;32mINFO:\033[0m ", log.LstdFlags|log.Lshortfile),
		warn:   log.New(errOut, "\033[0;35mWARN:\033[0m ", log.LstdFlags|log.Lshortfile),
		err:    log.New(errOut, "\033[0;31mERROR:\033[0m ", log.LstdFlags|log.Lshortfile),
		fatal:  log.New(errOut, "\033[0;33mFATAL:\033[0m ", log.LstdFlags|log.Lshortfile),
	}
	return l
}
func (l *logger) setOutPut(w io.Writer) {
	var out, errOut io.Writer
	if w != nil {
		out = w
		errOut = w
	} else {
		out = os.Stdout
		errOut = os.Stderr
	}
	l.debug.SetOutput(out)
	l.info.SetOutput(out)
	l.warn.SetOutput(errOut)
	l.err.SetOutput(errOut)
	l.fatal.SetOutput(errOut)
	l.writer = w
}

func SetLogLevel(level LogLevel) {
	Logger().level = level
}

func SetOutput(filename string) {
	l := Logger()
	if filename == "" {
		l.setOutPut(nil)
		if l.writer != nil {
			if val, ok := l.writer.(*lumberjack.Logger); ok {
				val.Close()
			}
		}
	} else {
		if l.writer == nil {
			l.setOutPut(&lumberjack.Logger{
				Filename:   filename,
				MaxSize:    500, // megabytes
				MaxBackups: 10,
				MaxAge:     28,   //days
				Compress:   true, // disabled by default
			})
		} else {
			if val, ok := l.writer.(*lumberjack.Logger); ok {
				val.Filename = filename
			}
		}
	}
}

func Logger() *logger {
	if gLogger == nil {
		l := New(LogLevel_Info, nil)
		gLogger = l
	}
	return gLogger
}

// Debug log debug protocol with cyan color.
func Debug(format string, v ...interface{}) {
	Logger().debug.Output(2, fmt.Sprintf(format, v...))
}

// Info log normal protocol.
func Info(format string, v ...interface{}) {
	Logger().info.Output(2, fmt.Sprintf(format, v...))
}

// Warn log error protocol
func Warn(format string, v ...interface{}) {
	Logger().warn.Output(2, fmt.Sprintf(format, v...))
}

// Error log error protocol with red color.
func Error(format string, v ...interface{}) {
	Logger().err.Output(2, fmt.Sprintf(format, v...))
}

// Fatal log error protocol
func Fatal(format string, v ...interface{}) {
	Logger().fatal.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}
