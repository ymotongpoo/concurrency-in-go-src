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
	GoVers = []string{
		"1.8.7",
		"1.9.7",
		"1.10.4",
		"1.11",
	}

	Pkgs = []string{
		"golang.org/x/time/rate",
	}

	GoRoot    = "/opt/go/go%s"
	GoBinPath = "/opt/go/go%s/bin/go"
)

func main() {
	for _, ver := range GoVers {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		// install required packages
		gopath := filepath.Join(cwd, ver)
		os.Setenv("GOPATH", gopath)
		for _, p := range Pkgs {
			cmd := exec.Command(fmt.Sprintf(GoBinPath, ver), "get", p)
			cmd = prepareExecEnv(cmd, ver, os.Stderr)
			cmd.Run()
		}

		file, err := os.Create(ver + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Fatal(err)
			}
		}()
		if err = filepath.Walk(cwd, runGofile(ver, file)); err != nil {
			log.Fatal(err)
		}
	}
}

func runGofile(ver string, out io.Writer) filepath.WalkFunc {
	binPath := fmt.Sprintf(GoBinPath, ver)
	return func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.HasSuffix(info.Name(), "_test.go") {
				fmt.Fprintf(out, "\n>>>>> %s\n", path)
				cmd := exec.Command(binPath, "test", path)
				cmd = prepareExecEnv(cmd, ver, out)
				cmd.Run()
			} else if strings.HasSuffix(info.Name(), ".go") {
				fmt.Fprintf(out, "\n>>>>> %s\n", path)
				cmd := exec.Command(binPath, "run", path)
				cmd = prepareExecEnv(cmd, ver, out)
				cmd.Run()
			}
		}
		return nil
	}
}

func prepareExecEnv(cmd *exec.Cmd, ver string, out io.Writer) *exec.Cmd {
	cmd.Stderr = out
	cmd.Stdout = out
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	gopath := filepath.Join(cwd, ver)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GOROOT", fmt.Sprintf(GoRoot, ver))
	return cmd
}
