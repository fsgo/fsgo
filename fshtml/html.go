// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/2

package fshtml

type HTML interface {
	HTML() ([]byte, error)
}

type Bytes []byte

func (b Bytes) HTML() ([]byte, error) {
	return b, nil
}

type String string

func (s String) HTML() ([]byte, error) {
	return []byte(s), nil
}
