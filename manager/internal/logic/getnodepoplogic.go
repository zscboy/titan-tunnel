package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"titan-ipoverlay/ippop/rpc/serverapi"
	"titan-ipoverlay/manager/internal/config"
	"titan-ipoverlay/manager/internal/svc"
	"titan-ipoverlay/manager/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	NodeAccessPointDefaultKey = "Default"
)

type Location struct {
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	IP       string `json:"ip"`
}

type LocationData struct {
	Location *Location `json:"location"`
}
type LocationResp struct {
	Code int           `json:"code"`
	Data *LocationData `json:"data"`
	Msg  string        `json:"msg"`
}

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
	remoteIP := l.ctx.Value("Remote-IP")
	// logx.Debugf("remoteIP:%s", remoteIP)
	location, err := l.getLocalInfo(remoteIP.(string))
	if err != nil {
		return nil, fmt.Errorf("getLocalInfo failed:%v", err)
	}

	logx.Debugf("GetNodePop, location:%v", *location)
	if location.Province == "Hong Kong" {
		location.Country = "HongKong"
	}

	podConfig := l.getPodConfig(location.Country)
	if podConfig == nil {
		return nil, fmt.Errorf("not found pop config for node %s", req.NodeId)
	}

	server := l.getPodServer(podConfig.Id)
	if server == nil {
		return nil, fmt.Errorf("not found pop for node %s", req.NodeId)
	}

	getTokenResp, err := server.API.GetNodeAccessToken(l.ctx, &serverapi.GetNodeAccessTokenReq{NodeId: req.NodeId})
	if err != nil {
		return nil, err
	}

	accessPoint, ok := podConfig.NodeAccessPoints[location.Country]
	if ok {
		return &types.GetNodePopResp{ServerURL: accessPoint, AccessToken: getTokenResp.Token}, nil
	}

	accessPoint, ok = podConfig.NodeAccessPoints[NodeAccessPointDefaultKey]
	if !ok {
		return nil, fmt.Errorf("no default access point found")
	}

	logx.Debugf("GetNodePop, accessPoint:%s", accessPoint)
	return &types.GetNodePopResp{ServerURL: accessPoint, AccessToken: getTokenResp.Token}, nil
}

func (l *GetNodePopLogic) getPodServer(id string) *svc.Server {
	for podID, pop := range l.svcCtx.Servers {
		if podID == id {
			return pop
		}
	}
	return nil
}

func (l *GetNodePopLogic) getPodConfig(area string) *config.Pop {
	for _, pop := range l.svcCtx.Config.Pops {
		if pop.Area == area {
			return &pop
		}
	}

	for _, pop := range l.svcCtx.Config.Pops {
		if pop.Area == l.svcCtx.Config.DefaultArea {
			return &pop
		}
	}
	return nil
}

func (l *GetNodePopLogic) getLocalInfo(ip string) (*Location, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	url := fmt.Sprintf("%s?key=%s&ip=%s&language=en", l.svcCtx.Config.Geo.API, l.svcCtx.Config.Geo.Key, ip)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bs, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("StatusCode %d, msg:%s", resp.StatusCode, string(bs))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	locationResp := &LocationResp{}
	err = json.Unmarshal(body, locationResp)
	if err != nil {
		return nil, err
	}

	if locationResp.Code != 0 && locationResp.Code != 200 {
		return nil, fmt.Errorf("code:%d, msg:%s", locationResp.Code, locationResp.Msg)
	}

	return locationResp.Data.Location, nil
}
