// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/4

package fsflag

import (
	"errors"
	"strconv"

	"github.com/fsgo/fsgo/fstypes"
)

type Types interface {
	fstypes.Ordered | ~bool
}

func getParserFunc(zero any) func(str string) (any, error) {
	switch zero.(type) {
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
