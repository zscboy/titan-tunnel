package cmds

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"titan-vm/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type NvmeInfo struct {
}

func NewNvmeInfo() *NvmeInfo {
	return &NvmeInfo{}
}

func (nvmeInfo *NvmeInfo) List() *pb.CmdListNvmeResponse {
	logx.Debug("HostInfo.Get")
	return nvmeInfo.list()
}

func (nvmeInfo *NvmeInfo) list() *pb.CmdListNvmeResponse {
	rsp := &pb.CmdListNvmeResponse{Nvmes: make([]*pb.NvmeInfo, 0)}
	if runtime.GOOS != "linux" {
		return rsp
	}

	blockPath := "/sys/block"
	entries, err := os.ReadDir(blockPath)
	if err != nil {
		return rsp
	}

	nvmes := make([]*pb.NvmeInfo, 0)
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "nvme") {
			continue
		}

		addressPath := filepath.Join(blockPath, name, "device", "address")
		addressBytes, err := os.ReadFile(addressPath)
		if err != nil {
			continue
		}

		address := strings.TrimSpace(string(addressBytes))
		nvme := &pb.NvmeInfo{Name: name, PciAddr: address}
		nvmes = append(nvmes, nvme)
	}

	return &pb.CmdListNvmeResponse{Nvmes: nvmes}
}
