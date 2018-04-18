This is a gerneral kubernetes device plugin. You can realize your device plugin with it very simple.

## Features

- applicable for kuberrnetes 1.10.x
- publish devices to kubernetes
- helps to mount device to container
- restart after kubelet restart, re-build socket
- custom your behaviour before container start

## Requires

- kubernetes: v1.10.x
- go: v1.9+

## Quick Start

**Download**


```
git clone https://github.com/ikaven1024/k8s-device-plugin.git $GOPATH/src/deviceplugin
```

**Create a main.go**

```go
package main

import (
	"deviceplugin"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

func main() {
	update := make(chan []*pluginapi.Device)
	go yourDeviceManager(update)

	conf := deviceplugin.Config{
		ResourceName: "example.com/your-device",
		SocketName:   "your-device.sock",
		Update:       update,
	}
	deviceplugin.Run(conf)
}
```

**Implement your device manager**

`yourDeviceManager` is an implement of your device manager. It accept an go chan, and send all devices to it when devices is changed (added, deleted, or health change).

## Config

- **ResourceName** (Requied): The name of resource register in kubernetes. The name shall be like `yourdomin/name`, and your domain shall not be `kubernetes.io`, whick is reserved.
- **SocketName**(Requied): The socket file name. Then `/var/lib/kubelet/device-plugins/<your-socket-name>`will be created.
- **PreStartFunc**(Optional): It is called before each container start if set.
- **AllocateFunc**(Optional): Handling acclocation request.

## Use of your extended resources

See [extended resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#extended-resources) for details.

## Example

- [Directory Device](./examples/dir_device.go)