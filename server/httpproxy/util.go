package httpproxy

import (
	"net"
	"net/http"
	"strings"
)

func getRemoteIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Real-IP")
	if len(ip) != 0 {
		return ip, nil
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			return ip, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	return ip, nil
}
