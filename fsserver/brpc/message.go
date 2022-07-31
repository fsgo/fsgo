// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/16

package brpc

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

type Message struct {
	Meta       *Meta
	Attachment io.Reader

	body []byte
	msg  proto.Message
}

func (req *Message) WithMeta(fn func(m *Meta)) {
	if req.Meta == nil {
		req.Meta = &Meta{}
	}
	fn(req.Meta)
}

func (req *Message) ServiceName() string {
	return req.Meta.Request.GetServiceName()
}

func (req *Message) MethodName() string {
	return req.Meta.Request.GetMethodName()
}

func (req *Message) loadAttachment() ([]byte, error) {
	if req.Attachment == nil {
		return nil, nil
	}
	if b, ok := req.Attachment.(*bytes.Buffer); ok {
		return b.Bytes(), nil
	}
	bf := &bytes.Buffer{}
	size, err := io.Copy(bf, req.Attachment)
	if e := tryCloseReader(req.Attachment); e != nil {
		return nil, fmt.Errorf("close attachment failed, err=%w", e)
	}
	if err != nil {
		return nil, fmt.Errorf("read attachment failed, err=%w, readSize=%d", err, size)
	}
	return bf.Bytes(), nil
}

func (req *Message) WroteTo(w io.Writer) (int64, error) {
	if req.Meta == nil {
		return 0, errors.New("meta is nil")
	}
	if req.Meta.GetRequest() == nil && req.Meta.GetResponse() == nil {
		return 0, errors.New("meta.Message and meta.Response are nil")
	}
	if req.msg == nil {
		return 0, errors.New("should bind payload Message first")
	}
	body, err := proto.Marshal(req.msg)
	if err != nil {
		return 0, err
	}
	attachment, err := req.loadAttachment()
	if err != nil {
		return 0, err
	}
	req.Meta.AttachmentSize = int32(len(attachment))

	meta, err := proto.Marshal(req.Meta)
	if err != nil {
		return 0, fmt.Errorf("%w, Marshal meta failed: %s", ErrInvalidMeta, err.Error())
	}
	h := Header{
		MetaSize: uint32(len(meta)),
		BodySize: uint32(len(meta)+len(body)) + uint32(req.Meta.GetAttachmentSize()),
	}
	wrote, err := h.WroteTo(w)
	if err != nil {
		return wrote, fmt.Errorf("write header failed: %w, wrote=%d", err, wrote)
	}
	ws, err := w.Write(meta)
	wrote += int64(ws)
	if err != nil {
		return wrote, fmt.Errorf("write meta failed: %w, wrote=%d", err, wrote)
	}
	ws, err = w.Write(body)
	wrote += int64(ws)
	if err != nil {
		return wrote, fmt.Errorf("write body failed: %w, wrote=%d", err, wrote)
	}

	if len(attachment) > 0 {
		ws, err = w.Write(body)
		wrote += int64(ws)
		if err != nil {
			return wrote, fmt.Errorf("write attachment failed: %w, wrote=%d", err, wrote)
		}
	}

	return 0, nil
}

func (req *Message) Unmarshal(msg proto.Message) error {
	return proto.Unmarshal(req.body, msg)
}

func tryCloseReader(r io.Reader) error {
	if rc, ok := r.(io.Closer); ok {
		return rc.Close()
	}
	return nil
}
