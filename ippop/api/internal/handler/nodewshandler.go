package handler

import (
	"net/http"

	"titan-ipoverlay/ippop/api/internal/logic"
	"titan-ipoverlay/ippop/api/internal/svc"
	"titan-ipoverlay/ippop/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func nodeWSHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.NodeWSReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewNodeWSLogic(r.Context(), svcCtx)
		err := l.NodeWS(w, r, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		}
	}
}
