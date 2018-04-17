package deviceplugin

import pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"

type DevicePlugin interface {
	SetPreStartFunc(f PreStartFunc)
	SetAllocateFunc(f AllocateFunc)

	Start() error
	Stop() error

	AddOrUpdateDevice(devs ...*pluginapi.Device)
	RemoveDevice(devs ...*pluginapi.Device)
	ReplaceDevice(devs ...*pluginapi.Device)
}
