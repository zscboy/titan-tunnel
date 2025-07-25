package handler

import (
	"context"
	"net"
	"net/http"
	"strings"

	"titan-tunnel/manager/internal/logic"
	"titan-tunnel/manager/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func getAuthTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		ctx := context.WithValue(r.Context(), "Remote-IP", ip)
		l := logic.NewGetAuthTokenLogic(ctx, svcCtx)
		resp, err := l.GetAuthToken()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

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
