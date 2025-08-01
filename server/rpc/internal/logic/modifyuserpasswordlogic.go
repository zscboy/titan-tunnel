package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"titan-tunnel/server/api/model"
	"titan-tunnel/server/rpc/internal/svc"
	"titan-tunnel/server/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyUserPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewModifyUserPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserPasswordLogic {
	return &ModifyUserPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ModifyUserPasswordLogic) ModifyUserPassword(in *pb.ModifyUserPasswordReq) (*pb.UserOperationResp, error) {
	user, err := model.GetUser(l.svcCtx.Redis, in.UserName)
	if err != nil {
		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	if user == nil {
		return &pb.UserOperationResp{ErrMsg: fmt.Sprintf("user %s not exist", in.UserName)}, nil
	}

	hash := md5.Sum([]byte(in.NewPassword))
	passwordMD5 := hex.EncodeToString(hash[:])

	user.PasswordMD5 = passwordMD5

	err = model.SaveUser(l.svcCtx.Redis, user)
	if err != nil {
		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	deleteUserCacheLogic := NewDeleteUserCache(l.ctx, l.svcCtx)
	if err := deleteUserCacheLogic.DeleteUserCache(in.UserName); err != nil {

		return &pb.UserOperationResp{ErrMsg: err.Error()}, nil
	}

	return &pb.UserOperationResp{Success: true}, nil
}
