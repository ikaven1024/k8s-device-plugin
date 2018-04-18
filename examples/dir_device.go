package main

/**
  Directories under `/tmp/dir-devices` is regarded as a kind of device, allocating to container.
*/

import (
	"log"
	"os"
	"path/filepath"

	"deviceplugin"

	"github.com/fsnotify/fsnotify"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

func main() {
	update := make(chan []*pluginapi.Device)
	go dirManager(update)

	log.Println("Example")
	conf := deviceplugin.Config{
		ResourceName: "example.com/dir",
		SocketName:   "dir.sock",
		Update:       update,

		AllocateFunc: func(ids []string) (*pluginapi.ContainerAllocateResponse, error) {
			resp := &pluginapi.ContainerAllocateResponse{}
			for _, id := range ids {
				resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
					ContainerPath: "/tmp/dir/" + id,
					HostPath:      "/tmp/dir/" + id,
				})
			}
			return resp, nil
		},
	}

	deviceplugin.Run(conf, nil)
}

const dirRoot = "/tmp/dir-devices"

func dirManager(update chan<- []*pluginapi.Device) {
	err := os.MkdirAll(dirRoot, os.ModeDir)
	assert(err)

	w, err := fsnotify.NewWatcher()
	assert(err)
	defer w.Close()
	w.Add(dirRoot)

	update <- listDevice()

	for {
		select {
		case e := <-w.Events:
			if e.Op&fsnotify.Create > 0 || e.Op&fsnotify.Remove > 0 || e.Op&fsnotify.Rename > 0 {
				update <- listDevice()
			}
		case err = <-w.Errors:
			log.Println("Error", err)
		}
	}
}

func listDevice() []*pluginapi.Device {
	var devs []*pluginapi.Device
	filepath.Walk(dirRoot, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			devs = append(devs, &pluginapi.Device{info.Name(), pluginapi.Healthy})
		}
		return nil
	})
	return devs
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}
