// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/4/4

package number

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseNumber[T Number](str string, zero any) (T, error) {
	str = strings.TrimSpace(str)
	switch zero.(type) {
	case int:
		num, err := strconv.Atoi(str)
		return T(num), err
	case int8:
		num, err := strconv.ParseInt(str, 10, 8)
		return T(num), err
	case int16:
		num, err := strconv.ParseInt(str, 10, 16)
		return T(num), err
	case int32:
		num, err := strconv.ParseInt(str, 10, 32)
		return T(num), err
	case int64:
		num, err := strconv.ParseInt(str, 10, 64)
		return T(num), err
	case uint:
		num, err := strconv.ParseUint(str, 10, 0)
		return T(num), err
	case uint8:
		num, err := strconv.ParseUint(str, 10, 8)
		return T(num), err
	case uint16:
		num, err := strconv.ParseUint(str, 10, 16)
		return T(num), err
	case uint32:
		num, err := strconv.ParseUint(str, 10, 32)
		return T(num), err
	case uint64:
		num, err := strconv.ParseUint(str, 10, 64)
		return T(num), err
	case float32:
		num, err := strconv.ParseFloat(str, 32)
		return T(num), err
	case float64:
		num, err := strconv.ParseFloat(str, 64)
		return T(num), err
	default:
		return 0, fmt.Errorf("unsupport number type: %T", zero)
	}
}
