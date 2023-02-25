// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/6/26

package internal

import (
	"strings"

	"github.com/fsgo/fsgo/fsnet/fsconn/conndump"
)

// IsAction 判断输入的参数是否允许的 action
func IsAction(a string, ac conndump.MessageAction) bool {
	if len(a) == 0 {
		return true
	}
	switch ac {
	case conndump.MessageAction_Read:
		return strings.Contains(a, "r")
	case conndump.MessageAction_Write:
		return strings.Contains(a, "w")
	case conndump.MessageAction_Close:
		return strings.Contains(a, "c")
	default:
		return false
	}
}
