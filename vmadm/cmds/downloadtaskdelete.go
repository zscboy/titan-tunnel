package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	types "titan-vm/vms/api/export"

	"github.com/urfave/cli/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

var downloadTaskDelete = &cli.Command{
	Name:  "delete",
	Usage: "delete downloading task not the local file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "task-id",
			Usage: "--task-id",
			Value: "",
		},
	},
	Action: func(cctx *cli.Context) error {
		id := cctx.Args().Get(0)
		if len(id) == 0 {
			return fmt.Errorf("need id, example: vmadm vm create <id> [options]")
		}

		request := types.DownloadTaskDeleteRequest{
			Id:     id,
			TaskId: cctx.String("task-id"),
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

		url := fmt.Sprintf("http://%s/cmd/downloadtaskdelete", cctx.String("server"))
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("status code %d, error:%s", resp.StatusCode, string(body))
		}

		var result types.DownloadTaskDeleteResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		if !result.Success {
			return fmt.Errorf(result.ErrMsg)
		}
		return nil
	},
}
