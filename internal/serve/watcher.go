package serve

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	w        *fsnotify.Watcher
	paths    []string
	ignore   []string
	debounce time.Duration
	trigger  chan struct{}
	stop     chan struct{}
}

func NewWatcher(paths []string, debounce time.Duration) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	watcher := &Watcher{
		w:        w,
		paths:    paths,
		debounce: debounce,
		trigger:  make(chan struct{}, 1),
		stop:     make(chan struct{}),
		ignore: []string{
			".git",
			".DS_Store",
			"public",
		},
	}

	for _, root := range paths {
		if err := watcher.addDirRecursive(root); err != nil {
			return nil, err
		}
	}

	return watcher, nil
}

func (w *Watcher) shouldIgnore(path string) bool {
	base := filepath.Base(path)

	for _, ig := range w.ignore {
		if base == ig {
			return true
		}
		if strings.Contains(path, string(filepath.Separator)+ig+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func (w *Watcher) addDirRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && !w.shouldIgnore(path) {
			return w.w.Add(path)
		}
		return nil
	})
}

func (w *Watcher) Events() <-chan struct{} {
	return w.trigger
}

func (w *Watcher) Start() {
	go func() {
		var timer *time.Timer

		for {
			select {
			case <-w.stop:
				return

			case ev, ok := <-w.w.Events:
				if !ok {
					return
				}

				if w.shouldIgnore(ev.Name) {
					continue
				}

				// Detect new directories and add them
				if ev.Op&fsnotify.Create != 0 {
					fi, err := os.Stat(ev.Name)
					if err == nil && fi.IsDir() {
						if err := w.addDirRecursive(ev.Name); err != nil {
							fmt.Fprintf(os.Stderr, "WARN: failed to watch newly created directory %s: %v\n", ev.Name, err)
						}
					}
				}

				// Normal file events
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(w.debounce, func() {
						select {
						case w.trigger <- struct{}{}:
						default:
						}
					})
				}

			case err := <-w.w.Errors:
				fmt.Fprintf(os.Stderr, "WARN: fsnotify error: %v\n", err)
			}
		}
	}()
}

func (w *Watcher) Close() error {
	close(w.stop)
	return w.w.Close()
}
