// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/10/8

package fsflag

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"
)

type EnvFlags struct{}

func (e *EnvFlags) StringVar(p *string, name string, envKey string, value string, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		value = ev
	}
	usage += " ( Env Key: " + envKey + " )"
	flag.StringVar(p, name, value, usage)
}

func (e *EnvFlags) IntVar(p *int, name string, envKey string, value int, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.Atoi(ev)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += " ( Env Key: " + envKey + " )"
	flag.IntVar(p, name, value, usage)
}

func (e *EnvFlags) Int64Var(p *int64, name string, envKey string, value int64, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.ParseInt(ev, 10, 64)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += " ( Env Key: " + envKey + " )"
	flag.Int64Var(p, name, value, usage)
}

func (e *EnvFlags) Uint64Var(p *uint64, name string, envKey string, value uint64, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.ParseUint(ev, 10, 64)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += " ( Env Key: " + envKey + " )"
	flag.Uint64Var(p, name, value, usage)
}

func (e *EnvFlags) UintVar(p *uint, name string, envKey string, value uint, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.ParseUint(ev, 10, strconv.IntSize)
		if err == nil {
			value = uint(nv)
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += " ( Env Key: " + envKey + " )"
	flag.UintVar(p, name, value, usage)
}

func (e *EnvFlags) DurationVar(p *time.Duration, name string, envKey string, value time.Duration, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := time.ParseDuration(ev)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += " ( Env Key: " + envKey + " )"
	flag.DurationVar(p, name, value, usage)
}

func (e *EnvFlags) BoolVar(p *bool, name string, envKey string, value bool, usage string) {
	ev := os.Getenv(envKey)
	if ev != "" {
		nv, err := strconv.ParseBool(ev)
		if err == nil {
			value = nv
		} else {
			log.Fatalf("parser flag %q from env.%q=%q failed: %v\n", name, envKey, ev, err)
		}
	}
	usage += " ( Env Key: " + envKey + " )"
	flag.BoolVar(p, name, value, usage)
}
