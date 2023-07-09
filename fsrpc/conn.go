// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/22

package fsrpc

import (
	"bytes"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

func ReadProtocol(rd io.Reader) error {
	p := make([]byte, len(Protocol))
	_, err := io.ReadFull(rd, p)
	if err != nil {
		return err
	}
	if !bytes.Equal(p, Protocol) {
		return fmt.Errorf("%w, got=%q", ErrInvalidProtocol, p)
	}
	return nil
}

func WriteProtocol(w io.Writer) error {
	_, err := w.Write(Protocol)
	return err
}

func readProtoMessage[T proto.Message](rd io.Reader, length int, obj T) (T, error) {
	bf := make([]byte, length)
	_, err := io.ReadFull(rd, bf)
	if err != nil {
		return obj, fmt.Errorf("read %T failed: %w", obj, err)
	}
	err = proto.Unmarshal(bf, obj)
	return obj, err
}
