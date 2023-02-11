// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/28

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fsgo/fsgo/fsfs"
)

var logName = flag.String("name", "stdout.log", "file name")
var extRule = flag.String("rotate", "1hour", "file rotate rule, allow: "+strings.Join(fsfs.RotateRuleNames(), ", "))

func init() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprint(out, "redirect stdin pipe stream to rotate file\n")
		fmt.Fprint(out, "site: github.com/fsgo/fsgo/cmds/stdlog\n")
		fmt.Fprint(out, "version: 20210828\n")
		fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprint(out, "as default, file name is like 'stdout.log.2021082819'\n")
	}
}

func main() {
	flag.Parse()
	pipeRun()
}

func pipeRun() {
	if len(*logName) == 0 {
		log.Fatal("log_name is empty")
	}
	toRotateFile(*logName, os.Stdin)
}

func toRotateFile(name string, from io.Reader) {
	f := &fsfs.Rotator{
		Path:    name,
		ExtRule: *extRule,
	}

	defer f.Close()

	if err := f.Init(); err != nil {
		log.Fatalln(err.Error())
	}
	_, _ = io.Copy(f, from)
}
