// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/4/19

package grace

// 此文件包含所有和环境变量相关的逻辑

import (
	"os"
	"strconv"
)

const envActionKey = "FsgoGraceAction" // master 将动作传递给子进程

var envActionValue = os.Getenv(envActionKey)

const envMasterPidKey = "FsgoGraceMasterPID" // master 将自己的 pid 传给子进程

var envMasterPPIDValue = os.Getenv(envMasterPidKey)

// 创建子进程时，需要额外携带的环境变量
func envsForSubProcess() []string {
	return []string{
		envActionKey + "=" + actionSubStart,
		envMasterPidKey + "=" + pidStr,
	}
}

var (
	pidStr  = strconv.Itoa(os.Getpid())
	ppidStr = strconv.Itoa(os.Getppid())
)

func init() {
	// 将当前进程的的这两个环境变量删除掉，以避免传递给此进程派生出的孙子进程
	if envMasterPPIDValue != "" {
		_ = os.Unsetenv(envActionKey)
		_ = os.Unsetenv(envMasterPidKey)
	}
}

// IsSubProcess 判断当前进程是否是由 master 进程派生的子进程
func IsSubProcess() bool {
	// 通过检查 ppid，来判断当前进程是否直接由子进程派生出来的
	return envMasterPPIDValue == ppidStr
}
