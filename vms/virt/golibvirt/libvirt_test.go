package golibvirt

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"
	"titan-vm/vms/internal/config"
	"titan-vm/vms/pb"

	"github.com/digitalocean/go-libvirt"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
	"github.com/zeromicro/go-zero/core/conf"
)

const serverAddr = "ws://localhost:7777/vm?uuid=b9a3a90e-2b14-11f0-884e-57cfb3f3dd63&transport=raw&vmapi=libvirt"
const configFile = "../../etc/vms.yaml"

func loadConfig() config.Config {
	var c config.Config
	conf.MustLoad(configFile, &c)
	return c
}

func TestListVM(t *testing.T) {
	const serverURL = "ws://localhost:7777/api/ws/vm"
	const hostID = "30ad25ce-3610-11f0-8a84-8b2a4d0acbcf"

	// goLibvirt := GoLibvirt{serverURL: serverURL}

	goLibvirt := NewGoLibvirt(serverURL)
	rsp, err := goLibvirt.ListVMInstance(context.Background(), &pb.ListVMInstanceReqeust{
		Id: hostID,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	for _, vm := range rsp.VmInfos {
		t.Logf("name:%s, state:%s", vm.Name, vm.State)
	}
}

func TestDomainListNetwork(t *testing.T) {
	const serverURL = "ws://localhost:7777/api/ws/vm"
	const hostID = "b9a3a90e-2b14-11f0-884e-57cfb3f3dd63"

	// goLibvirt := GoLibvirt{serverURL: serverURL}

	goLibvirt := NewGoLibvirt(serverURL)
	rsp, err := goLibvirt.ListVMNetwrokInterfaceWithLibvirt(context.Background(), &pb.ListVMNetwrokInterfaceReqeust{
		Id:     hostID,
		VmName: "Test",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	for _, iface := range rsp.Interfaces {
		t.Logf("Name:%s, Type:%s, source:%s, source model:%s model:%s, mac:%s", iface.Name, iface.Type, iface.Source, iface.SourceModel, iface.Model, iface.Mac)
	}
}

func TestHostListNetwork(t *testing.T) {
	const serverURL = "ws://localhost:7777/vm"
	const hostID = "b9a3a90e-2b14-11f0-884e-57cfb3f3dd63"

	// goLibvirt := GoLibvirt{serverURL: serverURL}

	goLibvirt := NewGoLibvirt(serverURL)
	rsp, err := goLibvirt.ListHostNetworkInterfaceWithLibvirt(context.Background(), &pb.ListHostNetworkInterfaceRequest{
		Id: hostID,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	for _, iface := range rsp.Interfaces {
		t.Logf("Name:%s, mac:%s, active:%v", iface.Name, iface.Mac, iface.Active)
	}
}

func TestDomainAddInterface(t *testing.T) {
	const serverURL = "ws://localhost:7777/vm"
	const hostID = "b9a3a90e-2b14-11f0-884e-57cfb3f3dd63"

	// goLibvirt := GoLibvirt{serverURL: serverURL}

	goLibvirt := NewGoLibvirt(serverURL)

	err := goLibvirt.AddNetworkInterfaceWithLibvirt(context.Background(), &pb.AddNetworkInterfaceRequest{
		Id:              hostID,
		VmName:          "Test",
		Type:            pb.InterfaceType_INTERFACE_TYPE_DIRECT,
		SourceDirectDev: "virbr0",
		Model:           pb.InterfaceSourceDirectModel_INTERFACE_SOURCE_DIRECT_MODEL_PASSTHROUGH,
	})
	if err != nil {
		t.Log(err.Error())
	}

}

func TestDomainDeleteInterface(t *testing.T) {
	const serverURL = "ws://localhost:7777/api/ws/vm"
	const hostID = "30ad25ce-3610-11f0-8a84-8b2a4d0acbcf"

	// goLibvirt := GoLibvirt{serverURL: serverURL}

	goLibvirt := NewGoLibvirt(serverURL)

	err := goLibvirt.DeleteNetworkInterfaceWithLibvirt(context.Background(), &pb.DeleteNetworkInterfaceRequest{
		Id:     hostID,
		VmName: "test",
		Mac:    "52:54:00:9d:8f:f4",
	})
	if err != nil {
		t.Log(err.Error())
	}

}

func TestInterfaceCreate(t *testing.T) {
	newInterface := libvirtxml.DomainInterface{
		Model: &libvirtxml.DomainInterfaceModel{Type: "virtio"},
		// Source: &libvirtxml.DomainInterfaceSource{Network: &libvirtxml.DomainInterfaceSourceNetwork{Network: "default"}},
		Source: &libvirtxml.DomainInterfaceSource{Direct: &libvirtxml.DomainInterfaceSourceDirect{Dev: "eth0", Mode: "bridge"}},
	}
	xml, err := newInterface.Marshal()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("xml:%s", xml)
}

func TestListHostNetwork(t *testing.T) {
	const serverURL = "ws://localhost:7777/vm"
	const hostID = "b9a3a90e-2b14-11f0-884e-57cfb3f3dd63"

	// goLibvirt := GoLibvirt{serverURL: serverURL}

	goLibvirt := NewGoLibvirt(serverURL)
	lv, err := goLibvirt.connectHost(hostID)
	if err != nil {
		t.Fatalf("connect %s", err.Error())
	}

	interfaces, _, err := lv.ConnectListAllInterfaces(1, libvirt.ConnectListInterfacesInactive)
	if err != nil {
		t.Fatalf("connect %s", err.Error())
	}

	for _, iface := range interfaces {
		t.Logf("name:%s, mac:%s", iface.Name, iface.Mac)
	}

}

func traverseStruct(v reflect.Value, prefix string) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			fmt.Printf("%s(nil)\n", prefix)
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		fmt.Printf("%s: %v\n", prefix, v.Interface())
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		fieldName := fieldType.Name
		tag := fieldType.Tag.Get("xml")

		// 拼接路径（层级）
		path := prefix
		if path != "" {
			path += "."
		}
		path += fieldName

		// 输出基本类型或递归
		if field.Kind() == reflect.Struct || field.Kind() == reflect.Ptr {
			traverseStruct(field, path)
		} else {
			fmt.Printf("%s (%s, tag=%q): %v\n", path, field.Type(), tag, field.Interface())
		}
	}
}

func TestCreateVM(t *testing.T) {
	lv, err := newLibvirt(serverAddr)
	if err != nil {
		log.Fatalf("new Libvirt failed:%s", err.Error())
	}
	defer lv.Disconnect()

	domain := createInstanceXML("abc", "/root/os/NiuLinkOS-v1.1.7-2411141913.iso", "/var/lib/libvirt/images/abc.qcow2", 4, 4096)
	if err != nil {
		log.Fatalf("createVm %v", err)
	}
	xml, err := domain.Marshal()
	if err != nil {
		log.Fatalf("Marshal %v", err)
	}

	dom, err := lv.DomainDefineXML(xml)
	if err != nil {
		log.Fatalf("DomainDefineXML %v", err)
	}

	err = lv.DomainCreate(dom)
	if err != nil {
		log.Fatalf("DomainCreate %v", err)
	}
	t.Logf("create domain %s, success", dom.Name)

}

func TestStartVM(t *testing.T) {
	const serverURL = "ws://localhost:8020/libvirt"
	const hostID = "cf50877e-2009-11f0-acf8-ab30429ee397"

	// goLibvirt := GoLibvirt{serverURL: serverURL}

	goLibvirt := NewGoLibvirt(serverURL)
	if err := goLibvirt.StartVM(context.Background(), &pb.StartVMRequest{Id: hostID, VmName: "abc"}); err != nil {
		log.Fatalf("startVM %v", err)
	}
}

func TestStopVM(t *testing.T) {
	const serverURL = "ws://localhost:8020/libvirt"
	const hostID = "cf50877e-2009-11f0-acf8-ab30429ee397"

	// goLibvirt := GoLibvirt{serverURL: serverURL}

	goLibvirt := NewGoLibvirt(serverURL)
	goLibvirt.StopVM(context.Background(), &pb.StopVMRequest{Id: hostID, VmName: "abc"})
}

func TestDeleteVM(t *testing.T) {
	const serverURL = "ws://localhost:8020/libvirt"
	const hostID = "cf50877e-2009-11f0-acf8-ab30429ee397"

	goLibvirt := NewGoLibvirt(serverURL)
	goLibvirt.StopVM(context.Background(), &pb.StopVMRequest{Id: hostID, VmName: "abc"})
	goLibvirt.DeleteVM(context.Background(), &pb.DeleteVMRequest{Id: hostID, VmName: "abc"})
}

func TestGetNodeInfo(t *testing.T) {
	const serverURL = "ws://localhost:8020/libvirt"
	const hostID = "cf50877e-2009-11f0-acf8-ab30429ee397"

	goLibvirt := NewGoLibvirt(serverURL)

	lv, err := goLibvirt.connectHost(hostID)
	if err != nil {
		t.Fatalf("connect host %s failed:%s", hostID, err.Error())
	}

	hostname, err := lv.ConnectGetHostname()
	if err != nil {
		log.Fatalf("get host name failed: %v", err)
	}

	version, err := lv.ConnectGetVersion()
	if err != nil {
		log.Fatalf("get host name failed: %v", err)
	}

	rModel, rMemory, rCpus, rMhz, rNodes, rSockets, rCores, rThreads, err := lv.NodeGetInfo()
	if err != nil {
		log.Fatalf("get node info failed: %v", err)
	}

	fmt.Printf("宿主机名: %s\n", hostname)
	fmt.Printf("Libvirt版本: %d\n", version)
	fmt.Printf("CPU模型: %v\n", rModel)
	fmt.Printf("CPU核心数: %d\n", rCpus)
	fmt.Printf("内存总量: %d\n", rMemory)
	fmt.Printf("cpu 频率: %d\n", rMhz)
	fmt.Printf("rNodes: %d\n", rNodes)
	fmt.Printf("rSockets: %d\n", rSockets)
	fmt.Printf("rCores: %d\n", rCores)
	fmt.Printf("rThreads: %d\n", rThreads)
	// lv.NodeListDevices()

}

func TestCreateVol(t *testing.T) {
	const serverURL = "ws://localhost:7777/vm"
	const hostID = "b9a3a90e-2b14-11f0-884e-57cfb3f3dd63"

	goLibvirt := NewGoLibvirt(serverURL)

	request := &pb.CreateVolWithLibvirtReqeust{Id: hostID, Name: "test-2.qcow2", Pool: "images", Capacity: 100, Format: "qcow2"}
	rVol, err := goLibvirt.CreateVolWithLibvirt(context.Background(), request)
	if err != nil {
		t.Fatalf("CreateVolWithLibvirt failed:%s", err.Error())
	}

	t.Logf("vol %#v", rVol)

}

func TestGetDefaultPool(t *testing.T) {
	const serverURL = "ws://localhost:7777/vm"
	const hostID = "b9a3a90e-2b14-11f0-884e-57cfb3f3dd63"

	goLibvirt := NewGoLibvirt(serverURL)

	lv, err := goLibvirt.connectHost(hostID)
	if err != nil {
		t.Fatalf("connect host %s failed:%s", hostID, err.Error())
	}

	storagePool, err := lv.StoragePoolLookupByName("default")
	if err != nil {
		t.Fatalf("lookup storage pool defulat failed:%s", err.Error())
	}

	t.Logf("vol %#v", storagePool)

}

func TestListVMDisk(t *testing.T) {
	const serverURL = "ws://localhost:7777/api/ws/vm"
	// const hostID = "b9a3a90e-2b14-11f0-884e-57cfb3f3dd63"
	const hostID = "30ad25ce-3610-11f0-8a84-8b2a4d0acbcf"

	// goLibvirt := GoLibvirt{serverURL: serverURL}

	goLibvirt := NewGoLibvirt(serverURL)
	rsp, err := goLibvirt.ListVMDiskWithLibvirt(context.Background(), &pb.ListVMDiskRequest{Id: hostID, VmName: "centos"})
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, disk := range rsp.Disks {
		t.Logf("diskType:%s, sourcePath:%s, targetDev:%s, targetBus:%s",
			disk.DiskType, disk.SourcePath, disk.TargetDev, disk.TargetBus)
	}

}

func TestVMAddDisk(t *testing.T) {
	const serverURL = "ws://localhost:7777/vm"
	// const hostID = "b9a3a90e-2b14-11f0-884e-57cfb3f3dd63"
	const hostID = "30ad25ce-3610-11f0-8a84-8b2a4d0acbcf"

	// goLibvirt := GoLibvirt{serverURL: serverURL}
	request := &pb.AddDiskRequest{
		Id:               hostID,
		VmName:           "centos",
		DiskType:         pb.VMDiskType_NVME,
		SourcePciAddrBus: 0x0b,
	}

	goLibvirt := NewGoLibvirt(serverURL)
	err := goLibvirt.AddDiskWithLibvirt(context.Background(), request)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("add disk success")

}

func TestVMDeleteDisk(t *testing.T) {
	const serverURL = "ws://localhost:7777/vm"
	// const hostID = "b9a3a90e-2b14-11f0-884e-57cfb3f3dd63"
	const hostID = "30ad25ce-3610-11f0-8a84-8b2a4d0acbcf"

	// goLibvirt := GoLibvirt{serverURL: serverURL}
	request := &pb.DeleteDiskRequest{
		Id:               hostID,
		VmName:           "centos",
		DiskType:         pb.VMDiskType_NVME,
		SourcePciAddrBus: 0x3,
	}

	goLibvirt := NewGoLibvirt(serverURL)
	err := goLibvirt.DeleteDiskWithLibvirt(context.Background(), request)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("delete disk success")

}

func TestCreateTemplate(t *testing.T) {
	domain := createInstanceXML("test", "/var/lib/libvirt/image/test.iso", "/var/lib/libvirt/image/test.qrcow2", 4, 8092)

	xml, err := domain.Marshal()
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Logf("xml:%s", xml)

}
