package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"titan-tunnel/manager/internal/logic"
	"titan-tunnel/manager/internal/svc"
)

func getPopsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetPopsLogic(r.Context(), svcCtx)
		resp, err := l.GetPops()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
