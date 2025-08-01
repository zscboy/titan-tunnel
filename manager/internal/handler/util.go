package handler

import (
	"net"
	"net/http"
	"strings"
)

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	return ip
}
