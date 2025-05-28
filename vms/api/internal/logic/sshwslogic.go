package logic

import (
	"context"
	"net/http"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/api/ws"

	"github.com/zeromicro/go-zero/core/logx"
)

type SshWSLogic struct {
	logx.Logger
	svcCtx *svc.ServiceContext
}

func NewSshWSLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SshWSLogic {
	return &SshWSLogic{
		Logger: logx.WithContext(ctx),
		svcCtx: svcCtx,
	}
}

func (l *SshWSLogic) SshWS(w http.ResponseWriter, r *http.Request, req *types.SSHWSReqeust) error {
	sshws := ws.NewSSHWS(l.svcCtx.TunMgr)
	err := sshws.ServeWS(w, r, req)
	if err != nil {
		logx.Error("ssh ws: ", err.Error())
	}
	return err
}
