// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/22

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fsgo/fsgo/fsnet/fsconn/conndump"

	"github.com/fsgo/fsgo/cmds/rpcdump/internal"
)

var connID = flag.Int64("cid", -1, `filter only which conn ID.
-1 : disable other conditions
0  : enable other  conditions
>0 : filter only this connID
`)
var action = flag.String("a", "rwc", "filter action. r: Read, w:Write, c:Close; rc: Read and Close")
var service = flag.String("s", "", "filter only which service")
var detail = flag.Bool("d", true, "print details")

// Usage:
// cat all messages:
// dumpcat dump.pb.202205222200
//
// cat cid=1's messages:
// dumpcat -cid=1 dump.pb.202205222200
func main() {
	flag.Parse()

	for _, fp := range flag.Args() {
		catFile(fp)
	}
}

func catFile(fp string) {
	f, err := os.Open(fp)
	if err != nil {
		log.Println("open file ", fp, " failed, ", err)
		return
	}

	conndump.Scan(f, func(msg *conndump.Message) bool {
		if *connID < 0 {
			fmt.Println(internal.FormatMessage(msg, *detail))
			return true
		}

		if !filter(msg) {
			return true
		}
		if *detail {
			fmt.Println(internal.FormatMessage(msg, true))
		}
		_, _ = os.Stdout.Write(msg.GetPayload())
		if *detail {
			fmt.Println()
		}
		return true
	})
}

func filter(msg *conndump.Message) bool {
	if *connID > 0 && *connID != msg.GetConnID() {
		return false
	}

	if !internal.IsAction(*action, msg.GetAction()) {
		return false
	}

	if len(*service) > 0 && *service != msg.GetService() {
		return false
	}
	return true
}
