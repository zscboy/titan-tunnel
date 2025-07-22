package logic

import (
	"context"
	"fmt"
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

	domain := l.svcCtx.Config.Socks5.ServerIP
	if len(l.svcCtx.Config.Domain) > 0 {
		domain = l.svcCtx.Config.Domain
	}

	wsServerURl := fmt.Sprintf("ws://%s:%d/ws/node", domain, l.svcCtx.Config.Port)

	return &types.GetNodePopResp{ServerURL: wsServerURl, AccessToken: string(token)}, nil
}

func (l *GetNodePopLogic) generateJwtToken(secret string, expire int64, nodeId string) ([]byte, error) {
	claims := jwt.MapClaims{
		"user": nodeId,
		"exp":  time.Now().Add(time.Second * time.Duration(expire)).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return []byte(tokenStr), nil
}
