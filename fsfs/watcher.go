// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsfs

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsgo/fsgo/fstime"
)

const (
	// WatcherEventUpdate contains create and update ent
	WatcherEventUpdate = "update"
	// WatcherEventDelete  delete event
	WatcherEventDelete = "delete"
)

// WatcherEvent event for watcher
type WatcherEvent struct {
	FileName string
	Type     string
}

// String event desc
func (we *WatcherEvent) String() string {
	return we.FileName + " " + we.Type
}

// Watcher 文件监听
type Watcher struct {
	timer *fstime.Interval
	rules []*watchRule

	Interval time.Duration
	Delay    time.Duration
	mux      sync.RWMutex
	started  bool
}

// Watch add  file watch with callback
func (w *Watcher) Watch(name string, callback func(event *WatcherEvent)) {
	if name == "" {
		panic("name is empty")
	}
	if callback == nil {
		panic("callback is nil")
	}
	wd := &watchRule{
		Name:     name,
		CallBack: callback,
		delay:    time.Second,
	}
	if w.Delay > 0 {
		wd.delay = w.Delay
	}
	w.rules = append(w.rules, wd)
}

func (w *Watcher) getInterval() time.Duration {
	if w.Interval > 0 {
		return w.Interval
	}
	return time.Second
}

// Start ticker start async
func (w *Watcher) Start() error {
	w.mux.Lock()
	defer w.mux.Unlock()
	if w.started {
		return errors.New("already started")
	}
	w.timer = &fstime.Interval{}
	w.timer.Add(w.scan)
	w.timer.Start(w.getInterval())
	w.started = true
	return nil
}

func (w *Watcher) scan() {
	defer func() {
		if re := recover(); re != nil {
			log.Println("[fsgo] fsfs.Watcher.scan panic:", re)
		}
	}()
	for _, rule := range w.rules {
		rule.scan()
	}
}

// Stop ticker stop
func (w *Watcher) Stop() {
	w.mux.Lock()
	defer w.mux.Unlock()
	if !w.started {
		return
	}
	w.timer.Stop()
	w.started = false
}

type watchRule struct {
	CallBack func(event *WatcherEvent)
	last     map[string]time.Time
	Name     string
	delay    time.Duration
}

func (wr *watchRule) checkDelay(modTime time.Time) bool {
	return modTime.Before(time.Now().Add(-1 * wr.delay))
}

func (wr *watchRule) scan() {
	matches, err := filepath.Glob(wr.Name)
	if err != nil {
		log.Printf("[fsgo] fsfs.Watch(%q) err: %v\n", wr.Name, err)
		return
	}
	if wr.last == nil {
		wr.last = map[string]time.Time{}
	}
	nowData := map[string]time.Time{}
	for _, name := range matches {
		info, err := os.Stat(name)
		if err != nil {
			continue
		}
		lastMod, has := wr.last[name]
		// 新增 或者有变更的情况
		if !has || !info.ModTime().Equal(lastMod) {
			if wr.checkDelay(info.ModTime()) {
				nowData[name] = info.ModTime()
				event := &WatcherEvent{
					FileName: name,
					Type:     WatcherEventUpdate,
				}
				wr.CallBack(event)
			} else {
				nowData[name] = info.ModTime().Add(-1)
			}
		} else {
			// 没有变化的情况
			nowData[name] = lastMod
		}
	}
	for name := range wr.last {
		// 针对已删除的场景
		if _, has := nowData[name]; !has {
			event := &WatcherEvent{
				FileName: name,
				Type:     WatcherEventDelete,
			}
			wr.CallBack(event)
		}
	}
	wr.last = nowData
}
