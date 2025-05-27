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

var imageDelete = &cli.Command{
	Name:      "delete",
	Usage:     "delete image from target node",
	UsageText: "delete <node-id> [options]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "path",
			Usage: "--path the image path",
			Value: "",
		},
	},
	Action: func(cctx *cli.Context) error {
		id := cctx.Args().Get(0)
		if len(id) == 0 {
			return fmt.Errorf("need id, example: vmadm vm create <id> [options]")
		}
		request := api.ListVMInstanceReqeust{
			Id: id,
		}

		logx.Debugf("request:%#v", request)

		jsonData, err := json.Marshal(request)
		if err != nil {
			return err
		}

		logx.Debug("jsonData:", string(jsonData))

		url := fmt.Sprintf("http://%s/vm/list", cctx.String("server"))
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("status code %d, error:%s", resp.StatusCode, string(body))
		}

		var result api.ListVMInstanceResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		for _, instance := range result.VmInfos {
			fmt.Printf("Name:%s, state:%s, ip:%s\n", instance.Name, instance.State, instance.Ip)
		}

		return nil
	},
}
