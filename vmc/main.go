package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"titan-vm/vmc/client"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "titan-vmc",
		Usage: "vms client",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "url",
				Usage: "--url=ws://localhost:8020/ws",
				Value: "ws://localhost:8020/ws",
			},
			&cli.StringFlag{
				Name:     "uuid",
				Usage:    "--uuid 08bd0658-1f61-11f0-8061-8bd115314f4c",
				Required: true,
				Value:    "",
			},
			&cli.StringFlag{
				Name:  "vmapi",
				Usage: "--vmapi libvirt or multipass",
				Value: "multipass",
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
			debug := cctx.Bool("debug")
			vmapi := cctx.String("vmapi")

			if debug {
				log.SetLevel(log.DebugLevel)
			}

			// ctx, done := context.WithCancel(cctx.Context)
			tun, err := client.NewTunnel(url, uuid, vmapi)
			if err != nil {
				log.Panic(err)
			}

			if err = tun.Connect(); err != nil {
				log.Panic(err)
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
		log.Fatal(err)
	}
}

func tunServe(tun *client.Tunnel, cancel context.CancelFunc) {
	defer cancel()
	for {
		tun.Serve()

		if tun.IsDestroy() {
			return
		}

		// wait 3 seconds to reconnet
		// time.Sleep(3 * time.Second)
		var err error
		var i = 0
		for ; i < 10; i++ {
			err = tun.Connect()
			if err == nil {
				break
			}

			log.Error("wait seconds to retry connect")
			time.Sleep(5 * time.Second)
		}

		if err != nil {
			log.Errorf("connected failed:%s", err.Error())
			return
		}
	}
}
