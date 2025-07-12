package logic

import (
	"context"
	"time"

	"titan-tunnel/server/api/internal/svc"
	"titan-tunnel/server/api/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetNodePopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNodePopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodePopLogic {
	return &GetNodePopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNodePopLogic) GetNodePop(req *types.GetNodePopReq) (resp *types.GetNodePopResp, err error) {
	token, err := l.generateJwtToken(l.svcCtx.Config.JwtAuth.AccessSecret, l.svcCtx.Config.JwtAuth.AccessExpire, req.NodeId)
	if err != nil {
		return nil, err
	}
	return &types.GetNodePopResp{ServerURL: l.svcCtx.Config.ServerURL, AccessToken: token}, nil
}

func (l *GetNodePopLogic) generateJwtToken(secret string, expire int64, userName string) (string, error) {
	claims := jwt.MapClaims{
		"user": userName,
		"exp":  time.Now().Add(time.Second * time.Duration(expire)).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))

}
