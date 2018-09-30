//    Copyright 2018 Yoshi Yamaguchi
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ExecFile = "walktest.go"

	GoVers = []string{
		"1.8.7",
		"1.9.7",
		"1.10.4",
		"1.11",
	}

	Pkgs = []string{
		"golang.org/x/time/rate",
	}

	GoRoot     = "/opt/go/go%s"
	GoBinPath  = "/opt/go/go%s/bin/go"
	GoPathRoot string
)

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	GoPathRoot = filepath.Join(cwd, "gopath")
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// install required packages per Go version
	for _, ver := range GoVers {
		for _, p := range Pkgs {
			cmd := exec.Command(fmt.Sprintf(GoBinPath, ver), "get", p)
			cmd = prepareExecEnv(cmd, ver, os.Stderr)
			cmd.Run()
		}
	}

	if err := filepath.Walk(cwd, runGofile()); err != nil {
		log.Fatal(err)
	}
}

func runGofile() filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(path, GoPathRoot) {
			return nil
		}
		if !info.IsDir() {
			if strings.HasSuffix(info.Name(), "_test.go") {
				return runPerVer("test", path, info)
			} else if info.Name() != ExecFile && strings.HasSuffix(info.Name(), ".go") {
				return runPerVer("run", path, info)
			}
		}
		return nil
	}
}

func runPerVer(option string, path string, info os.FileInfo) error {
	file, err := os.Create(info.Name() + ".txt")
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	for _, ver := range GoVers {
		fmt.Fprintf(file, "\n>>>>> %s\n", ver)
		cmd := exec.Command(fmt.Sprintf(GoBinPath, ver), option, path)
		cmd = prepareExecEnv(cmd, ver, file)
		cmd.Run()
	}
	return nil
}

func prepareExecEnv(cmd *exec.Cmd, ver string, out io.Writer) *exec.Cmd {
	cmd.Stderr = out
	cmd.Stdout = out
	gopath := filepath.Join(GoPathRoot, ver)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GOROOT", fmt.Sprintf(GoRoot, ver))
	return cmd
}
