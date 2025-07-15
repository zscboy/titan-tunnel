package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"titan-tunnel/server/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserCacheLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserCache(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserCacheLogic {
	return &DeleteUserCacheLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteUserCacheLogic) DeleteUserCache(userName string) error {
	url := fmt.Sprintf("http://localhost:%d/user/cache/delete", l.svcCtx.Config.APIServer)

	deleteUserCacheReq := struct {
		UserName string `json:"user_name"`
	}{
		UserName: userName,
	}

	jsonData, err := json.Marshal(deleteUserCacheReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		buf, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status code %s, error:%s", resp.StatusCode, string(buf))
	}

	return nil
}
