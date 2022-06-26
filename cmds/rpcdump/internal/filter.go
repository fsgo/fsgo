// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/6/26

package internal

import (
	"github.com/fsgo/fsgo/fsnet/fsconn/conndump"
)

// IsAction 判断输入的参数是否允许的 action
func IsAction(a string, ac conndump.MessageAction) bool {
	if len(a) == 0 {
		return true
	}
	for i := 0; i < len(a); i++ {
		switch a[i] {
		case 'r':
			return ac == conndump.MessageAction_Read
		case 'w':
			return ac == conndump.MessageAction_Write
		case 'c':
			return ac == conndump.MessageAction_Close
		}
	}
	return false
}
