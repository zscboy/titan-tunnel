package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"time"

	"titan-tunnel/client/tunnel"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultUDPTimeout = 120
	defaultTCPTimeout = 3
)

var globalCancel context.CancelFunc

func startTunnel(jsonParams string /* cUrl, cUuid *C.char, udpTimeout, tcpTimeout C.int, debug C.int*/) *JSONCallResult {

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

	ctx, cancel := context.WithCancel(context.Background())
	globalCancel = cancel

	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				logx.Info("Tunnel stopped by context")
				return
			default:
				tun.Serve()
				if tun.IsDestroy() {
					return
				}
				logx.Error("wait 5 seconds to retry connect")
				time.Sleep(5 * time.Second)

				for {
					select {
					case <-ctx.Done():
						return
					default:
						if err := tun.Connect(); err == nil {
							break
						}
						logx.Error("connect failed again, retry in 5s")
						time.Sleep(5 * time.Second)
					}
				}
			}
		}
	}()

	return &JSONCallResult{Code: 0, Msg: "success"}
}

func stopTunnel() *JSONCallResult {
	if globalCancel != nil {
		globalCancel()
	}

	return &JSONCallResult{Code: 0, Msg: "success"}
}
