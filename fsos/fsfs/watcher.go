// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/30

package fsfs

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Watcher 文件监听
type Watcher struct {
	Interval time.Duration
	Delay    time.Duration
	rules    []*watchRule
	tk       *time.Ticker
}

const (
	WatcherEventUpdate = "update"
	WatcherEventDelete = "delete"
)

// WatcherEvent event for watcher
type WatcherEvent struct {
	FileName string
	Type     string
}

func (we *WatcherEvent) String() string {
	return we.FileName + " " + we.Type
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

func (w *Watcher) Start() error {
	if w.tk != nil {
		return errors.New("already started")
	}
	w.tk = time.NewTicker(w.getInterval())
	go func() {
		for {
			select {
			case _, ok := <-w.tk.C:
				if !ok {
					return
				}
				w.scan()
			}
		}
	}()
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

func (w *Watcher) Stop() {
	if w.tk == nil {
		return
	}
	w.tk.Stop()
	w.tk = nil
}

type watchRule struct {
	Name     string
	CallBack func(event *WatcherEvent)
	last     map[string]os.FileInfo
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
		wr.last = map[string]os.FileInfo{}
	}
	nowData := map[string]os.FileInfo{}
	for _, name := range matches {
		info, err := os.Stat(name)
		if err != nil {
			continue
		}
		lastInfo, has := wr.last[name]
		// 新增 或者有变更的情况
		if !has || !info.ModTime().Equal(lastInfo.ModTime()) {
			if wr.checkDelay(info.ModTime()) {
				nowData[name] = info
				event := &WatcherEvent{
					FileName: name,
					Type:     WatcherEventUpdate,
				}
				wr.CallBack(event)
			}
		} else {
			// 没有变化的情况
			nowData[name] = info
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
