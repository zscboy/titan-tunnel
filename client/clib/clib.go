package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"time"

	"titan-tunnel/client/tunnel"

	"github.com/zeromicro/go-zero/core/logx"
)

var globalCancel context.CancelFunc

//export StartTunnel
func StartTunnel(cUrl, cUuid *C.char, udpTimeout, tcpTimeout C.int, debug C.int) C.int {
	url := C.GoString(cUrl)
	uuid := C.GoString(cUuid)

	if debug != 0 {
		logx.SetLevel(logx.DebugLevel)
	} else {
		logx.SetLevel(logx.InfoLevel)
	}

	tun, err := tunnel.NewTunnel(url, uuid, int(udpTimeout), int(tcpTimeout))
	if err != nil {
		logx.Error("NewTunnel error:", err)
		return -1
	}

	if err = tun.Connect(); err != nil {
		logx.Error("Connect error:", err)
		return -2
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

	return 0
}

//export StopTunnel
func StopTunnel() {
	if globalCancel != nil {
		globalCancel()
	}
}
