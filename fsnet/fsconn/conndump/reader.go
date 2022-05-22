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
