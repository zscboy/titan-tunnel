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
)

var imageList = &cli.Command{
	Name:      "list",
	Usage:     "list image from target node",
	UsageText: "list <node-id> [options]",
	Action: func(cctx *cli.Context) error {
		id := cctx.Args().Get(0)
		if len(id) == 0 {
			return fmt.Errorf("need id, example: vmadm iamge list <id> [options]")
		}
		request := api.ListImageRequest{
			Id: id,
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("http://%s/vm/image/list", cctx.String("server"))
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

		var result api.ListImageResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		for _, image := range result.Images {
			fmt.Printf("Name:%s\n", image)
		}

		return nil
	},
}
