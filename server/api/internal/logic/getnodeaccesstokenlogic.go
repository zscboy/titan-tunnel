package logic

import (
	"context"
	"time"

	"titan-tunnel/server/api/internal/svc"
	"titan-tunnel/server/api/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetNodeAccessTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNodeAccessTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodeAccessTokenLogic {
	return &GetNodeAccessTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNodeAccessTokenLogic) GetNodeAccessToken(req *types.AccessTokenReq) (resp *types.AccessTokenResp, err error) {
	// todo: add your logic here and delete this line

	return l.generateJwtToken(l.svcCtx.Config.JwtAuth.AccessSecret, l.svcCtx.Config.JwtAuth.AccessExpire, req.NodeId)
}

func (l *GetNodeAccessTokenLogic) generateJwtToken(secret string, expire int64, nodeId string) (*types.AccessTokenResp, error) {
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

	return &types.AccessTokenResp{Token: tokenStr}, nil
}
