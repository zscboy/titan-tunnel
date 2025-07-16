package logic

import (
	"context"
	"fmt"
	"time"

	"titan-tunnel/manager/internal/svc"
	"titan-tunnel/manager/internal/types"

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
	// get node pop by node ip
	pop := l.getNodePop(req)
	if pop == nil {
		return nil, fmt.Errorf("not found pop for node %s", req.NodeId)
	}
	return &types.GetNodePopResp{ServerURL: pop.WSServerURL, AccessToken: token}, nil
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

func (l *GetNodePopLogic) getNodePop(req *types.GetNodePopReq) *svc.Server {
	for _, pop := range l.svcCtx.Servers {
		return pop
	}
	return nil
}
