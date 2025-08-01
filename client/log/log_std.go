//go:build !android

package log

import "github.com/zeromicro/go-zero/core/logx"

func logInfo(tag, msg string) {
	logx.WithCallerSkip(2).Infof("[%s] %s", tag, msg)
}
