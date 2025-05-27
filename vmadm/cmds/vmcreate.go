package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	api "titan-vm/vms/api/export"

	"github.com/urfave/cli/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

var vmCreate = &cli.Command{
	Name:      "create",
	Usage:     "create {id}",
	UsageText: "create <id> [options]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "--name vm name",
			Value: "",
		},
		&cli.IntFlag{
			Name:  "cpu",
			Usage: "--cpu cpu number",
			Value: 0,
		},
		&cli.IntFlag{
			Name:  "memory",
			Usage: "--memory unit MB",
			Value: 0,
		},
		&cli.IntFlag{
			Name:  "disk-size",
			Usage: "--disk-size unit GB",
			Value: 0,
		},
		&cli.StringFlag{
			Name:  "image",
			Usage: "--image",
			Value: "",
		},
	},
	Action: func(cctx *cli.Context) error {
		id := cctx.Args().Get(0)
		if len(id) == 0 {
			return fmt.Errorf("need id, example: vmadm vm create <id> [options]")
		}

		logx.Debugf("Image path: %s", cctx.String("image"))
		request := api.CreateVMRequest{
			Id:       id,
			VmName:   cctx.String("name"),
			Cpu:      int32(cctx.Int("cpu")),
			Memory:   int32(cctx.Int("memory")),
			DiskSize: int32(cctx.Int("disk-size")),
			Image:    cctx.String("image"),
		}

		logx.Debugf("request:%#v", request)

		jsonData, err := json.Marshal(request)
		if err != nil {
			return err
		}

		logx.Debug("jsonData:", string(jsonData))

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		url := fmt.Sprintf("http://%s/vm/create", cctx.String("server"))
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("status code %d, error:%s", resp.StatusCode, string(body))
		}

		var result api.VMOperationResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		if !result.Success {
			return fmt.Errorf(result.Message)
		}

		return nil
	},
}
