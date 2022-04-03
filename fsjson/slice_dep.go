// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/4/3

package fsjson

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type float interface {
	~float32 | ~float64
}

type number interface {
	signed | unsigned | float
}

type numberType int8

const (
	numberTypeSigned = iota
	numberTypeUnsigned
	numberTypeFloat
)

func numberSliceUnmarshalJSON[T number](content []byte, dt numberType) ([]T, error) {
	if bytes.Equal(content, emptyString) || bytes.Equal(content, null) {
		return nil, nil
	}

	head := content[0]
	tail := content[len(content)-1]

	if (head == '"' && tail == '"') || (head == '[' && tail == ']') {
		values := strings.Split(string(content[1:len(content)-1]), ",")
		numbers := make([]T, 0, len(values))
		for i := 0; i < len(values); i++ {
			txt := strings.Trim(values[i], `"`)
			value, err := parseNumber[T](txt, dt)
			if err != nil {
				return nil, err
			}
			numbers = append(numbers, value)
		}
		if len(numbers) > 0 {
			return numbers, nil
		}
		return nil, nil
	}

	// 其他情况, eg:
	// {"IDS":123}
	value, err := parseNumber[T](string(content), dt)
	if err != nil {
		return nil, err
	}
	return []T{value}, nil
}

func parseNumber[T number](str string, dt numberType) (T, error) {
	str = strings.TrimSpace(str)
	switch dt {
	case numberTypeSigned:
		ret, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return 0, err
		}
		return T(ret), nil
	case numberTypeUnsigned:
		ret, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return 0, err
		}
		return T(ret), nil
	case numberTypeFloat:
		ret, err := strconv.ParseFloat(str, 10)
		if err != nil {
			return 0, err
		}
		return T(ret), nil
	}
	return 0, fmt.Errorf("unsupport number type: %v", dt)
}
