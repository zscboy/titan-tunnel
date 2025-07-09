package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"titan-tunnel/server/internal/svc"
	"titan-tunnel/server/internal/types"
	"titan-tunnel/server/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyUserPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewModifyUserPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserPasswordLogic {
	return &ModifyUserPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyUserPasswordLogic) ModifyUserPassword(req *types.ModifyUserPasswordReq) (resp *types.UserOperationResp, err error) {
	user, err := model.GetUser(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if user == nil {
		return &types.UserOperationResp{ErrMsg: fmt.Sprintf("user %s not exist", req.UserName)}, nil
	}

	hash := md5.Sum([]byte(req.NewPassword))
	passwordMD5 := hex.EncodeToString(hash[:])

	user.PasswordMD5 = passwordMD5

	err = model.SaveUser(l.svcCtx.Redis, user)
	if err != nil {
		return &types.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	l.svcCtx.TunMgr.DeleteUserFromCache(req.UserName)
	return &types.UserOperationResp{Success: true}, nil
}
