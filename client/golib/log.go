package main

/*
#include <android/log.h>
*/
import "C"
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
	defer freeCString(cTag)
	defer freeCString(cMsg)
	defer C.free(unsafe.Pointer(cMsg))
	logToAndroid(C.ANDROID_LOG_DEBUG, cTag, cMsg)
}

func LogDebug(tag, msg string) {
	if runtime.GOOS == "android" {
		androidLogDebug(tag, msg)
	} else {
		logx.Debugf("[%s] %s", tag, msg)
	}
}
