// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/3/27

package vsql

import (
	"database/sql"
	"testing"
)

func TestRegisterDB(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		RegisterDB("test", NewDBOnlyCtx(&sql.DB{}))
		db, err := sql.Open(DriverName, "test")
		if err != nil {
			t.Fatalf("has error:%v", err)
		}
		if db == nil {
			t.Fatalf("db is nil")
		}
	})
}
