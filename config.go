package deviceplugin

import (
	"fmt"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/apis/core/v1/helper"
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
	_ = v1.ResourceName(c.ResourceName)

	if !helper.IsExtendedResourceName(v1.ResourceName(c.ResourceName)) {
		return fmt.Errorf("%v is not a valid resource name", c.ResourceName)
	}

	if len(c.ResourceName) == 0 {
		return fmt.Errorf("resource name cannot be empty")
	}

	if len(c.SocketName) == 0 {
		return fmt.Errorf("socket cannot be empty")
	}

	return nil
}
