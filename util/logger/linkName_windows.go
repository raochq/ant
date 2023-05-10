package logger

import (
	"os"
	"syscall"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
)

func withLinkName(linkName string) rotatelogs.Option {
	// window下的Symlink有问题，打个补丁
	return rotatelogs.WithHandler(rotatelogs.HandlerFunc(func(event rotatelogs.Event) {
		if ev, ok := event.(*rotatelogs.FileRotatedEvent); ok {
			os.Remove(linkName)
			err := os.Symlink(ev.CurrentFile(), linkName)
			if err != nil {
				var errNo syscall.Errno
				if errors.As(err, &errNo) {
					if errNo == syscall.ERROR_ALREADY_EXISTS {
						return
					}
				}
				os.Link(ev.CurrentFile(), linkName)
			}
		}
	}))
}
