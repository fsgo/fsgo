// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/4

package fsflag

import (
	"fmt"
	"strconv"
	"time"
)

type Duration int64

func (d *Duration) String() string {
	return time.Duration(*d).String()
}

func (d *Duration) Set(s string) error {
	*d = Duration(0)
	ds, err1 := time.ParseDuration(s)
	if err1 == nil {
		*d = Duration(ds)
		return nil
	}
	da, err2 := strconv.Atoi(s)
	if err2 != nil || da < 0 {
		err3 := fmt.Errorf("expect a duration like '300ms' or a int value like '100' as '100ms', but got %q", s)
		return err3
	}
	*d = Duration(time.Duration(da) * time.Millisecond)
	return nil
}

func (d *Duration) Duration() time.Duration {
	return time.Duration(*d)
}
