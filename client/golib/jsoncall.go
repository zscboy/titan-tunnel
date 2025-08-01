package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

// input
type JSONCallArgs struct {
	Method string `json:"method"`
	Params string `json:"params"`
}

// output
type JSONCallResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func jsonCallResult(result *JSONCallResult) *C.char {
	resultJson, err := json.Marshal(result)
	if err != nil {
		resultJson = []byte(fmt.Sprintf("marsal result error %v", err.Error()))
	}

	return C.CString(string(resultJson))
}

func freeCString(jsonStrPtr *C.char) {
	C.free(unsafe.Pointer(jsonStrPtr))
}

func jsonCall(jsonStrPtr *C.char) *C.char {
	args := JSONCallArgs{}
	jsonStr := C.GoString(jsonStrPtr)
	err := json.Unmarshal([]byte(jsonStr), &args)
	if err != nil {
		return jsonCallResult(&JSONCallResult{Code: -1, Msg: err.Error()})
	}

	result := &JSONCallResult{}
	switch args.Method {
	case "start":
		result = startTunnel(args.Params)
	case "stop":
		result = stopTunnel()
	default:
		result.Code = -1
		result.Msg = fmt.Sprintf("Method %s not found", args.Method)
	}

	return jsonCallResult(result)
}
