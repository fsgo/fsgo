// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/6/25

package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/fsgo/fsgo/cmds/rpcdump/internal"
	"github.com/fsgo/fsgo/fsio"
	"github.com/fsgo/fsgo/fsnet/fsconn/conndump"
)

var connID = flag.Int64("cid", 0, "filter the connID field")
var action = flag.String("a", "rwc", "filter the action filed. r: Read,w:Write,c:Close; rc: Read and Close")
var service = flag.String("s", "", "filter the service field")
var to = flag.String("to", "", "replay data to. can be a net addr, eg 127.0.0.1:8080, default to stdout")
var conc = flag.Int("conc", 1, "number of multiple requests to make at a time")

// Usage:
// cat all messages:
// dumpreplay dump.pb.202205222200
//
// cat cid=1's messages:
// dumpreplay -cid=1 dump.pb.202205222200

var w *writer

func main() {
	flag.Parse()
	w = newWriter()
	defer w.Close()

	for _, fp := range flag.Args() {
		replayFile(fp)
	}
}

func replayFile(fp string) {
	f, err := os.Open(fp)
	if err != nil {
		log.Println("open file ", fp, " failed, ", err)
		return
	}

	var wg sync.WaitGroup
	cs := &conndump.ChanScanner{
		Filter: filter,
		Receiver: func(msgs <-chan *conndump.Message) bool {
			wg.Add(1)
			go func() {
				defer wg.Done()
				replay(msgs)
			}()
			return true
		},
	}
	cs.Scan(f)
	cs.Close()

	wg.Wait()
}

func filter(msg *conndump.Message) bool {
	if *connID > 0 && *connID != msg.GetConnID() {
		return false
	}

	if !internal.IsAction(*action, msg.GetAction()) {
		return false
	}

	if len(*service) > 0 && *service != msg.GetService() {
		return false
	}
	return true
}

func replay(msgs <-chan *conndump.Message) {
	for msg := range msgs {
		w.Write(msg)
	}
}

func newWriter() *writer {
	w1 := &writer{
		Concurrency: *conc,
		To:          *to,
		buffers:     map[int64][]*conndump.Message{},
		distWriter:  map[int64]io.WriteCloser{},
	}
	if w1.Concurrency < 0 {
		w1.Concurrency = 1
	}
	return w1
}

type writer struct {
	To          string
	buffers     map[int64][]*conndump.Message
	distWriter  map[int64]io.WriteCloser
	Concurrency int
	mux         sync.RWMutex
}

func (rw *writer) Write(msg *conndump.Message) {
	rw.mux.RLock()
	cw := rw.distWriter[msg.GetConnID()]
	rw.mux.RUnlock()
	if cw != nil {
		if msg.GetAction() == conndump.MessageAction_Close {
			rw.mux.Lock()
			delete(rw.distWriter, msg.GetConnID())
			rw.mux.Unlock()
			return
		}
		_, _ = cw.Write(msg.GetPayload())
		return
	}

	rw.mux.Lock()
	cw = rw.distWriter[msg.GetConnID()]

	if cw != nil {
		if msg.GetAction() == conndump.MessageAction_Close {
			delete(rw.distWriter, msg.GetConnID())
			rw.mux.Unlock()
			_ = cw.Close()
			return
		}
		rw.mux.Unlock()
		_, _ = cw.Write(msg.GetPayload())
		return
	}

	if len(rw.distWriter) >= rw.Concurrency {
		w.buffers[msg.GetConnID()] = append(w.buffers[msg.GetConnID()], msg)
		rw.mux.Unlock()
		return
	}

	w.buffers[msg.GetConnID()] = append(w.buffers[msg.GetConnID()], msg)

	rw.sendBufferClosed()

	for k, msgs := range rw.buffers {
		cw = newConn()
		for i := 0; i < len(msgs); i++ {
			_, _ = cw.Write(msgs[i].GetPayload())
		}
		w.distWriter[msgs[0].GetConnID()] = cw
		delete(rw.buffers, k)
		if len(rw.distWriter) >= rw.Concurrency {
			break
		}
	}
	rw.mux.Unlock()
}

func (rw *writer) sendBufferClosed() {
	for k, msgs := range rw.buffers {
		if msgs[len(msgs)-1].GetAction() == conndump.MessageAction_Close {
			cw := newConn()
			for i := 0; i < len(msgs); i++ {
				_, _ = cw.Write(msgs[i].GetPayload())
			}
			_ = cw.Close()
			delete(rw.buffers, k)
		}
	}
}

func (rw *writer) Close() {
	rw.mux.Lock()
	defer rw.mux.Unlock()

	for k, msgs := range rw.buffers {
		cw := newConn()
		for i := 0; i < len(msgs); i++ {
			_, _ = cw.Write(msgs[i].GetPayload())
		}
		_ = cw.Close()
		delete(rw.buffers, k)
	}

	for _, cw := range rw.distWriter {
		_ = cw.Close()
	}
}

func newConn() io.WriteCloser {
	if len(*to) == 0 {
		return fsio.NopWriteCloser(os.Stdout)
	}

	for i := 0; ; i++ {
		c, err := net.DialTimeout("tcp", *to, 3*time.Second)
		if err != nil {
			log.Println("connect to ", *to, "failed:", err)
			if i > 2 {
				time.Sleep(100 * time.Millisecond)
			} else if i > 10 {
				time.Sleep(time.Second)
			}
			continue
		}
		return c
	}
}
