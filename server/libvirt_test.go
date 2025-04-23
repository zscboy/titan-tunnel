package server

import (
	"fmt"
	"log"
	"testing"

	"github.com/digitalocean/go-libvirt"
)

const serverAddr = "ws://localhost:8020/libvirt?uuid=cf50877e-2009-11f0-acf8-ab30429ee397"

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

	err = createVM(lv, "abc", "/home/abc/NiuLinkOS-v1.1.7-2411141913.iso", "/var/lib/libvirt/images/NiuLinkOS.qcow2", 4, 4096)
	if err != nil {
		log.Fatalf("createVm %v", err)
	}

}

func TestStartVM(t *testing.T) {
	const serverURL = "ws://localhost:8020/libvirt"
	const agentID = "cf50877e-2009-11f0-acf8-ab30429ee397"

	goLibvirt := GoLibvirt{serverURL: serverURL}

	libvirt, err := goLibvirt.newLibvirt(agentID)
	if err != nil {
		log.Fatalf("newLibvirt %v", err)
	}

	if err := goLibvirt.startVM(libvirt, "abc"); err != nil {
		log.Fatalf("startVM %v", err)
	}
}

func TestStopVM(t *testing.T) {
	const serverURL = "ws://localhost:8020/libvirt"
	const agentID = "cf50877e-2009-11f0-acf8-ab30429ee397"

	goLibvirt := GoLibvirt{serverURL: serverURL}

	libvirt, err := goLibvirt.newLibvirt(agentID)
	if err != nil {
		log.Fatalf("newLibvirt %v", err)
	}

	if err := goLibvirt.stopVM(libvirt, "abc"); err != nil {
		log.Fatalf("stopVM %v", err)
	}
}

func TestDeleteVM(t *testing.T) {
	const serverURL = "ws://localhost:8020/libvirt"
	const agentID = "cf50877e-2009-11f0-acf8-ab30429ee397"

	goLibvirt := GoLibvirt{serverURL: serverURL}

	libvirt, err := goLibvirt.newLibvirt(agentID)
	if err != nil {
		log.Fatalf("newLibvirt %v", err)
	}

	if err := goLibvirt.stopVM(libvirt, "abc"); err != nil {
		log.Fatalf("stopVM %v", err)
	}

	if err := goLibvirt.deleteVM(libvirt, "abc"); err != nil {
		log.Fatalf("deleteVM %v", err)
	}
}
