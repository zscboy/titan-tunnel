package handler

import (
	"net/http"

	"titan-tunnel/server/api/internal/logic"
	"titan-tunnel/server/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
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
