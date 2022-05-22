// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package fstypes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnableStatus(t *testing.T) {
	var s EnableStatus
	require.True(t, s.IsEnable())
	require.Equal(t, "enable", s.String())

	s.Disable()
	require.False(t, s.IsEnable())
	require.Equal(t, "disable", s.String())

	s.Enable()
	require.True(t, s.IsEnable())
}
