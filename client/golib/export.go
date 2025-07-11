package main

/*
#cgo CFLAGS: -I/usr/lib/jvm/java-17-openjdk-amd64/include
#cgo CFLAGS: -I/usr/lib/jvm/java-17-openjdk-amd64/include/linux
#include <stdlib.h>
#include <jni.h>

static inline const char* jx_GetStringUTFChars(JNIEnv *env, jstring str) {
    return (*env)->GetStringUTFChars(env, str, NULL);
}

static inline void jx_ReleaseStringUTFChars(JNIEnv *env, jstring str, const char* utf8chars) {
    (*env)->ReleaseStringUTFChars(env, str, utf8chars);
}

static inline jstring jx_NewStringUTF(JNIEnv *env, const char* str) {
    return (*env)->NewStringUTF(env, str);
}

*/
import "C"

//export Java_com_titan_app_ipservice_JSONCall
func Java_com_titan_app_ipservice_JSONCall(env *C.JNIEnv, thiz C.jobject, args C.jstring) C.jstring {
	var utf8Chars *C.char
	var goChars *C.char
	var result C.jstring

	/* Get the UTF-8 characters that represent our java string */
	utf8Chars = C.jx_GetStringUTFChars(env, args)

	goChars = jsonCall(utf8Chars)
	C.jx_ReleaseStringUTFChars(env, args, utf8Chars)

	result = C.jx_NewStringUTF(env, goChars)
	freeCString(goChars)

	return result
}

func main() {

}
