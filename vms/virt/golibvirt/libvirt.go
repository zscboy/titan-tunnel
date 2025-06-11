package golibvirt

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	"titan-vm/vms/pb"

	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	libvirtxml "github.com/libvirt/libvirt-go-xml"

	multipassPb "titan-vm/vms/virt/multipass/pb"
)

type GoLibvirt struct {
	serverURL string
	clients   sync.Map
}

func generateJwtToken(secret string, expire int64) (string, error) {
	claims := jwt.MapClaims{
		"user": "golibvirt",
		"exp":  time.Now().Add(time.Second * time.Duration(expire)).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))

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

	url := fmt.Sprintf("%s?transport=%s&vmapi=%s&id=%s", goLibvirt.serverURL, transport, vmapi, hostID)
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

	state, _, _, _, _, err := lv.DomainGetInfo(dom)
	if err != nil {
		return err
	}

	if libvirt.DomainState(state) == libvirt.DomainRunning {
		return fmt.Errorf("can not delete the running vm")
	}

	xmlDesc, err := lv.DomainGetXMLDesc(dom, 0)
	if err != nil {
		return err
	}

	var domain libvirtxml.Domain
	if err := xml.Unmarshal([]byte(xmlDesc), &domain); err != nil {
		return err
	}

	if err := lv.DomainUndefine(dom); err != nil {
		return err
	}

	return goLibvirt.deleteFirstDiskFromDomain(lv, domain)
}

func (goLibvirt *GoLibvirt) deleteFirstDiskFromDomain(lv *libvirt.Libvirt, domain libvirtxml.Domain) error {
	for _, disk := range domain.Devices.Disks {
		if disk.Device != "disk" && disk.Driver.Type != "qcow2" {
			continue
		}

		volPath := disk.Source.File.File
		storageVol, err := lv.StorageVolLookupByPath(volPath)
		if err != nil {
			return err
		}

		lv.StorageVolDelete(storageVol, 0)

		break
	}

	return nil
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
		lv.StoragePoolRefresh(pool, 0)
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

func (goLibvirt *GoLibvirt) DeleteImage(_ context.Context, request *pb.DeleteImageRequest) error {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer lv.Disconnect()

	vol, err := lv.StorageVolLookupByPath(request.Path)
	if err != nil {
		return err
	}

	if err = lv.StorageVolDelete(vol, 0); err != nil {
		return err
	}
	return nil
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

func (goLibvirt *GoLibvirt) ListHostNetworkInterfaceWithLibvirt(ctx context.Context, request *pb.ListHostNetworkInterfaceRequest) (*pb.ListHostNetworkInterfaceResponse, error) {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return nil, err
	}
	defer lv.Disconnect()

	ifs, _, err := lv.ConnectListAllInterfaces(1, libvirt.ConnectListInterfacesActive)
	if err != nil {
		return nil, err
	}

	interfaces := make([]*pb.HostNetworkInterface, 0, len(ifs))
	for _, iface := range ifs {
		netInterface := &pb.HostNetworkInterface{Name: iface.Name, Mac: iface.Mac, Active: true}
		interfaces = append(interfaces, netInterface)
	}

	ifs, _, err = lv.ConnectListAllInterfaces(1, libvirt.ConnectListInterfacesInactive)
	if err != nil {
		return nil, err
	}

	for _, iface := range ifs {
		netInterface := &pb.HostNetworkInterface{Name: iface.Name, Mac: iface.Mac, Active: false}
		interfaces = append(interfaces, netInterface)
	}

	return &pb.ListHostNetworkInterfaceResponse{Interfaces: interfaces}, nil
}

func (goLibvirt *GoLibvirt) ListVMNetwrokInterfaceWithLibvirt(ctx context.Context, request *pb.ListVMNetwrokInterfaceReqeust) (*pb.ListVMNetworkInterfaceResponse, error) {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return nil, err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return nil, err
	}

	xmlDesc, err := lv.DomainGetXMLDesc(dom, 0)
	if err != nil {
		return nil, err
	}

	var domain libvirtxml.Domain
	if err := xml.Unmarshal([]byte(xmlDesc), &domain); err != nil {
		return nil, err
	}

	interfaces := goLibvirt.listVMNetwrokInterfaces(&domain)
	return &pb.ListVMNetworkInterfaceResponse{Interfaces: interfaces}, nil
}

func (goLibvirt *GoLibvirt) listVMNetwrokInterfaces(domain *libvirtxml.Domain) []*pb.VMNetworkInterface {
	ifs := domain.Devices.Interfaces

	interfaces := make([]*pb.VMNetworkInterface, 0, len(ifs))
	for _, iface := range ifs {
		// _ = iface
		netInterface := &pb.VMNetworkInterface{
			Mac:   iface.MAC.Address,
			Model: iface.Model.Type,
		}

		if iface.Target != nil {
			netInterface.Name = iface.Target.Dev
		}

		if iface.Source.Network != nil {
			netInterface.Type = interfaceTypeNetwork
			netInterface.Source = iface.Source.Network.Network

		} else if iface.Source.Direct != nil {
			netInterface.Type = interfaceTypeDirect
			netInterface.Source = iface.Source.Direct.Dev
			netInterface.SourceModel = iface.Source.Direct.Mode

		} else if iface.Source.Bridge != nil {
			netInterface.Type = interfaceTypeBridge
			netInterface.Source = iface.Source.Bridge.Bridge
		}

		interfaces = append(interfaces, netInterface)
	}

	return interfaces
}

func (goLibvirt *GoLibvirt) AddNetworkInterfaceWithLibvirt(ctx context.Context, request *pb.AddNetworkInterfaceRequest) error {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return err
	}

	state, _, _, _, _, err := lv.DomainGetInfo(dom)
	if err != nil {
		return err
	}

	flags := uint32(libvirt.DomainDeviceModifyConfig)
	if state == uint8(libvirt.DomainRunning) {
		flags = uint32(libvirt.DomainDeviceModifyLive) | flags
	}

	newInterface := libvirtxml.DomainInterface{
		Model:  &libvirtxml.DomainInterfaceModel{Type: interfaceModelType},
		Source: &libvirtxml.DomainInterfaceSource{},
	}
	if request.Type == pb.InterfaceType_INTERFACE_TYPE_NETWORK {
		newInterface.Source.Network = &libvirtxml.DomainInterfaceSourceNetwork{
			Network: interfaceSourceNetwork,
		}
	} else if request.Type == pb.InterfaceType_INTERFACE_TYPE_DIRECT {
		newInterface.Driver = &libvirtxml.DomainInterfaceDriver{
			Name: interfaceDriverName, Queues: interfaceDriverQueues,
		}

		newInterface.Source.Direct = &libvirtxml.DomainInterfaceSourceDirect{
			Dev: request.SourceDirectDev,
		}

		if request.Model == pb.InterfaceSourceDirectModel_INTERFACE_SOURCE_DIRECT_MODEL_BRIDGE {
			newInterface.Source.Direct.Mode = interfaceSourceDirectModelBridge
		} else if request.Model == pb.InterfaceSourceDirectModel_INTERFACE_SOURCE_DIRECT_MODEL_PASSTHROUGH {
			newInterface.Source.Direct.Mode = interfaceSourceDirectModelPassthrough
		} else {
			return fmt.Errorf("unsupport interface source direct model %d", request.Model)
		}
	} else {
		return fmt.Errorf("unsupport interface type %d", request.Type)
	}

	xml, err := newInterface.Marshal()
	if err != nil {
		return err
	}

	return lv.DomainAttachDeviceFlags(dom, xml, flags)
}
func (goLibvirt *GoLibvirt) DeleteNetworkInterfaceWithLibvirt(ctx context.Context, request *pb.DeleteNetworkInterfaceRequest) error {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return err
	}

	state, _, _, _, _, err := lv.DomainGetInfo(dom)
	if err != nil {
		return err
	}

	xmlDesc, err := lv.DomainGetXMLDesc(dom, 0)
	if err != nil {
		return err
	}

	var domainXML libvirtxml.Domain
	err = domainXML.Unmarshal(xmlDesc)
	if err != nil {
		return err
	}

	var targetInterface *libvirtxml.DomainInterface
	for _, iface := range domainXML.Devices.Interfaces {
		if iface.MAC != nil && iface.MAC.Address == request.Mac {
			targetInterface = &iface
			break
		}
	}
	if targetInterface == nil {
		return fmt.Errorf("interface with MAC %s not found", request.Mac)
	}

	xml, err := targetInterface.Marshal()
	if err != nil {
		return err
	}

	err = lv.DomainDetachDeviceFlags(dom, xml, uint32(libvirt.DomainDeviceModifyConfig))
	if err != nil {
		return err
	}

	if state == uint8(libvirt.DomainRunning) {
		return lv.DomainDetachDeviceFlags(dom, xml, uint32(libvirt.DomainDeviceModifyLive))
	}

	return nil
}

//	func (goLibvirt *GoLibvirt) ListHostDiskWithLibvirt(ctx context.Context, request *pb.ListHostDiskRequest) (*pb.ListDiskResponse, error) {
//		return nil, nil
//	}
func (goLibvirt *GoLibvirt) ListVMDiskWithLibvirt(ctx context.Context, request *pb.ListVMDiskRequest) (*pb.ListVMDiskResponse, error) {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return nil, err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return nil, err
	}

	xmlDesc, err := lv.DomainGetXMLDesc(dom, 0)
	if err != nil {
		return nil, err
	}

	var domain libvirtxml.Domain
	if err := xml.Unmarshal([]byte(xmlDesc), &domain); err != nil {
		return nil, err
	}

	disks := goLibvirt.listVMDisks(&domain)

	return &pb.ListVMDiskResponse{Disks: disks}, nil
}

func (goLibvirt *GoLibvirt) listVMDisks(domain *libvirtxml.Domain) []*pb.VMDisk {
	disks := domain.Devices.Disks

	vmDisks := make([]*pb.VMDisk, 0, len(disks))
	for _, d := range disks {
		if d.Source.File == nil && d.Source.Block == nil {
			continue
		}

		disk := &pb.VMDisk{
			TargetDev: d.Target.Dev,
			TargetBus: d.Target.Bus,
		}
		if d.Source.File != nil {
			disk.DiskType = pb.VMDiskType_FILE
			disk.SourcePath = d.Source.File.File
		} else if d.Source.Block != nil {
			disk.DiskType = pb.VMDiskType_BLOCK
			disk.SourcePath = d.Source.Block.Dev
		}

		vmDisks = append(vmDisks, disk)
	}

	return vmDisks

}

func (goLibvirt *GoLibvirt) listHostdev(domain *libvirtxml.Domain) []*pb.VMHostdev {
	hostdevs := domain.Devices.Hostdevs

	vmHostdevs := make([]*pb.VMHostdev, 0, len(hostdevs))
	for _, d := range hostdevs {
		hostdev := &pb.VMHostdev{
			SourceAddrDomain: int32(*d.SubsysPCI.Source.Address.Domain),
			SourceAddrBus:    int32(*d.SubsysPCI.Source.Address.Bus),
			SourceAddrSlot:   int32(*d.SubsysPCI.Source.Address.Slot),
		}
		vmHostdevs = append(vmHostdevs, hostdev)
	}

	return vmHostdevs

}

func (goLibvirt *GoLibvirt) GetVMInfo(ctx context.Context, request *pb.GetVMInfoRequest) (*pb.GetVMInfoResponse, error) {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return nil, err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return nil, err
	}

	xmlDesc, err := lv.DomainGetXMLDesc(dom, 0)
	if err != nil {
		return nil, err
	}

	var domain libvirtxml.Domain
	if err := xml.Unmarshal([]byte(xmlDesc), &domain); err != nil {
		return nil, err
	}

	cpu := domain.VCPU.Value
	memory := domain.Memory.Value
	disks := goLibvirt.listVMDisks(&domain)
	interfaces := goLibvirt.listVMNetwrokInterfaces(&domain)
	hostdevs := goLibvirt.listHostdev(&domain)
	vncPort, _ := goLibvirt.getVncPort(&domain)

	return &pb.GetVMInfoResponse{Cpu: uint32(cpu), Memory: uint64(memory), Disks: disks, Interfaces: interfaces, Hostdevs: hostdevs, VncPort: int32(vncPort)}, nil
}

func (goLibvirt *GoLibvirt) AddDiskWithLibvirt(ctx context.Context, request *pb.AddDiskRequest) error {
	if request.DiskType != pb.VMDiskType_FILE && request.DiskType != pb.VMDiskType_BLOCK && request.DiskType != pb.VMDiskType_NVME {
		return fmt.Errorf("unsupport disk type %d", request.DiskType)
	}

	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return err
	}

	state, _, _, _, _, err := lv.DomainGetInfo(dom)
	if err != nil {
		return err
	}

	flags := uint32(libvirt.DomainDeviceModifyConfig)
	if state == uint8(libvirt.DomainRunning) {
		flags = uint32(libvirt.DomainDeviceModifyLive) | flags
	}

	disk := libvirtxml.DomainDisk{
		Target: &libvirtxml.DomainDiskTarget{Dev: request.TargetDev, Bus: request.TargetBus},
		Source: &libvirtxml.DomainDiskSource{},
	}

	if len(disk.Target.Bus) == 0 {
		disk.Target.Bus = "virtio"
	}

	if request.DiskType == pb.VMDiskType_FILE {
		disk.Source.File = &libvirtxml.DomainDiskSourceFile{File: request.SourcePath}
	} else if request.DiskType == pb.VMDiskType_BLOCK {
		disk.Source.Block = &libvirtxml.DomainDiskSourceBlock{Dev: request.SourcePath}
	}

	xml, err := disk.Marshal()
	if err != nil {
		return err
	}

	return lv.DomainAttachDeviceFlags(dom, xml, flags)
}
func (goLibvirt *GoLibvirt) DeleteDiskWithLibvirt(ctx context.Context, request *pb.DeleteDiskRequest) error {
	if request.DiskType != pb.VMDiskType_FILE && request.DiskType != pb.VMDiskType_BLOCK && request.DiskType != pb.VMDiskType_NVME {
		return fmt.Errorf("unsupport disk type %d", request.DiskType)
	}

	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return err
	}

	state, _, _, _, _, err := lv.DomainGetInfo(dom)
	if err != nil {
		return err
	}

	flags := uint32(libvirt.DomainDeviceModifyConfig)
	if state == uint8(libvirt.DomainRunning) {
		flags = uint32(libvirt.DomainDeviceModifyLive) | flags
	}

	xmlDesc, err := lv.DomainGetXMLDesc(dom, 0)
	if err != nil {
		return err
	}

	var domain libvirtxml.Domain
	if err := xml.Unmarshal([]byte(xmlDesc), &domain); err != nil {
		return err
	}

	disks := domain.Devices.Disks
	for _, disk := range disks {
		if disk.Target.Dev == request.TargetDev {
			xml, err := disk.Marshal()
			if err != nil {
				return err
			}
			return lv.DomainDetachDeviceFlags(dom, xml, flags)
		}
	}

	return fmt.Errorf("not found disk with target dev %s", request.TargetDev)
}

func (goLibvirt *GoLibvirt) AddHostdevWithLibvirt(ctx context.Context, request *pb.AddHostdevRequest) error {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return err
	}

	state, _, _, _, _, err := lv.DomainGetInfo(dom)
	if err != nil {
		return err
	}

	flags := uint32(libvirt.DomainDeviceModifyConfig)
	if state == uint8(libvirt.DomainRunning) {
		flags = uint32(libvirt.DomainDeviceModifyLive) | flags
	}

	addressPCIDomain := uint(0)
	addressPCIBus := uint(request.SourceAddrBus)
	addressPCISlot := uint(request.SourceAddrSlot)
	hostDev := libvirtxml.DomainHostdev{
		Managed: "yes",
		SubsysPCI: &libvirtxml.DomainHostdevSubsysPCI{
			Source: &libvirtxml.DomainHostdevSubsysPCISource{Address: &libvirtxml.DomainAddressPCI{
				Domain: &addressPCIDomain,
				Bus:    &addressPCIBus,
				Slot:   &addressPCISlot,
			}},
		},
	}

	xml, err := hostDev.Marshal()
	if err != nil {
		return err
	}

	return lv.DomainAttachDeviceFlags(dom, xml, flags)
}

func (goLibvirt *GoLibvirt) DeleteHostdevWithLibvirt(ctx context.Context, request *pb.DeleteHostdevRequest) error {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return err
	}

	state, _, _, _, _, err := lv.DomainGetInfo(dom)
	if err != nil {
		return err
	}

	// flags := uint32(libvirt.DomainDeviceModifyConfig)
	// if state == uint8(libvirt.DomainRunning) {
	// 	flags = uint32(libvirt.DomainDeviceModifyLive) | flags
	// }

	xmlDesc, err := lv.DomainGetXMLDesc(dom, 0)
	if err != nil {
		return err
	}

	var domain libvirtxml.Domain
	if err := xml.Unmarshal([]byte(xmlDesc), &domain); err != nil {
		return err
	}
	// add nvme disk
	hostDevs := domain.Devices.Hostdevs
	for _, dev := range hostDevs {
		if dev.SubsysPCI == nil || dev.SubsysPCI.Source == nil || dev.SubsysPCI.Source.Address == nil {
			continue
		}

		if dev.SubsysPCI.Source.Address.Bus == nil {
			continue
		}

		domain := dev.SubsysPCI.Source.Address.Domain
		bus := dev.SubsysPCI.Source.Address.Bus
		slot := dev.SubsysPCI.Source.Address.Slot
		if *domain != uint(request.SourceAddrDomain) ||
			*bus != uint(request.SourceAddrBus) ||
			*slot != uint(request.SourceAddrSlot) {
			continue
		}

		xml, err := dev.Marshal()
		if err != nil {
			return err
		}

		err = lv.DomainDetachDeviceFlags(dom, xml, uint32(libvirt.DomainDeviceModifyConfig))
		if err != nil {
			return err
		}

		if state == uint8(libvirt.DomainRunning) {
			return lv.DomainDetachDeviceFlags(dom, xml, uint32(libvirt.DomainDeviceModifyLive))
		}

		return nil
	}

	return fmt.Errorf("not found hostdev with domain %d, bus %d, slot %d", request.SourceAddrDomain, request.SourceAddrBus, request.SourceAddrSlot)
}

func (goLibvirt *GoLibvirt) GetVncPortWithLibvirt(ctx context.Context, request *pb.VMVncPortRequest) (*pb.VMVncPortResponse, error) {
	lv, err := goLibvirt.connectHost(request.Id)
	if err != nil {
		return nil, err
	}
	defer lv.Disconnect()

	dom, err := lv.DomainLookupByName(request.GetVmName())
	if err != nil {
		return nil, err
	}

	xmlDesc, err := lv.DomainGetXMLDesc(dom, 0)
	if err != nil {
		return nil, err
	}

	var domain libvirtxml.Domain
	if err := xml.Unmarshal([]byte(xmlDesc), &domain); err != nil {
		return nil, err
	}

	port, err := goLibvirt.getVncPort(&domain)
	if err != nil {
		return nil, err
	}

	return &pb.VMVncPortResponse{Port: int32(port)}, nil
}

func (goLibvirt *GoLibvirt) getVncPort(domain *libvirtxml.Domain) (int, error) {
	if domain.Devices == nil {
		return 0, nil
	}

	graphics := domain.Devices.Graphics
	if len(graphics) == 0 {
		return 0, nil
	}

	if graphics[0].VNC == nil {
		return 0, nil
	}

	return graphics[0].VNC.Port, nil
}

func newLibvirt(urlStr string) (*libvirt.Libvirt, error) {
	conn, resp, err := websocket.DefaultDialer.Dial(urlStr, nil)
	if err != nil {
		var msg []byte
		if resp != nil {
			msg, _ = io.ReadAll(resp.Body)
		}
		return nil, fmt.Errorf("websocket dial error: %s, msg:%s, url:%s", err.Error(), string(msg), urlStr)
	}

	l := libvirt.NewWithDialer(dialers.NewAlreadyConnected(conn.NetConn()))
	if err := l.Connect(); err != nil {
		return nil, err
	}

	return l, nil
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
		return "Stop"
	case libvirt.DomainCrashed:
		return "Crashed"
	default:
		return "Unknown"
	}
}
