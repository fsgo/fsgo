// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

//go:generate protoc --go_out=. *.proto

// Protocol 协议头，每次创建网络连接后，在发送首个 Request 或者 Response 之前，发送的的内容
var Protocol = []byte{'F', 'S', 'R', 'P', 'C'}

// HeaderLen 每次发送的数据的头部，用户描述数据的长度
//
// 消息格式为：
// |--5 Byte(Header)--|--------------Body--------------|
const HeaderLen = 9

type HeaderType uint8

const (
	HeaderTypeInvalid  HeaderType = 0
	HeaderTypeRequest  HeaderType = 1
	HeaderTypeResponse HeaderType = 2
	HeaderTypePayload  HeaderType = 3
)

func (h HeaderType) String() string {
	switch h {
	case HeaderTypeInvalid:
		return "0-invalid"
	case HeaderTypeRequest:
		return "1-request"
	case HeaderTypeResponse:
		return "2-response"
	case HeaderTypePayload:
		return "3-payload"
	default:
		return fmt.Sprintf("%d-unknown", h)
	}
}

type Header struct {
	Type   HeaderType
	Length uint32
}

func (h Header) Write(w io.Writer) error {
	b := []byte{
		byte(h.Type), // 数据类型
		0, 0, 0, 0,   // 数据长度
		0, 0, 0, 0, // 校验ma
	}
	binary.LittleEndian.PutUint32(b[1:], h.Length)
	sum := crc32.ChecksumIEEE(b[:5])
	binary.LittleEndian.PutUint32(b[5:], sum)
	_, err1 := w.Write(b)
	return err1
}

func ReadHeader(rd io.Reader) (Header, error) {
	bf := make([]byte, HeaderLen)
	_, err := io.ReadFull(rd, bf)
	if err != nil {
		return Header{}, err
	}
	got := crc32.ChecksumIEEE(bf[:5])
	want := binary.LittleEndian.Uint32(bf[5:])
	if got != want {
		return Header{}, fmt.Errorf("%w: got checksum %d, want %d", ErrInvalidHeader, got, want)
	}
	ty := HeaderType(bf[0])
	return Header{
		Type:   ty,
		Length: binary.LittleEndian.Uint32(bf[1:]),
	}, nil
}
