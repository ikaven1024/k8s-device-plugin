package main

import (
	"github.com/ikaven1024/device-plugin"
	"log"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"os"
	"github.com/fsnotify/fsnotify"
	"path/filepath"
)

func main() {
	log.Println("Example")
	p := deviceplugin.New("dir", pluginapi.DevicePluginPath + "dir.sock")
	err := p.Start()
	assert(err)
	defer p.Stop()

	start(p)
}

const dirRoot = "/tmp/dir-devices"

func start(plugin deviceplugin.DevicePlugin )  {
	err := os.MkdirAll(dirRoot, os.ModeDir)
	assert(err)

	w, err := fsnotify.NewWatcher()
	assert(err)
	defer w.Close()
	w.Add(dirRoot)

	plugin.ReplaceDevice(listDevice()...)
	for {
		select {
		case e := <-w.Events:
			switch e.Op {
			case fsnotify.Create:
				plugin.AddOrUpdateDevice(&pluginapi.Device{filepath.Base(e.Name), pluginapi.Healthy})
			case fsnotify.Remove:
				plugin.RemoveDevice(&pluginapi.Device{filepath.Base(e.Name), pluginapi.Healthy})
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
