package golibvirt

import (
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

func createInstanceXML(name, isoPath, diskPath string, vcpu uint, memoryMB uint) *libvirtxml.Domain {
	domain := &libvirtxml.Domain{
		Type: "kvm",
		Name: name,
		Metadata: &libvirtxml.DomainMetadata{
			XML: `
			<libosinfo:libosinfo xmlns:libosinfo="http://libosinfo.org/xmlns/libvirt/domain/1.0">
			</libosinfo:libosinfo>`,
			// <libosinfo:os id="http://centos.org/centos/7.0"/>
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

	controller := libvirtxml.DomainController{Type: "pci", Model: "pcie-root", Index: uintPtr(0)}
	domain.Devices.Controllers = append(domain.Devices.Controllers, controller)
	for i := 1; i < 15; i++ {
		controller := libvirtxml.DomainController{Type: "pci", Model: "pcie-root-port", Index: uintPtr(uint(i))}
		domain.Devices.Controllers = append(domain.Devices.Controllers, controller)
	}

	return domain
}

func uintPtr(v uint) *uint {
	return &v
}
