package httpproxy

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const (
	ServerRequest         = 1
	ClientResponseHeaders = 2
	ClientResponseError   = 3
)

type Message struct {
	Type    int         `json:"Type"`
	Payload interface{} `json:"Payload"`
}

type HTTPRequest struct {
	ID     string              `json:"ID"`
	Method string              `json:"Method"`
	URL    string              `json:"URL"`
	Header map[string][]string `json:"Header"`
	Body   []byte              `json:"Body"`
}

type HTTPResponseHeader struct {
	ID         string              `json:"ID"`
	StatusCode int32               `json:"StatusCode"`
	Status     string              `json:"Status"`
	Header     map[string][]string `json:"Header"`
}

type HTTPResponseError struct {
	ID    string `json:"ID"`
	Error string `json:"Error"`
}

type TargetInfo struct {
	conn     net.Conn
	req      *HTTPRequest
	userName string
	// extraBytes []byte
}

func getContentLength(header http.Header) (int64, error) {
	// values := header.Values("Content-Length") // 自动处理大小写
	// if len(values) == 0 {
	// 	return 0, nil
	// }
	// logx.Debugf("values:%v", values)
	// return strconv.ParseInt(values[0], 10, 64)
	targetKey := "content-length" // 目标key（小写）
	for key, values := range header {
		if strings.EqualFold(key, targetKey) { // 不区分大小写比较
			if len(values) == 0 {
				return 0, nil
			}
			return strconv.ParseInt(values[0], 10, 64)
		}
	}
	return 0, nil
}

func (r *HTTPResponseHeader) rebuildHTTPHeaders() strings.Builder {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("HTTP/%d.%d %d %s\r\n", 1, 1, r.StatusCode, r.Status))
	for k, v := range r.Header {
		for _, vv := range v {
			b.WriteString(fmt.Sprintf("%s: %s\r\n", k, vv))
		}
	}
	b.WriteString("\r\n")
	return b
}
