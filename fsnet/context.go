// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/7/31

package fsnet

type ctxKey uint8

const (
	ctxKeyDialerHook ctxKey = iota
	ctxKeyResolverHook
)
