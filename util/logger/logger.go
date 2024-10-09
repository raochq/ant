package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

var (
	defaultHandler = NewLevelHandle(slog.Default().Handler(), slog.LevelInfo)
)

func init() {
	slog.SetDefault(slog.New(defaultHandler))
	// hook 错误日志
	Default().AddHookTextOut(os.Args[0]+".error", slog.LevelError)
}
func Default() *LevelHandler {
	return defaultHandler
}

type LevelHandler struct {
	handler slog.Handler
	level   slog.Level
	hooks   []slog.Handler
}

func NewLevelHandle(handler slog.Handler, level slog.Level, hooks ...slog.Handler) *LevelHandler {
	return &LevelHandler{
		handler: handler,
		level:   level,
		hooks:   hooks,
	}
}
func (h *LevelHandler) Enabled(_ context.Context, Level slog.Level) bool {
	return Level > h.level
}
func (h *LevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewLevelHandle(h.handler.WithAttrs(attrs), h.level, h.hooks...)
}

func (h *LevelHandler) WithGroup(name string) slog.Handler {
	return NewLevelHandle(h.handler.WithGroup(name), h.level, h.hooks...)
}

func (h *LevelHandler) Handle(ctx context.Context, r slog.Record) error {
	if value := ctx.Value("slog"); value != nil {
		if v, ok := value.([]slog.Attr); ok {
			r.AddAttrs(v...)
		}
	}
	for _, hook := range h.hooks {
		if hook.Enabled(ctx, r.Level) {
			hook.Handle(ctx, r)
		}
	}
	return h.handler.Handle(ctx, r)
}

func (h *LevelHandler) hook(handler slog.Handler) {
	h.hooks = append(h.hooks, handler)
}

func (h *LevelHandler) AddHookJSONOut(filename string, level slog.Level) {
	out := newLogOut(filename)
	if out == nil {
		return
	}
	hook := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})

	h.hook(hook)
}

func (h *LevelHandler) AddHookTextOut(filename string, level slog.Level) {
	out := newLogOut(filename)
	if out == nil {
		return
	}
	hook := slog.NewTextHandler(out, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})

	h.hook(hook)
}

func newLogOut(filename string) io.Writer {
	if filename == "" {
		return nil
	}
	dir := filepath.Dir(filename)
	if _, e := os.Stat(dir); e != nil {
		if os.IsNotExist(e) {
			if e := os.MkdirAll(dir, 0777); e != nil {
				slog.Error("create log dir failed", "error", e.Error())
				return nil
			}
		}
	}

	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	if ext == "" {
		ext = ".log"
	}
	out, err := rotatelogs.New(
		filename+".%y%m%d_%H"+ext,
		rotatelogs.WithLinkName(filename+ext), // 生成软链，指向最新日志文件
		// withLinkName(filename+ext),            // 生成软链，指向最新日志文件
		//rotatelogs.WithMaxAge(28*24*time.Hour), // 文件最大保存时间
		rotatelogs.WithMaxAge(-1),              // 保存文件个数
		rotatelogs.WithRotationCount(10),       // 保存文件个数
		rotatelogs.WithRotationTime(time.Hour), // 日志切割时间间隔
	)

	if err != nil {
		slog.Error("create log file failed", "error", err.Error(), "filename", filename)
		return nil
	}
	return out
}

func SetLogLevel(level slog.Level) {
	defaultHandler.level = level
}

func AddOutputFile(filename string, lv string) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(lv)); err == nil {
		if level < defaultHandler.level {
			defaultHandler.level = level
		}
		defaultHandler.AddHookJSONOut(filename, level)
	}
}
