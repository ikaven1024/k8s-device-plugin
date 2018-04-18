package deviceplugin

import (
	"fmt"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

type Config struct {
	//+ required
	ResourceName string
	SocketName   string
	Update       <-chan []*pluginapi.Device

	//+ optional
	PreStartFunc PreStartFunc
	AllocateFunc AllocateFunc
}

func (c *Config) Validate() error {
	if len(c.ResourceName) == 0 {
		return fmt.Errorf("resource name cannot be empty")
	}

	if len(c.SocketName) == 0 {
		return fmt.Errorf("socket cannot be empty")
	}

	return nil
}
