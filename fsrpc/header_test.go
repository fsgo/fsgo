// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/25

package fsrpc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadHeader(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		h1 := Header{
			Type:   HeaderTypeRequest,
			Length: 99,
		}
		bf := &bytes.Buffer{}
		require.NoError(t, h1.Write(bf))
		require.Equal(t, HeaderLen, bf.Len())

		h2, err2 := ReadHeader(bf)
		require.NoError(t, err2)
		require.Equal(t, h1, h2)
	})
}
