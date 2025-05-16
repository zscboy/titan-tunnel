package ws

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"titan-vm/pb"
	"titan-vm/vms/ws/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/protobuf/proto"
)

type CmdHandler struct {
	tunMgr *TunnelManager
}

func newCmdHandler(tunMgr *TunnelManager) *CmdHandler {
	return &CmdHandler{tunMgr: tunMgr}

}

func (cmd *CmdHandler) downloadImageHandler(w http.ResponseWriter, r *http.Request) {
	var req types.DownloadImageRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	tun := cmd.tunMgr.getTunnel(req.Id)
	if tun == nil {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("not found %s", req.Id))
		return
	}

	// tun.addWait()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	request := &pb.CmdDownloadImageRequest{Url: req.URL}
	bytes, err := proto.Marshal(request)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}

	resp := &pb.CmdDownloadTaskControlResponse{}
	payload := &pb.Command{Type: pb.CommandType_DownloadImage, Data: bytes}
	err = tun.sendCommand(ctx, payload, resp)
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
	} else {
		httpx.OkJsonCtx(r.Context(), w, resp)
	}

}
