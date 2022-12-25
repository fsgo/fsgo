// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/12/25

package fscmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/fsgo/fsgo/fstypes"
)

type FlagTypes interface {
	fstypes.Ordered | ~bool
}

type SliceFlag[T FlagTypes] struct {
	Sep    string
	Values []T
}

func (ss *SliceFlag[T]) GetSep() string {
	if ss.Sep == "" {
		return ","
	}
	return ss.Sep
}

func (ss *SliceFlag[T]) String() string {
	list := make([]string, 0, len(ss.Values))
	for _, s := range ss.Values {
		list = append(list, fmt.Sprintf("%v", s))
	}
	return strings.Join(list, ss.GetSep())
}

func (ss *SliceFlag[T]) getParserFunc() func(str string) (any, error) {
	var zero T

	var tmp1 any = zero
	switch tmp1.(type) {
	case int8:
		return func(str string) (any, error) {
			v, err := strconv.ParseInt(str, 10, 8)
			return int8(v), err
		}
	case int16:
		return func(str string) (any, error) {
			v, err := strconv.ParseInt(str, 10, 16)
			return int16(v), err
		}
	case int32:
		return func(str string) (any, error) {
			v, err := strconv.ParseInt(str, 10, 32)
			return int32(v), err
		}
	case int:
		return func(str string) (any, error) {
			return strconv.Atoi(str)
		}
	case int64:
		return func(str string) (any, error) {
			return strconv.ParseInt(str, 10, 64)
		}
	case uint8:
		return func(str string) (any, error) {
			v, err := strconv.ParseUint(str, 10, 8)
			return uint8(v), err
		}
	case uint16:
		return func(str string) (any, error) {
			v, err := strconv.ParseUint(str, 10, 16)
			return uint16(v), err
		}
	case uint:
		return func(str string) (any, error) {
			v, err := strconv.ParseUint(str, 10, 0)
			return uint(v), err
		}
	case uint32:
		return func(str string) (any, error) {
			v, err := strconv.ParseUint(str, 10, 32)
			return uint32(v), err
		}
	case uint64:
		return func(str string) (any, error) {
			return strconv.ParseUint(str, 10, 64)
		}
	case bool:
		return func(str string) (any, error) {
			return strconv.ParseBool(str)
		}
	case float32:
		return func(str string) (any, error) {
			v, err := strconv.ParseFloat(str, 32)
			return float32(v), err
		}
	case float64:
		return func(str string) (any, error) {
			return strconv.ParseFloat(str, 64)
		}
	case string:
		return func(str string) (any, error) {
			return str, nil
		}
	default:
		return func(str string) (any, error) {
			return zero, errors.New("not support type")
		}
	}
}

func (ss *SliceFlag[T]) Set(s2 string) error {
	ss.Values = nil

	parser := ss.getParserFunc()
	arr := strings.Split(s2, ss.GetSep())
	for _, s := range arr {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		v, err := parser(s)
		if err != nil {
			return err
		}
		ss.Values = append(ss.Values, v.(T))
	}
	return nil
}
