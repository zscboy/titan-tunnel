//go:build android

package log

/*
#cgo LDFLAGS: -llog
#include <android/log.h>
*/
import "C"
import "unsafe"

func logInfo(tag, msg string) {
	cTag := C.CString(tag)
	cMsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cTag))
	defer C.free(unsafe.Pointer(cMsg))
	C.__android_log_print(C.ANDROID_LOG_INFO, cTag, C.CString("%s"), cMsg)
}
