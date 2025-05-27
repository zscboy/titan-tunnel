package logic

import (
	"context"
	"net/http"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/api/ws"

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

func (l *NodeWSLogic) NodeWS(w http.ResponseWriter, r *http.Request, req *types.NodeWSRequest) error {
	nodeWS := ws.NewNodeWS(l.svcCtx.TunMgr)
	return nodeWS.ServeWS(w, r, req)
}
