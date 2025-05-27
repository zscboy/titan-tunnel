package golibvirt

const (
	transport = "raw"
	vmapi     = "libvirt"

	interfaceTypeNetwork   = "network"
	interfaceTypeDirect    = "direct"
	interfaceTypeBridge    = "bridge"
	interfaceSourceNetwork = "default"

	interfaceModelType = "virtio"

	interfaceDriverName   = "vhost"
	interfaceDriverQueues = 16

	interfaceSourceDirectModelBridge      = "bridge"
	interfaceSourceDirectModelPassthrough = "passthrough"
)
