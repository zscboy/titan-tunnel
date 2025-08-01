package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"titan-tunnel/server/api/internal/logic"
	"titan-tunnel/server/api/internal/svc"
	"titan-tunnel/server/api/internal/types"
)

func deleteUserCacheHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteUserCache
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewDeleteUserCacheLogic(r.Context(), svcCtx)
		err := l.DeleteUserCache(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
