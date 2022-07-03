// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/6/26

package internal

import (
	"strconv"
	"strings"

	"github.com/fsgo/fsgo/fsnet/fsconn/conndump"
)

func FormatMessage(msg *conndump.Message, long bool) string {
	var b strings.Builder
	b.WriteString("ID:")
	b.WriteString(strconv.FormatInt(msg.GetID(), 10))
	b.WriteString(" Action:")
	b.WriteString(msg.GetAction().String())
	b.WriteString(" Service:")
	b.WriteString(msg.GetService())
	b.WriteString(" ConnID:")
	b.WriteString(strconv.FormatInt(msg.GetConnID(), 10))
	b.WriteString("_")
	b.WriteString(strconv.FormatInt(msg.GetSubID(), 10))
	b.WriteString(" Addr:")
	b.WriteString(msg.GetAddr())

	b.WriteString(" Time:")
	b.WriteString(msg.GetTime().AsTime().Local().Format("20060102 15:04:05.000"))
	b.WriteString(" Payload:(")
	p := msg.GetPayload()
	b.WriteString(strconv.Itoa(len(p)))
	b.WriteString(")")
	if long {
		b.WriteString(strconv.Quote(string(p)))
	} else {
		if len(p) > 48 {
			p = p[:48]
		}
		b.WriteString(strconv.Quote(string(p)))
	}
	return b.String()
}
