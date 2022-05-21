// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/21

package number

import (
	"fmt"
)

// Checker 用于测试的时候，进行计数检查
type Checker int64

// Want 对比测试，失败则返回 error
func (tn *Checker) Want(want int) error {
	if int(*tn) != want {
		return fmt.Errorf("not samm, number=%d want=%d", tn, want)
	}
	return nil
}

// Inc 计数+1
func (tn *Checker) Inc() {
	*tn++
}
