// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/10/31

package envfile

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var errNotEnvKV = fmt.Errorf("not env k-v pair")

const txtEnvMax = 1024 * 1024

// ParserFile parser env file
func ParserFile(ctx context.Context, fp string) ([]string, error) {
	info, err := os.Stat(fp)
	if err != nil {
		return nil, err
	}

	//  若是文本文件，尺寸应该是比较小
	if info.Size() < txtEnvMax {
		bf, err := ioutil.ReadFile(fp)
		if err != nil {
			return nil, err
		}

		result, errParser := parserEnv(bf)
		if errParser == nil {
			return result, nil
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, fp)
	cmd.Stderr = os.Stderr
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parserEnv(data)
}

var keyReg = regexp.MustCompile("^[a-zA-Z_][a-zA-Z_0-9]*$")

func parserEnv(bf []byte) ([]string, error) {
	var result []string
	lines := bytes.Split(bf, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		// 以 # 开头的是注释
		if bytes.HasPrefix(line, []byte("#")) {
			continue
		}
		// 不包含 = 说明不是 kv 对
		if !bytes.ContainsAny(line, "=") {
			errParser := fmt.Errorf("line(%q) %w", line, errNotEnvKV)
			return nil, errParser
		}
		arr := strings.SplitN(string(line), "=", 2)
		k := strings.TrimSpace(arr[0])
		v := strings.TrimSpace(arr[1])
		if !keyReg.MatchString(k) {
			errParser := fmt.Errorf("line(%q) %w", line, errNotEnvKV)
			return nil, errParser
		}
		result = append(result, k+"="+v)
	}
	return result, nil
}
