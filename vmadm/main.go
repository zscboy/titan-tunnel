package main

import (
	"fmt"
	"os"
	"titan-vm/vmadm/cmds"

	"github.com/urfave/cli/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

var rootCmds = []*cli.Command{
	cmds.SshCmd,
	cmds.VmCmds,
	cmds.NodeCmds,
	cmds.DownloadCmds,
	cmds.ImageCmds,
}

func main() {
	app := &cli.App{
		Name:  "vmadmin",
		Usage: "vms admin",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "server",
				Usage: "--server=localhost:7777",
				Value: "localhost:7777",
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
			debug := cctx.Bool("debug")
			if debug {
				logx.SetLevel(logx.DebugLevel)
			}
			logx.Debugf("debug:%#v", debug)

			return nil
		},
		Commands: rootCmds,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
}
