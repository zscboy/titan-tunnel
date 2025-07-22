package httpproxy

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type ProxyRequest struct {
	ID     string              `json:"id"`
	Method string              `json:"Method"`
	URL    string              `json:"URL"`
	Header map[string][]string `json:"Header"`
}

type TargetInfo struct {
	conn       net.Conn
	req        *ProxyRequest
	userName   string
	extraBytes []byte
}

type ProxyResponse struct {
	ID         string              `json:"ID"`
	StatusCode int32               `json:"StatusCode"`
	Status     string              `json:"Status"`
	ProtoMajor int32               `json:"ProtoMajor"`
	ProtoMinor int32               `json:"ProtoMinor"`
	Header     map[string][]string `json:"Header"`
}

func getContentLength(header map[string][]string) (int64, error) {
	values, ok := header["Content-Length"]
	if !ok || len(values) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(values[0], 10, 64)
}

func (r *ProxyResponse) rebuildHTTPHeaders() strings.Builder {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(" HTTP/%d.%d %d %s\r\n", r.ProtoMajor, r.ProtoMinor, r.StatusCode, r.Status))
	for k, v := range r.Header {
		for _, vv := range v {
			b.WriteString(fmt.Sprintf("%s: %s\r\n", k, vv))
		}
	}
	b.WriteString("\r\n")
	return b
}
