//go:build !android

package log

import "github.com/zeromicro/go-zero/core/logx"

func LogDebug(tag, msg string) {
	logx.Debugf("[%s] %s", tag, msg)
}
