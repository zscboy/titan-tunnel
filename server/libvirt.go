package server

import (
	"fmt"

	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"github.com/gorilla/websocket"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

type GoLibvirt struct {
	serverURL string
}

func (goLibvirt *GoLibvirt) newLibvirt(agentID string) (*libvirt.Libvirt, error) {
	url := fmt.Sprintf("%s?uuid=%s", goLibvirt.serverURL, agentID)
	return newLibvirt(url)
}

func (goLibvirt *GoLibvirt) createVM(lv *libvirt.Libvirt, name, isoPath, diskPath string, vcpu uint, ramMB uint) error {
	return createVM(lv, name, isoPath, diskPath, vcpu, ramMB)
}

func (goLibvirt *GoLibvirt) startVM(lv *libvirt.Libvirt, vmName string) error {
	dom, err := lv.DomainLookupByName(vmName)
	if err != nil {
		return fmt.Errorf("can not find vm %s: %v", vmName, err)
	}

	return lv.DomainCreate(dom)
}

func (goLibvirt *GoLibvirt) stopVM(lv *libvirt.Libvirt, vmName string) error {
	dom, err := lv.DomainLookupByName(vmName)
	if err != nil {
		return fmt.Errorf("can not find vm %s: %v", vmName, err)
	}

	return lv.DomainDestroy(dom)
}

func (goLibvirt *GoLibvirt) deleteVM(lv *libvirt.Libvirt, vmName string) error {
	dom, err := lv.DomainLookupByName(vmName)
	if err != nil {
		return fmt.Errorf("can not find vm %s: %v", vmName, err)
	}

	return lv.DomainUndefine(dom)
}

func (goLibvirt *GoLibvirt) listVM(lv *libvirt.Libvirt, vmName string) error {
	return nil
}

func (goLibvirt *GoLibvirt) updateVM(lv *libvirt.Libvirt, domainXML string) error {
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

func createVM(lv *libvirt.Libvirt, name, isoPath, diskPath string, vcpu uint, ramMB uint) error {
	// 1. 自动生成 XML 配置
	domCfg := &libvirtxml.Domain{
		Type: "kvm",
		Name: name,
		Memory: &libvirtxml.DomainMemory{
			Value: ramMB,
			Unit:  "MiB",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: vcpu,
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Type: "hvm",
			},
			BootDevices: []libvirtxml.DomainBootDevice{
				{Dev: "hd"},    // 优先从硬盘启动
				{Dev: "cdrom"}, // 其次从光盘
			},
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
						Dev: "hda",
						Bus: "ide",
					},
					ReadOnly: &libvirtxml.DomainDiskReadOnly{},
				},
			},
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
			Graphics: []libvirtxml.DomainGraphic{
				{
					VNC: &libvirtxml.DomainGraphicVNC{
						Port: -1, // 自动分配端口
					},
				},
			},
		},
	}

	xml, err := domCfg.Marshal()
	if err != nil {
		return fmt.Errorf("generate xml failed: %s", err.Error())
	}

	dom, err := lv.DomainDefineXML(xml)
	if err != nil {
		return fmt.Errorf("define vm failed: %s", err.Error())
	}

	if err := lv.DomainCreate(dom); err != nil {
		return fmt.Errorf("create vm failed: %v", err)
	}

	return nil
}
