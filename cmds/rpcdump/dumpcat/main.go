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
)

var cid = flag.Int64("cid", -1, "filter only which conn ID")
var service = flag.String("s", "", "filter only which service")
var detail = flag.Bool("d", true, "print detail data")

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
		if filter(msg) {
			if *detail {
				fmt.Println(msg.String())
			}
			_, _ = os.Stdout.Write(msg.GetPayload())
			if *detail {
				fmt.Println()
			}
		} else {
			fmt.Println(msg.String())
		}
		return true
	})
}

func filter(msg *conndump.Message) bool {
	if *cid < 0 {
		return false
	}
	if *cid > 0 && *cid != msg.GetConnID() {
		return false
	}

	if len(*service) > 0 && *service != msg.GetService() {
		return false
	}
	return true
}
