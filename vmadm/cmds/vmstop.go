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

var vmStop = &cli.Command{
	Name:  "stop",
	Usage: "stop {id}",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "--name vm name",
			Value: "",
		},
	},
	Action: func(cctx *cli.Context) error {
		id := cctx.Args().Get(0)
		if len(id) == 0 {
			return fmt.Errorf("need id, example: vmadm vm create <id> [options]")
		}
		request := api.StartVMRequest{
			Id:     id,
			VmName: cctx.String("name"),
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

		url := fmt.Sprintf("http://%s/vm/stop", cctx.String("server"))
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
