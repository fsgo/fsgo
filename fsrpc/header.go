// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"encoding/binary"
	"io"
)

//go:generate protoc --go_out=. meta.proto

// Protocol 协议头，每次创建网络连接后，在发送首个 Request 或者 Response 之前，发送的的内容
var Protocol = []byte{'F', 'S', 'R', 'P', 'C'}

// HeaderLen 每次发送的数据的头部，用户描述数据的长度
//
// 消息格式为：
// |--5 Byte(Header)--|--------------Body--------------|
const HeaderLen = 5

type HeaderType uint8

const (
	HeaderTypeInvalid  HeaderType = 0
	HeaderTypeRequest  HeaderType = 1
	HeaderTypeResponse HeaderType = 2
	HeaderTypePayload  HeaderType = 3

	HeaderTypeMin = HeaderTypeRequest
	HeaderTypeMax = HeaderTypePayload
)

type Header struct {
	Type   HeaderType
	Length uint32
}

func (h Header) Write(w io.Writer) error {
	_, err := w.Write([]byte{byte(h.Type)})
	if err != nil {
		return err
	}
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, h.Length)
	_, err1 := w.Write(b)
	return err1
}

func ReadHeader(rd io.Reader) (Header, error) {
	bf := make([]byte, HeaderLen)
	_, err := io.ReadFull(rd, bf)
	if err != nil {
		return Header{}, err
	}
	ty := HeaderType(bf[0])
	return Header{
		Type:   ty,
		Length: binary.LittleEndian.Uint32(bf[1:]),
	}, nil
}
