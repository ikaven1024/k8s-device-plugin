package deviceplugin

import (
	"log"
	
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"github.com/fsnotify/fsnotify"
)

func Run(config Config) error {
	if err := config.Validate(); err != nil {
		return err
	}

	watcher, err := newFSWatcher(pluginapi.KubeletSocket)
	if err != nil {
		return err
	}
	defer watcher.Close()

	for {
		if err := runOnce(config, watcher); err != nil {
			return err
		}
	}
}

func runOnce(config Config, watcher *fsnotify.Watcher) error {
	plugin := ForConfig(config)
	if err := plugin.Start(); err != nil {
		log.Println(err)
		return nil
	}


	for {
		select {
		case event := <- watcher.Events:
			if event.Op & fsnotify.Create == fsnotify.Create {
				log.Println("Kubelet is restarted. Restart device plugin.")
				return nil
			}
		case err := <- watcher.Errors:
			log.Println(err)
		}

	}
}
