// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/7/9

package sugar

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTernary(t *testing.T) {
	require.Equal(t, 1, Ternary(true, 1, 2))
	require.Equal(t, 2, Ternary(false, 1, 2))
}
