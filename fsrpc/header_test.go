// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/25

package fsrpc

import (
	"bytes"
	"testing"

	"github.com/fsgo/fst"
)

func TestReadHeader(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		h1 := Header{
			Type:   HeaderTypeRequest,
			Length: 99,
		}
		bf := &bytes.Buffer{}
		fst.NoError(t, h1.Write(bf))
		fst.Equal(t, HeaderLen, bf.Len())

		h2, err2 := ReadHeader(bf)
		fst.NoError(t, err2)
		fst.Equal(t, h1, h2)
	})
}
