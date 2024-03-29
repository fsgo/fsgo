// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/31

package fsfs

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/fsgo/fsgo/fssync"
)

// WatchFile watch file
type WatchFile struct {
	FileName string
	Parser   func(content []byte) error

	onStop func()

	afterChanges fssync.Slice[func()]

	mux     sync.RWMutex
	started bool
}

// Start  watch start
func (wf *WatchFile) Start() error {
	wf.mux.Lock()
	defer wf.mux.Unlock()

	if wf.started {
		return errors.New("already started")
	}
	if err := wf.Load(); err != nil {
		return err
	}

	watcher := &Watcher{
		Interval: time.Second,
		Delay:    time.Second,
	}
	watcher.Watch(wf.FileName, func(event WatcherEvent) {
		_ = wf.Load()
		all := wf.afterChanges.Load()
		for _, fn := range all {
			fn()
		}
	})

	wf.onStop = watcher.Stop
	err := watcher.Start()
	if err == nil {
		wf.started = true
	}
	return err
}

// Load load file
func (wf *WatchFile) Load() error {
	if len(wf.FileName) == 0 {
		return errors.New("fileName is empty")
	}
	if wf.Parser == nil {
		return errors.New("parser func is nil")
	}
	bf, err := os.ReadFile(wf.FileName)
	if err != nil {
		return err
	}
	return wf.Parser(bf)
}

// OnFileChange register file change callback
func (wf *WatchFile) OnFileChange(fn func()) {
	wf.afterChanges.Add(fn)
}

// Stop watch stop
func (wf *WatchFile) Stop() {
	wf.mux.Lock()
	defer wf.mux.Unlock()
	if !wf.started {
		return
	}
	wf.onStop()
	wf.started = false
}
