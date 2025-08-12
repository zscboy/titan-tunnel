package bootstrap

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 24 hours
	updateIntervals = 12 * 60 * 60
	bootstrapFile   = "bootstrap.json"
)

//go:embed bootstrap.json
var bootstrapJSON []byte

type Config struct {
	// bootstrap is url
	Bootstraps []string `json:"bootstraps"`
}

type BootstrapMgr struct {
	dir        string
	bootstraps []string
}

func NewBootstrapMgr(dir string) (*BootstrapMgr, error) {
	err := os.MkdirAll(dir, 0644)
	if err != nil {
		return nil, err
	}

	bootstrapFilePath := filepath.Join(dir, bootstrapFile)
	bytes, err := os.ReadFile(bootstrapFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			bytes = bootstrapJSON
			err := os.WriteFile(bootstrapFilePath, bytes, 0644)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var cfg Config
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}
	bmgr := &BootstrapMgr{dir: dir, bootstraps: cfg.Bootstraps}
	bmgr.getBootstrapsFromServer()

	go bmgr.TimedUpdate()

	return bmgr, nil
}

func (bmgr *BootstrapMgr) TimedUpdate() {
	ticker := time.NewTicker(updateIntervals * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		bmgr.getBootstrapsFromServer()

	}

}

func (bmgr *BootstrapMgr) Bootstraps() []string {
	return bmgr.bootstraps
}

func (bmgr *BootstrapMgr) getBootstrapsFromServer() {
	for _, bootstrapURL := range bmgr.bootstraps {
		bytes, err := bmgr.httpGet(bootstrapURL)
		if err != nil {
			logx.Errorf("BootstrapMgr.getBootstrapsFromServer, httpGet failed:%v, url:%s", err, bootstrapURL)
			continue
		}

		bootstrapFilePath := filepath.Join(bmgr.dir, bootstrapFile)
		err = os.WriteFile(bootstrapFilePath, bytes, 0644)
		if err != nil {
			logx.Errorf("BootstrapMgr.getBootstrapsFromServer, WriteFile failed:%v", err)
			continue
		}

		var cfg Config
		if err := json.Unmarshal(bytes, &cfg); err != nil {
			logx.Errorf("BootstrapMgr.getBootstrapsFromServer, Unmarshal failed:%v", err)
			continue
		}

		bmgr.bootstraps = cfg.Bootstraps
		break
	}
}

func (bmgr *BootstrapMgr) httpGet(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bs, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("StatusCode %d, msg:%s", resp.StatusCode, string(bs))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
