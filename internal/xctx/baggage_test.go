// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/7

package xctx

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

type tkType1 uint8

const (
	tk10 tkType1 = iota
)

type tkType2 uint8

const (
	tk20 tkType2 = iota
	tk21
)

func TestWithValues(t *testing.T) {
	d1 := &net.TCPAddr{}
	ctx1 := WithValues(context.Background(), tk10, d1)

	d2 := &net.Buffers{}
	ctx2 := WithValues(ctx1, tk20, d2)

	d3 := &net.Buffers{}
	ctx3 := WithValues(ctx2, tk21, d3)

	g1 := Values[tkType1, *net.TCPAddr](ctx3, tk10)
	require.Len(t, g1, 1)
	w1 := []*net.TCPAddr{d1}
	require.Equal(t, w1, g1)

	g2 := Values[tkType2, *net.Buffers](ctx3, tk20)
	require.Len(t, g2, 1)
	w2 := []*net.Buffers{d2}
	require.Equal(t, w2, g2)

	require.Empty(t, Values[tkType2, *net.Buffers](ctx1, tk21))
	require.Empty(t, Values[tkType2, *net.Buffers](ctx2, tk21))
}
