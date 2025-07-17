package main

import (
	"encoding/json"
	"time"

	"titan-tunnel/client/tunnel"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultUDPTimeout = 120
	defaultTCPTimeout = 3
)

// var globalCancel context.CancelFunc
var mytunnel *tunnel.Tunnel

func startTunnel(jsonParams string) *JSONCallResult {
	LogDebug("golib", "startTunnel: "+jsonParams)
	var input = struct {
		ServerURL  string `json:"server_url"`
		UUID       string `json:"uuid"`
		UDPTimeout int    `json:"udp_timeout"`
		TCPTimeout int    `json:"tcp_timeout"`
		Debug      bool   `json:"debug"`
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

	if len(input.ServerURL) == 0 || len(input.UUID) == 0 {
		return &JSONCallResult{Code: -1, Msg: "Params need server_url or uuid"}
	}

	if input.UDPTimeout == 0 {
		input.UDPTimeout = defaultUDPTimeout
	}

	if input.TCPTimeout == 0 {
		input.TCPTimeout = defaultTCPTimeout
	}

	tun, err := tunnel.NewTunnel(input.ServerURL, input.UUID, input.UDPTimeout, input.TCPTimeout)
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

		var err error
		for {
			err = tun.Connect()
			if err == nil {
				break
			}

			// logx.Error("wait seconds to retry connect")
			LogDebug("golib", "wait seconds to retry connect")
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
