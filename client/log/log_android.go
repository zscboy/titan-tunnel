//go:build android

package log

/*
#cgo LDFLAGS: -llog
#include <android/log.h>
#include <stdlib.h>


static void android_log_info(const char* tag, const char* msg) {
    __android_log_print(ANDROID_LOG_INFO, tag, "%s", msg);
}
*/
import "C"
import "unsafe"

func logInfo(tag, msg string) {
	cTag := C.CString(tag)
	cMsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cTag))
	defer C.free(unsafe.Pointer(cMsg))
	C.android_log_info(cTag, cMsg)
}
