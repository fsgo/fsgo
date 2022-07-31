// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/7/17

package brpc

func GetErrorCode(msg *Message) int32 {
	if msg.Meta == nil {
		return ErrorCodeUnknown
	}
	resp := msg.Meta.GetResponse()
	if resp == nil {
		return ErrorCodeUnknown
	}
	return resp.GetErrorCode()
}

const (
	ErrorCodeUnknown int32 = -255
)
