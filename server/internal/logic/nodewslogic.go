package logic

import (
	"context"
	"net/http"

	"titan-tunnel/server/internal/svc"
	"titan-tunnel/server/internal/types"
	"titan-tunnel/server/ws"

	"github.com/zeromicro/go-zero/core/logx"
)

type NodeWSLogic struct {
	logx.Logger
	svcCtx *svc.ServiceContext
}

func NewNodeWSLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NodeWSLogic {
	return &NodeWSLogic{
		Logger: logx.WithContext(ctx),
		svcCtx: svcCtx,
	}
}

func (l *NodeWSLogic) NodeWS(w http.ResponseWriter, r *http.Request, req *types.NodeWSReq) error {
	nodeWS := ws.NewNodeWS(l.svcCtx.TunMgr)
	return nodeWS.ServeWS(w, r, req)
}
