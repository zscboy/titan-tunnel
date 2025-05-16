package golibvirt

import (
	"context"
	"fmt"
	"log"
	"testing"
	"titan-vm/vms/pb"

	"github.com/digitalocean/go-libvirt"
)

const serverAddr = "ws://localhost:7777/vm?uuid=b9a3a90e-2b14-11f0-884e-57cfb3f3dd63&transport=raw&vmapi=libvirt"

func TestListVM(t *testing.T) {
	l, err := newLibvirt(serverAddr)
	if err != nil {
		log.Fatalf("new Libvirt failed:%s", err.Error())
	}
	defer l.Disconnect()

	v, err := l.ConnectGetVersion()
	if err != nil {
		log.Fatalf("failed to retrieve libvirt version: %v", err)
	}
	fmt.Println("Version:", v)

	// Return both running and stopped VMs
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := l.ConnectListAllDomains(1, flags)
	if err != nil {
		log.Fatalf("failed to retrieve domains: %v", err)
	}

	fmt.Println("ID\tName\t\tUUID")
	fmt.Println("--------------------------------------------------------")
	for _, d := range domains {
		fmt.Printf("%d\t%s\t%x\n", d.ID, d.Name, d.UUID)
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
