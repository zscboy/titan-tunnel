package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"titan-ipoverlay/ippop/rpc/internal/svc"
	"titan-ipoverlay/ippop/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNodeAccessTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetNodeAccessTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodeAccessTokenLogic {
	return &GetNodeAccessTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetNodeAccessTokenLogic) GetNodeAccessToken(in *pb.GetNodeAccessTokenReq) (*pb.GetNodeAccessTokenResp, error) {
	return l.getNodeAccessToken(in.NodeId)
}

func (l *GetNodeAccessTokenLogic) getNodeAccessToken(nodeId string) (*pb.GetNodeAccessTokenResp, error) {
	url := fmt.Sprintf("http://%s/node/access/token?nodeid=%s", l.svcCtx.Config.APIServer, nodeId)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		buf, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status code %d, error:%s", resp.StatusCode, string(buf))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	accessTokenResp := struct {
		Token string `json:"token"`
	}{}

	err = json.Unmarshal(body, &accessTokenResp)
	if err != nil {
		return nil, err
	}

	return &pb.GetNodeAccessTokenResp{Token: accessTokenResp.Token}, nil
}
