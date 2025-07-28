package main

import (
	"encoding/json"
	"time"

	"titan-tunnel/client/bootstrap"
	"titan-tunnel/client/log"
	"titan-tunnel/client/tunnel"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultUDPTimeout = 120
	defaultTCPTimeout = 3
)

// var globalCancel context.CancelFunc
var mytunnel *tunnel.Tunnel
var bootstrapMgr *bootstrap.BootstrapMgr

func startTunnel(jsonParams string) *JSONCallResult {
	log.LogInfo("golib", "startTunnel: "+jsonParams)
	var input = struct {
		UUID   string `json:"uuid"`
		Debug  bool   `json:"debug"`
		AppDir string `json:"app_dir"`
	}{}

	err := json.Unmarshal([]byte(jsonParams), &input)
	if err != nil {
		return &JSONCallResult{
			Code: -1,
			Msg:  err.Error(),
		}
	}

	if input.Debug {
		logx.SetLevel(logx.DebugLevel)
	} else {
		logx.SetLevel(logx.InfoLevel)
	}

	if len(input.UUID) == 0 {
		return &JSONCallResult{Code: -1, Msg: "Params need uuid"}
	}

	if len(input.AppDir) == 0 {
		return &JSONCallResult{Code: -1, Msg: "Params need app_dir"}
	}

	if bootstrapMgr == nil {
		bootstrapMgr, err = bootstrap.NewBootstrapMgr(input.AppDir)
		if err != nil {
			return &JSONCallResult{Code: -1, Msg: err.Error()}
		}
	}

	if len(bootstrapMgr.Bootstraps()) == 0 {
		return &JSONCallResult{Code: -1, Msg: "No bootstrap nodes found"}
	}

	tun, err := tunnel.NewTunnel(input.UUID, defaultUDPTimeout, defaultTCPTimeout, bootstrapMgr.Bootstraps())
	if err != nil {
		logx.Error("NewTunnel error:", err)
		return &JSONCallResult{Code: -1, Msg: err.Error()}
	}

	if err = tun.Connect(); err != nil {
		logx.Error("Connect error:", err)
		return &JSONCallResult{Code: -1, Msg: err.Error()}
	}

	mytunnel = tun

	go tunServe(tun)

	return &JSONCallResult{Code: 0, Msg: "success"}
}

func tunServe(tun *tunnel.Tunnel) {
	defer logx.Info("tun client stop")
	for {
		tun.Serve()

		if tun.IsDestroy() {
			return
		}

		for {
			if err := tun.Connect(); err == nil {
				break
			}
			logx.Error("wait seconds to retry connect")
			time.Sleep(5 * time.Second)
		}
	}
}

func stopTunnel() *JSONCallResult {
	if mytunnel != nil {
		mytunnel.Destroy()
		mytunnel = nil
	}

	return &JSONCallResult{Code: 0, Msg: "success"}
}
