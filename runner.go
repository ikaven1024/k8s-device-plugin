package deviceplugin

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

// Run keeps device plugin running, recover from error.
// sigCh receive signal to controller perform of plugin,
// True to restart, and False to exit.
func Run(config Config, sigCh <-chan bool) error {
	if err := config.Validate(); err != nil {
		return err
	}

	watcher, err := newFSWatcher(pluginapi.KubeletSocket)
	if err != nil {
		return err
	}
	defer watcher.Close()

	for {
		if restart, err := runOnce(config, watcher, sigCh); err != nil {
			return err
		} else if !restart {
			return nil
		}
	}
}

func runOnce(config Config, watcher *fsnotify.Watcher, sigCh <-chan bool) (bool, error) {
	plugin := ForConfig(config)
	if err := plugin.Start(); err != nil {
		return false, fmt.Errorf("fail to start plugin: %v", err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create {
				log.Println("Kubelet is restarted. Restart device plugin.")
				return true, nil
			}
		case err := <-watcher.Errors:
			log.Println(err)
		case sig := <-sigCh:
			if sig {
				log.Printf("Stoped by signal, will restart")
				return true, nil
			} else {
				log.Printf("Exit by signal")
				return false, nil
			}
		}
	}
}
