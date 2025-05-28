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

var downloadimage = &cli.Command{
	Name:  "image",
	Usage: "image download iamge from server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "url",
			Usage:    "--url image url",
			Value:    "",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "md5",
			Usage:    "--md5 image md5",
			Value:    "",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "path",
			Usage:    "--md5 save image local path",
			Value:    "",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		id := cctx.Args().Get(0)
		if len(id) == 0 {
			return fmt.Errorf("need id, example: vmadm vm create <id> [options]")
		}

		request := types.DownloadImageRequest{
			Id:   id,
			URL:  cctx.String("url"),
			MD5:  cctx.String("md5"),
			Path: cctx.String("path"),
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

		url := fmt.Sprintf("http://%s/cmd/downloadimage", cctx.String("server"))
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("status code %d, error:%s", resp.StatusCode, string(body))
		}

		var result types.DownloadImageResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		fmt.Println("task id ", result.TaskId)
		return nil
	},
}
