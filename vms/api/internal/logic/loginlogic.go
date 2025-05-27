package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"titan-vm/vms/api/internal/svc"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/model"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	account, err := model.GetAccount(l.svcCtx.Redis, req.UserName)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("user %s not exist", req.UserName)
	}

	hasher := md5.New()
	hasher.Write([]byte(req.Password))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	if hashString != account.PasswordMD5 {
		return nil, fmt.Errorf("password not match")
	}

	token, err := l.generateJwtToken(l.svcCtx.Config.JwtAuth.AccessSecret, l.svcCtx.Config.JwtAuth.AccessExpire, account.UserName)
	if err != nil {
		return nil, err
	}
	return &types.LoginResponse{Token: token}, nil
}

func (l *LoginLogic) generateJwtToken(secret string, expire int64, userName string) (string, error) {
	claims := jwt.MapClaims{
		"user": userName,
		"exp":  time.Now().Add(time.Second * time.Duration(expire)).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))

}
