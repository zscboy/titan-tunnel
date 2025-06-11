package logic

import (
	"context"
	"net/http"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/api/ws"

	"github.com/zeromicro/go-zero/core/logx"
)

type VmWSLogic struct {
	logx.Logger
	svcCtx *svc.ServiceContext
}

func NewVmWSLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VmWSLogic {
	return &VmWSLogic{
		Logger: logx.WithContext(ctx),
		svcCtx: svcCtx,
	}
}

func (l *VmWSLogic) VmWS(w http.ResponseWriter, r *http.Request, req *types.VMWSRequest) error {
	vmws := ws.NewVMWS(l.svcCtx.TunMgr)
	err := vmws.ServeWS(w, r, req)
	if err != nil {
		logx.Errorf("VmWSLogic.VmWS %s", err.Error())
	}

	return err
}
