package handler

import (
	"net/http"

	"titan-vm/vms/api/internal/logic"
	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func getHostInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HostInfoRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetHostInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetHostInfo(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
