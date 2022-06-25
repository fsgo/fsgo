// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package conndump

import (
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/proto"
)

// Scan 读取 dump 文件
func Scan(rd io.Reader, fn func(msg *Message) bool) error {
	head := make([]byte, 4)
	for {
		if _, err := io.ReadFull(rd, head); err != nil {
			return err
		}
		length := binary.LittleEndian.Uint32(head)
		body := make([]byte, length)
		if _, err := io.ReadFull(rd, body); err != nil {
			return err
		}
		var msg Message
		if err := proto.Unmarshal(body, &msg); err != nil {
			return err
		}
		if !fn(&msg) {
			return nil
		}
	}
}

type ChanScanner struct {
	Filter   func(msg *Message) bool
	Receiver func(<-chan *Message) bool

	chs map[int64]chan *Message
}

func (cs *ChanScanner) Scan(rd io.Reader) error {
	if cs.chs == nil {
		cs.chs = make(map[int64]chan *Message, 128)
	}
	err := Scan(rd, func(msg *Message) bool {
		if !cs.Filter(msg) {
			return true
		}
		connID := msg.GetConnID()
		c, has := cs.chs[connID]
		if has {
			c <- msg
			if msg.GetAction() == MessageAction_Close {
				delete(cs.chs, connID)
				close(c)
			}
			return true
		}

		// 不正常的数据，没有 Read 和 Write，直接来了一个 Close，则忽略掉
		if msg.GetAction() == MessageAction_Close {
			return true
		}
		c = make(chan *Message, 128)
		c <- msg
		cs.chs[connID] = c
		return cs.Receiver(c)
	})
	return err
}

func (cs *ChanScanner) Close() {
	if len(cs.chs) == 0 {
		return
	}
	for _, c := range cs.chs {
		close(c)
	}
}
