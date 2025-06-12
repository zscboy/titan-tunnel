package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"titan-vm/vmc/client"
	"titan-vm/vmc/wallet"

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

			&cli.StringFlag{
				Name:     "work-dir",
				Usage:    "--work-dir ./",
				Value:    "",
				Required: true,
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
			workDir := cctx.String("work-dir")

			if debug {
				log.SetLevel(log.DebugLevel)
			}

			wallet, err := loadWallet(workDir)
			if err != nil {
				log.Panic(err)
			}

			// ctx, done := context.WithCancel(cctx.Context)
			tun, err := client.NewTunnel(url, uuid, vmapi, wallet)
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

func loadWallet(workDir string) (*wallet.Wallet, error) {
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return nil, err
	}

	w, err := wallet.NewWallet(
		wallet.WithChainID(wallet.DefaultChainID),
		wallet.WithAccountPrefix(wallet.DefaultAccountPrefix),
		wallet.WithKeyringBackend(wallet.DefaultBackend),
		wallet.WithKeyDirectory(workDir),
	)
	if err != nil {
		return nil, err
	}

	infos, err := w.ListKeys()
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if info.GetName() == wallet.DefaultKeyName {
			addr, _ := w.GetAddress(wallet.DefaultKeyName)
			log.Debugf("wallet addr:%s", addr)
			return w, nil
		}
	}

	out, err := w.AddKey(wallet.DefaultKeyName, wallet.DefaultCoinType)
	if err != nil {
		return nil, err
	}

	log.Debugf("wallet addr:%s", out.Address)

	return w, nil
}
