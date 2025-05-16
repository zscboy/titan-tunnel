package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"titan-vm/vms/api/internal/logic"
	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
)

func createVolWithLibvirtHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateVolWithLibvirtReqeust
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewCreateVolWithLibvirtLogic(r.Context(), svcCtx)
		resp, err := l.CreateVolWithLibvirt(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
