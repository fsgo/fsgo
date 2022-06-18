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

var gid = flag.Int64("gid", 0, "filter only which gid")
var service = flag.String("s", "", "filter only which service")

// Usage:
// cat all messages:
// dumpcat dump.pb.202205222200
//
// cat gid=1's messages:
// dumpcat -gid=1 dump.pb.202205222200
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
			// 满足筛选条件的则只输出消息体，可直接将内容发送给 server 用于重放
			_, _ = os.Stdout.Write(msg.GetPayload())
		} else {
			fmt.Println(msg.String())
		}
		return true
	})
}

func filter(msg *conndump.Message) bool {
	if *gid < 0 {
		return false
	}
	if *gid > 0 && *gid != msg.GetGID() {
		return false
	}

	if len(*service) > 0 && *service != msg.GetService() {
		return false
	}
	return true
}
