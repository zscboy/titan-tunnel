package handler

import (
	"net/http"

	"titan-tunnel/server/api/internal/logic"
	"titan-tunnel/server/api/internal/svc"
	"titan-tunnel/server/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func modifyUserPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ModifyUserPasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewModifyUserPasswordLogic(r.Context(), svcCtx)
		resp, err := l.ModifyUserPassword(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
