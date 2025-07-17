package main

import (
	"runtime"
	"unsafe"

	"github.com/zeromicro/go-zero/core/logx"
)

func logToAndroid(level C.int, tag, msg *C.char) {
	C.__android_log_write(level, tag, msg)
}

// 封装为Go函数
func androidLogDebug(tag, msg string) {
	cTag := C.CString(tag)
	cMsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cTag))
	defer C.free(unsafe.Pointer(cMsg))
	logToAndroid(C.ANDROID_LOG_DEBUG, cTag, cMsg)
}

func LogDebug(tag, msg string) {
	if runtime.GOOS == "android" {
		androidLogDebug(tag, msg)
	} else {
		logx.Debug(tag, msg)
	}
}
