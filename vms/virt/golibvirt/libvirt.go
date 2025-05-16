package golibvirt

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"titan-vm/vms/pb"

	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"github.com/gorilla/websocket"
	libvirtxml "github.com/libvirt/libvirt-go-xml"

	multipassPb "titan-vm/vms/virt/multipass/pb"
)

const transport = "raw"
const vmapi = "libvirt"

type GoLibvirt struct {
	serverURL string
	clients   sync.Map
}

func NewGoLibvirt(serverURL string) *GoLibvirt {
	return &GoLibvirt{serverURL: serverURL}
}

func (goLibvirt *GoLibvirt) connectHost(hostID string) (*libvirt.Libvirt, error) {
	v, ok := goLibvirt.clients.Load(hostID)
	if ok {
		lv := v.(*libvirt.Libvirt)
		if lv.IsConnected() {
			return lv, nil
		}
		goLibvirt.clients.Delete(hostID)
	}

	url := fmt.Sprintf("%s?transport=%s&vmapi=%s&uuid=%s", goLibvirt.serverURL, transport, vmapi, hostID)
	lv, err := newLibvirt(url)
	if err != nil {
		return nil, err
	}

	goLibvirt.clients.Store(hostID, lv)

	return lv, nil
}

func (goLibvirt *GoLibvirt) CreateVMWithLibvirt(_ context.Context, request *pb.CreateVMWithLibvirtRequest) error {
	lv, err := goLibvirt.connectHost(request.GetId())
	if err != nil {
		return err
	}

	defer lv.Disconnect()

	domain := createInstanceXML(request.GetVmName(), request.GetIsoPath(), request.GetDiskPath(), uint(request.GetCpu()), uint(request.GetMemory()))
	xml, err := domain.Marshal()
	if err != nil {
		return fmt.Errorf("generate xml failed: %s", err.Error())
	}

	dom, err := lv.DomainDefineXML(xml)
	if err != nil {
		return fmt.Errorf("define vm failed: %s", err.Error())
	}

	return lv.DomainCreate(dom)
}

func (goLibvirt *GoLibvirt) CreateVMWithMultipass(_ context.Context, request *pb.CreateVMWithMultipassRequest, progressChan chan<- *multipassPb.LaunchProgress) error {
	return nil
}

func (goLibvirt *GoLibvirt) StartVM(_ context.Context, request *pb.StartVMRequest) error {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return fmt.Errorf("can not find vm %s: %v", request.GetVmName(), err)
	}

	return lv.DomainCreate(dom)
}

func (goLibvirt *GoLibvirt) StopVM(_ context.Context, request *pb.StopVMRequest) error {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return fmt.Errorf("can not find vm %s: %v", request.GetVmName(), err)
	}

	return lv.DomainDestroy(dom)
}

func (goLibvirt *GoLibvirt) DeleteVM(_ context.Context, request *pb.DeleteVMRequest) error {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return fmt.Errorf("can not find vm %s: %v", request.GetVmName(), err)
	}

	return lv.DomainUndefine(dom)
}

func (goLibvirt *GoLibvirt) ListVMInstance(_ context.Context, request *pb.ListVMInstanceReqeust) (*pb.ListVMInstanceResponse, error) {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return nil, err
	}
	defer lv.Disconnect()

	domains, _, err := lv.ConnectListAllDomains(1, 0)
	if err != nil {
		return nil, err
	}

	vmInfos := make([]*pb.VMInfo, 0, len(domains))
	for _, domain := range domains {
		state, _, _, _, _, err := lv.DomainGetInfo(domain)
		if err != nil {
			continue
		}
		vmInfo := &pb.VMInfo{Name: domain.Name, State: parseState(libvirt.DomainState(state))}
		vmInfos = append(vmInfos, vmInfo)
	}
	return &pb.ListVMInstanceResponse{VmInfos: vmInfos}, nil
}

func (goLibvirt *GoLibvirt) ListImage(_ context.Context, request *pb.ListImageRequest) (*pb.ListImageResponse, error) {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return nil, err
	}
	defer lv.Disconnect()

	pools, _, err := lv.ConnectListAllStoragePools(1, 0)
	if err != nil {
		return nil, err
	}

	images := make([]string, 0)
	for _, pool := range pools {
		volumes, _, err := lv.StoragePoolListAllVolumes(pool, 1, 0)
		if err != nil {
			continue
		}

		for _, vol := range volumes {
			if strings.HasSuffix(vol.Name, ".iso") || strings.HasSuffix(vol.Name, ".qcow2") || strings.HasSuffix(vol.Name, ".raw") {
				images = append(images, vol.Key)
			}
		}
	}
	return &pb.ListImageResponse{Images: images}, nil
}

func (goLibvirt *GoLibvirt) CreateVolWithLibvirt(ctx context.Context, request *pb.CreateVolWithLibvirtReqeust) (*pb.CreateVolWithLibvirtResponse, error) {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return nil, err
	}
	defer lv.Disconnect()

	storagePool, err := lv.StoragePoolLookupByName(request.Pool)
	if err != nil {
		return nil, err
	}

	vol := libvirtxml.StorageVolume{
		Name:     request.Name,
		Capacity: &libvirtxml.StorageVolumeSize{Unit: "G", Value: uint64(request.GetCapacity())},
		Target:   &libvirtxml.StorageVolumeTarget{Format: &libvirtxml.StorageVolumeTargetFormat{Type: request.Format}},
	}

	xmlString, err := vol.Marshal()
	if err != nil {
		return nil, err
	}

	rVol, err := lv.StorageVolCreateXML(storagePool, xmlString, 0)
	if err != nil {
		return nil, err
	}

	return &pb.CreateVolWithLibvirtResponse{Pool: rVol.Pool, Name: rVol.Name, Key: rVol.Key}, nil
}

func (goLibvirt *GoLibvirt) GetVol(ctx context.Context, request *pb.GetVolRequest) (*pb.GetVolResponse, error) {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return nil, err
	}
	defer lv.Disconnect()

	storagePool, err := lv.StoragePoolLookupByName(request.PoolName)
	if err != nil {
		return nil, err
	}

	storageVol, err := lv.StorageVolLookupByName(storagePool, request.VolName)
	if err != nil {
		return nil, err
	}

	_, capacity, _, err := lv.StorageVolGetInfo(storageVol)
	if err != nil {
		return nil, err
	}

	path, err := lv.StorageVolGetPath(storageVol)
	if err != nil {
		return nil, err
	}

	return &pb.GetVolResponse{Name: request.VolName, Pool: request.PoolName, Capacity: int32(capacity), Path: path}, nil
}

func (goLibvirt *GoLibvirt) UpdateVM(_ context.Context, request *pb.UpdateVMRequest) error {
	return nil
}

func newLibvirt(urlStr string) (*libvirt.Libvirt, error) {
	conn, _, err := websocket.DefaultDialer.Dial(urlStr, nil)
	if err != nil {
		return nil, err
	}

	l := libvirt.NewWithDialer(dialers.NewAlreadyConnected(conn.NetConn()))
	if err := l.Connect(); err != nil {
		return nil, err
	}

	return l, nil
}

func createInstanceXML(name, isoPath, diskPath string, vcpu uint, memoryMB uint) *libvirtxml.Domain {
	domain := &libvirtxml.Domain{
		Type: "kvm",
		Name: name,
		Metadata: &libvirtxml.DomainMetadata{
			XML: `
			<libosinfo:libosinfo xmlns:libosinfo="http://libosinfo.org/xmlns/libvirt/domain/1.0">
				<libosinfo:os id="http://centos.org/centos/7.0"/>
			</libosinfo:libosinfo>`,
		},
		Memory: &libvirtxml.DomainMemory{
			Unit:  "KiB",
			Value: uint(memoryMB * 1024),
		},
		VCPU: &libvirtxml.DomainVCPU{
			Placement: "static",
			Value:     uint(vcpu),
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Arch:    "x86_64",
				Machine: "pc-q35-6.2",
				Type:    "hvm",
			},
			BootDevices: []libvirtxml.DomainBootDevice{
				{Dev: "hd"},
				{Dev: "cdrom"},
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
		},
		CPU: &libvirtxml.DomainCPU{
			Mode:       "host-passthrough",
			Check:      "none",
			Migratable: "on",
		},
		Clock: &libvirtxml.DomainClock{
			Offset: "utc",
			Timer: []libvirtxml.DomainTimer{
				{Name: "rtc", TickPolicy: "catchup"},
				{Name: "pit", TickPolicy: "delay"},
				{Name: "hpet", Present: "no"},
			},
		},
		OnPoweroff: "destroy",
		OnReboot:   "restart",
		OnCrash:    "destroy",
		PM: &libvirtxml.DomainPM{
			SuspendToMem:  &libvirtxml.DomainPMPolicy{Enabled: "no"},
			SuspendToDisk: &libvirtxml.DomainPMPolicy{Enabled: "no"},
		},
		Devices: &libvirtxml.DomainDeviceList{
			Disks: []libvirtxml.DomainDisk{
				{
					Device: "disk",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "qcow2",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: diskPath,
						},
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "vda",
						Bus: "virtio",
					},
				},
				{
					Device: "cdrom",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "raw",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: isoPath,
						},
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "sda",
						Bus: "sata",
					},
					ReadOnly: &libvirtxml.DomainDiskReadOnly{},
				},
			},
			Controllers: []libvirtxml.DomainController{},
			Interfaces: []libvirtxml.DomainInterface{
				{
					Source: &libvirtxml.DomainInterfaceSource{
						Network: &libvirtxml.DomainInterfaceSourceNetwork{
							Network: "default",
						},
					},
					Model: &libvirtxml.DomainInterfaceModel{
						Type: "virtio",
					},
				},
			},
			Serials: []libvirtxml.DomainSerial{
				{Target: &libvirtxml.DomainSerialTarget{}},
			},
			Consoles: []libvirtxml.DomainConsole{
				{Target: &libvirtxml.DomainConsoleTarget{}},
			},
			Inputs: []libvirtxml.DomainInput{
				{Type: "tablet", Bus: "usb"},
				{Type: "mouse", Bus: "ps2"},
				{Type: "keyboard", Bus: "ps2"},
			},

			Graphics: []libvirtxml.DomainGraphic{
				{
					VNC: &libvirtxml.DomainGraphicVNC{
						Port:     -1,
						AutoPort: "yes",
					},
				},
			},
			Videos: []libvirtxml.DomainVideo{
				{
					Model: libvirtxml.DomainVideoModel{
						Type:    "vga",
						Primary: "yes",
					},
				},
			},
			MemBalloon: &libvirtxml.DomainMemBalloon{
				Model: "virtio",
			},
			RNGs: []libvirtxml.DomainRNG{
				{
					Model: "virtio",
					Backend: &libvirtxml.DomainRNGBackend{
						Random: &libvirtxml.DomainRNGBackendRandom{
							Device: "/dev/urandom",
						},
					},
				},
			},
		},
	}

	return domain
}

func parseState(state libvirt.DomainState) string {
	switch state {
	case libvirt.DomainRunning:
		return "Running"
	case libvirt.DomainBlocked:
		return "Blocked"
	case libvirt.DomainPaused:
		return "Paused"
	case libvirt.DomainShutdown:
		return "Shutting down"
	case libvirt.DomainShutoff:
		return "Shut off"
	case libvirt.DomainCrashed:
		return "Crashed"
	default:
		return "Unknown"
	}
}
