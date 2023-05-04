// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/5/3

package fsruntime

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

type StackRecords []runtime.StackRecord

func (sr StackRecords) String() string {
	if len(sr) == 0 {
		return ""
	}
	m := stackRecordMap(sr)
	bs := &strings.Builder{}
	var idx int
	for _, item := range m {
		idx++
		bs.WriteString(fmt.Sprintf("stack %d [%d]:\n", idx, item.Cnt))
		printStackRecord(bs, item.Record.Stack(), false)
	}
	return strings.TrimSpace(bs.String())
}

func printStackRecord(w io.Writer, stk []uintptr, allFrames bool) {
	show := allFrames
	frames := runtime.CallersFrames(stk)
	for {
		frame, more := frames.Next()
		name := frame.Function
		if name == "" {
			show = true
			fmt.Fprintf(w, "#\t%#x\n", frame.PC)
		} else if name != "runtime.goexit" && (show || !strings.HasPrefix(name, "runtime.")) {
			// Hide runtime.goexit and any runtime functions at the beginning.
			// This is useful mainly for allocation traces.
			show = true
			fmt.Fprintf(w, "#\t%#x\t%s+%#x\t%s:%d\n", frame.PC, name, frame.PC-frame.Entry, frame.File, frame.Line)
		}
		if !more {
			break
		}
	}
	if !show {
		// We didn't print anything; do it again,
		// and this time include runtime functions.
		printStackRecord(w, stk, true)
		return
	}
	fmt.Fprint(w, "\n")
}

func GoroutineStack() StackRecords {
	for {
		num := runtime.NumGoroutine() * 2
		p := make([]runtime.StackRecord, num)
		n, ok := runtime.GoroutineProfile(p)
		if ok {
			return p[:n]
		}
	}
}

type stackRecordTmp struct {
	Record runtime.StackRecord
	Cnt    int
}

func stackRecordMap(sr []runtime.StackRecord) map[uintptr]*stackRecordTmp {
	m := make(map[uintptr]*stackRecordTmp, len(sr))
	for i := 0; i < len(sr); i++ {
		z := sr[i]
		key := z.Stack0[0]
		old, has := m[key]
		if has {
			old.Cnt++
		} else {
			m[key] = &stackRecordTmp{
				Record: z,
				Cnt:    1,
			}
		}
	}
	return m
}

func StackRecordDiff(a StackRecords, b StackRecords) StackRecordDiffResult {
	am := stackRecordMap(a)
	bm := stackRecordMap(b)

	return StackRecordDiffResult{
		More: stackRecordMore(am, bm),
		Less: stackRecordMore(bm, am),
	}
}

type StackRecordDiffResult struct {
	More StackRecords
	Less StackRecords
}

func (srd StackRecordDiffResult) HasDiff() bool {
	return len(srd.More)+len(srd.Less) > 0
}

func (srd StackRecordDiffResult) String() string {
	bs := &strings.Builder{}
	m := len(srd.More)
	l := len(srd.Less)
	bs.WriteString(fmt.Sprintf("Total Diff: %d\n", m+l))
	bs.WriteString(fmt.Sprintf("+More(%d)\n", m))
	bs.WriteString(srd.More.String())
	bs.WriteString("\n")
	bs.WriteString(fmt.Sprintf("-Less(%d)\n", l))
	bs.WriteString(srd.Less.String())
	return bs.String()
}

func stackRecordMore(am map[uintptr]*stackRecordTmp, bm map[uintptr]*stackRecordTmp) StackRecords {
	var more []runtime.StackRecord
	for k, v := range am {
		cnt := v.Cnt
		if cur, has := bm[k]; has {
			cnt = v.Cnt - cur.Cnt
		}
		for i := 0; i < cnt; i++ {
			more = append(more, v.Record)
		}
	}
	return more
}
