package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"titan-tunnel/client/tunnel"

	"github.com/urfave/cli/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

func main() {
	app := &cli.App{
		Name:  "titan-tunnelc",
		Usage: "vms client",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "url",
				Usage: "--url=ws://localhost:8888/ws/node",
				Value: "ws://localhost:8888/ws/node",
			},
			&cli.StringFlag{
				Name:     "uuid",
				Usage:    "--uuid 08bd0658-1f61-11f0-8061-8bd115314f4c",
				Required: true,
				Value:    "",
			},

			&cli.IntFlag{
				Name:  "udp-timeout",
				Usage: "--udp-timeout 60, seconds",
				Value: 60,
			},
			&cli.IntFlag{
				Name:  "tcp-timeout",
				Usage: "--tcp-timeout 3, seconds",
				Value: 3,
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "--debug",
				Value: false,
			},
		},
		Before: func(cctx *cli.Context) error {
			return nil
		},
		Action: func(cctx *cli.Context) error {
			url := cctx.String("url")
			uuid := cctx.String("uuid")
			udpTimeout := cctx.Int("udp-timeout")
			tcpTimeout := cctx.Int("tcp-timeout")
			debug := cctx.Bool("debug")
			if debug {
				logx.SetLevel(logx.DebugLevel)
			} else {
				logx.SetLevel(logx.InfoLevel)
			}

			// ctx, done := context.WithCancel(cctx.Context)
			tun, err := tunnel.NewTunnel(url, uuid, udpTimeout, tcpTimeout)
			if err != nil {
				panic(err)
			}

			if err = tun.Connect(); err != nil {
				panic(err)
			}
			defer tun.Destroy()

			ctx, cancel := context.WithCancel(cctx.Context)
			go tunServe(tun, cancel)

			sigChan := make(chan os.Signal, 2)
			signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
			for {
				select {
				case <-sigChan:
					return nil
				case <-ctx.Done():
					return nil
				}
			}
		},
	}

	if err := app.Run(os.Args); err != nil {
		logx.Error(err)
	}
}

func tunServe(tun *tunnel.Tunnel, cancel context.CancelFunc) {
	defer cancel()
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

			logx.Error("wait seconds to retry connect")
			time.Sleep(5 * time.Second)
		}
	}
}
