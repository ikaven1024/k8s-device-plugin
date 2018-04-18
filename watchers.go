package deviceplugin

import (
	"os"
	"os/signal"

	"github.com/fsnotify/fsnotify"
)

func newFSWatcher(files ...string) (*fsnotify.Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if err = w.Add(f); err != nil {
			return nil, err
		}
	}
	return w, nil
}

func newOSWatcher(sigs ...os.Signal) chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	return ch
}
