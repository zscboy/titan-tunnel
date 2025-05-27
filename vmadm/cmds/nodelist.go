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

var nodeList = &cli.Command{
	Name:  "list",
	Usage: "list node",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "start",
			Usage: "--start 0",
			Value: 0,
		},
		&cli.IntFlag{
			Name:  "end",
			Usage: "--end 20",
			Value: 20,
		},
	},
	Action: func(cctx *cli.Context) error {
		request := api.ListNodeReqeust{
			Start: cctx.Int("start"),
			End:   cctx.Int("end"),
		}

		logx.Debugf("request:%#v", request)

		jsonData, err := json.Marshal(request)
		if err != nil {
			return err
		}

		logx.Debug("jsonData:", string(jsonData))

		url := fmt.Sprintf("http://%s/vm/node/list", cctx.String("server"))
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

		var result api.ListNodeResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		for _, node := range result.Nodes {
			fmt.Printf("Id:%s, os:%s, vmType:%s, ip:%s\n", node.Id, node.OS, node.VmType, node.IP)
		}

		return nil
	},
}
