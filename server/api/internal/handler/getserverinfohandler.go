package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"titan-tunnel/server/api/internal/logic"
	"titan-tunnel/server/api/internal/svc"
)

func getServerInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetServerInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetServerInfo()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
