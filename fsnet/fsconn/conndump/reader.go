// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package conndump

import (
	"encoding/binary"
	"errors"
	"io"
	"time"

	"google.golang.org/protobuf/proto"
)

// Scan 读取 dump 文件
func Scan(rd io.Reader, fn func(msg *Message) bool) error {
	head := make([]byte, 4)
	for {
		if _, err := io.ReadFull(rd, head); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
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

// ChanScanner 将有多个连接(ConnID)的数据流按照同一个连接分组筛选输出
//
//	比如数据流里的数据是:
//	conn1.ReadData、conn1.ReadData、conn2.ReadData、conn1.ReadData、conn1.CLose、conn2.CLose
//	最终输出为：
//	数据流 1：conn1.ReadData、conn1.ReadData、conn1.ReadData、conn1.CLose
//	数据流 2：conn2.ReadData、conn2.CLose
type ChanScanner struct {
	// Filter 可选，消息过滤器，若返回 false，则这条消息忽略掉
	Filter func(msg *Message) bool

	// Receiver 必填，接收消息 channel 的回调，此 channel 内的消息是同一个网络连接内
	Receiver func(<-chan *Message) bool

	// Timeout 可选，Message 消息超时时间
	// 	若有较多不完整的消息，可以配置该参数
	// 	若一条消息有 Read 或者 Write，但是超过 Timeout 没有 Close，则将此消息的 channel 关闭
	// 	默认为 0-不检查超时情况
	Timeout time.Duration

	chs map[int64]chan *Message

	// 存储消息的超时情况，key => connID，value => Message.Time
	timeouts map[int64]int64

	loopID int64
}

func (cs *ChanScanner) doFilter(msg *Message) bool {
	if cs.Filter == nil {
		return true
	}
	return cs.Filter(msg)
}

func (cs *ChanScanner) closeMsgChanel(connID int64) {
	close(cs.chs[connID])
	delete(cs.chs, connID)
	if cs.Timeout > 0 {
		delete(cs.timeouts, connID)
	}
}

// Scan 读取数据流
func (cs *ChanScanner) Scan(rd io.Reader) error {
	if cs.chs == nil {
		cs.chs = make(map[int64]chan *Message, 128)
		cs.timeouts = make(map[int64]int64, 128)
	}
	return Scan(rd, cs.receiveMessage)
}

func (cs *ChanScanner) receiveMessage(msg *Message) bool {
	cs.loopID++
	// 每读取到 1024 条消息，检查一次过期情况
	if cs.Timeout > 0 && cs.loopID%1024 == 0 {
		cs.checkTimeout(msg.GetTime().AsTime().UnixNano())
	}
	if !cs.doFilter(msg) {
		return true
	}
	connID := msg.GetConnID()
	c, has := cs.chs[connID]
	if has {
		c <- msg
		if msg.GetAction() == MessageAction_Close {
			cs.closeMsgChanel(connID)
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

	if cs.Timeout > 0 {
		cs.timeouts[connID] = msg.GetTime().AsTime().UnixNano()
	}

	return cs.Receiver(c)
}

func (cs *ChanScanner) checkTimeout(now int64) {
	for k, v := range cs.timeouts {
		if time.Duration(now-v) > cs.Timeout {
			cs.closeMsgChanel(k)
		}
	}
}

// Close 关闭
//
//	若有不完整的未关闭的流，此时也会同意关闭掉
func (cs *ChanScanner) Close() {
	if len(cs.chs) == 0 {
		return
	}
	for _, c := range cs.chs {
		close(c)
	}
}
