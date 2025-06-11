package cmds

import (
	"encoding/json"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
	"titan-vm/pb"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/zeromicro/go-zero/core/logx"
)

type HostInfo struct {
}

func NewHostInfo() *HostInfo {
	return &HostInfo{}
}

func (hostInfo *HostInfo) Get() *pb.CmdHostInfoResponse {
	logx.Debug("HostInfo.Get")
	return hostInfo.get()
}

func (hostInfo *HostInfo) get() *pb.CmdHostInfoResponse {
	cpuInfo := &pb.CPU{Num: int32(runtime.NumCPU()), Arch: runtime.GOARCH}

	cpuPercent, _ := cpu.Percent(time.Second, false)
	cpuInfo.Usage = float32(cpuPercent[0])

	// 获取 CPU 详细信息
	info, _ := cpu.Info()
	for _, c := range info {
		cpuInfo.Vendor = c.ModelName
	}

	memState, _ := mem.VirtualMemory()
	memInfo := pb.Memory{Total: int64(memState.Total), Used: int64(memState.Used)}

	interfaces, _ := net.Interfaces()

	ifaces := make([]*pb.NetwrokInterface, 0, len(interfaces))
	for _, iface := range interfaces {
		// fmt.Printf("Interface: %s, MAC: %s, state:%s\n", iface.Name, iface.HardwareAddr, iface.Flags.String())
		networkInterface := &pb.NetwrokInterface{Name: iface.Name, Mac: iface.HardwareAddr.String(), Flags: uint32(iface.Flags)}
		ifaces = append(ifaces, networkInterface)

	}

	response := &pb.CmdHostInfoResponse{Cpu: cpuInfo, Memory: &memInfo, Interfaces: ifaces}

	if runtime.GOOS == "windows" {
		disks, _ := hostInfo.getDiskInfoForWindows()
		response.Disks = disks
	} else if runtime.GOOS == "linux" {
		disks, _ := hostInfo.getDiskInfoForLinux()
		response.Disks = disks
	}

	return response
}

type BlockDevice struct {
	Name       string        `json:"name"`
	Size       int64         `json:"size"`
	Type       string        `json:"type"`
	Fstype     string        `json:"fstype"`
	Mountpoint string        `json:"mountpoint"`
	Fsused     string        `json:"fsused"`
	Children   []BlockDevice `json:"children"`
}

type LsblkOutput struct {
	Blockdevices []BlockDevice `json:"blockdevices"`
}

func (hostInfo *HostInfo) getDiskInfoForLinux() ([]*pb.Disk, error) {
	cmd := exec.Command("lsblk", "-b", "-p", "-o", "NAME,SIZE,TYPE,FSTYPE,MOUNTPOINT,FSUSED", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result LsblkOutput
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}

	disks := make([]*pb.Disk, 0, len(result.Blockdevices))
	for _, blockDevice := range result.Blockdevices {
		if strings.HasPrefix(blockDevice.Name, "/dev/loop") || strings.HasPrefix(blockDevice.Name, "/dev/ram") {
			continue
		}

		ds := processBlockDevice(&blockDevice)
		disks = append(disks, ds...)
	}

	return disks, nil
}

func (hostInfo *HostInfo) getDiskInfoForWindows() ([]*pb.Disk, error) {
	partitions, _ := disk.Partitions(false)
	disks := make([]*pb.Disk, 0, len(partitions))
	for _, part := range partitions {
		d := &pb.Disk{Name: part.Device, MountPoint: part.Mountpoint, FsType: part.Fstype}
		disks = append(disks, d)

		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			continue
		}

		d.Total = usage.Total
		d.Used = usage.Used
		d.Type = "part"
	}

	return disks, nil
}

func parseSize(bytesStr string) uint64 {
	b, err := strconv.ParseUint(bytesStr, 10, 64)
	if err != nil {
		return 0
	}
	return b
}

func blockDeviceToPBDisk(blockDevice *BlockDevice) *pb.Disk {
	disk := &pb.Disk{
		Name:       blockDevice.Name,
		Type:       blockDevice.Type,
		Total:      uint64(blockDevice.Size),
		Used:       parseSize(blockDevice.Fsused),
		MountPoint: blockDevice.Mountpoint,
		FsType:     blockDevice.Fstype,
	}
	return disk
}

func processBlockDevice(device *BlockDevice) []*pb.Disk {
	disk := blockDeviceToPBDisk(device)

	disks := make([]*pb.Disk, 0)
	disks = append(disks, disk)
	for _, child := range device.Children {
		if child.Size <= 0 {
			continue
		}
		ds := processBlockDevice(&child)
		disks = append(disks, ds...)
	}

	return disks
}
