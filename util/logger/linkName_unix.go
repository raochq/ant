//go:build !windows
// +build !windows

package logger

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

func withLinkName(linkName string) rotatelogs.Option {
	return rotatelogs.WithLinkName(linkName)
}
