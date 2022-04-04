// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/4/3

package fsjson

import (
	"bytes"
	"strings"

	"github.com/fsgo/fsgo/internal/number"
)

func numberSliceUnmarshalJSON[T number.Number](content []byte, zero any) ([]T, error) {
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
			value, err := number.ParseNumber[T](txt, zero)
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
	value, err := number.ParseNumber[T](string(content), zero)
	if err != nil {
		return nil, err
	}
	return []T{value}, nil
}
