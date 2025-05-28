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

var downloadTaskList = &cli.Command{
	Name:  "list",
	Usage: "delete downloading task not the local file",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) error {
		id := cctx.Args().Get(0)
		if len(id) == 0 {
			return fmt.Errorf("need id, example: vmadm vm create <id> [options]")
		}

		request := types.DownloadTaskListRequest{
			Id: id,
		}

		logx.Debugf("request:%#v", request)

		jsonData, err := json.Marshal(request)
		if err != nil {
			return err
		}

		logx.Debug("jsonData:", string(jsonData))

		url := fmt.Sprintf("http://%s/cmd/downloadtasklist", cctx.String("server"))
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

		var result types.DownloadTaskListResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		for _, task := range result.Tasks {
			fmt.Printf("TaskId:%s, md5:%s, path:%s, totalSize:%d, dwonloadSize:%d, Running:%v, URL:%s, success:%v, err:%s\n", task.TaskId, task.MD5, task.Path, task.TotalSize, task.DownloadSize, task.Running, task.URL, task.Success, task.ErrMsg)
		}
		return nil
	},
}
