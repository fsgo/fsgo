// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/8/5

package brpc_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fsgo/fsgo/fsserver/brpc"
)

func TestHeader_WroteTo(t *testing.T) {
	for i := 0; i < 100; i++ {
		h := brpc.Header{
			BodySize: uint32(i),
			MetaSize: uint32(i) + 100,
		}
		bf := &bytes.Buffer{}
		n, err := h.WroteTo(bf)
		require.NoError(t, err)
		require.Equal(t, int64(12), n)
		require.NotEmpty(t, bf.String())

		h1, err := brpc.ReadHeader(bf)
		require.NoError(t, err)
		require.Equal(t, h, h1)
	}
}
