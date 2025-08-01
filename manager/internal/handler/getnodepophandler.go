package handler

import (
	"context"
	"net/http"

	"titan-tunnel/manager/internal/logic"
	"titan-tunnel/manager/internal/svc"
	"titan-tunnel/manager/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func getNodePopHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetNodePopReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		ip := getClientIP(r)

		ctx := context.WithValue(r.Context(), "Remote-IP", ip)
		l := logic.NewGetNodePopLogic(ctx, svcCtx)
		resp, err := l.GetNodePop(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
