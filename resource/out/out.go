package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	runner := NewRunner()
	if err := runner.run(); err != nil {
		log.Fatalln(err)
	}
}

type Runner struct {
	stdIn  io.Reader
	stdOut io.Writer
	stdErr io.Writer
	exec   func(name string, arg ...string) *exec.Cmd
}

func NewRunner() Runner {
	return Runner{
		stdIn:  os.Stdin,
		stdOut: os.Stdout,
		stdErr: os.Stderr,
		exec:   exec.Command,
	}
}

func (r *Runner) run() error {
	var input OutRequest
	if err := json.NewDecoder(r.stdIn).Decode(&input); err != nil {
		return err
	}
	if !input.Source.Valid() {
		return errors.New("missing mandatory source field")
	}
	if !input.Params.Valid() {
		return errors.New("missing mandatory params field")
	}
	cmd := r.exec(
		"java",
		"-jar",
		"synopsys-detect-5.3.3.jar",
		"--blackduck.url="+input.Source.Url,
		"--blackduck.username="+input.Source.Username,
		"--blackduck.password="+input.Source.Password)
	cmd.Dir = input.Params.Directory
	cmd.Stdout = r.stdOut
	cmd.Stderr = r.stdErr
	fmt.Fprintf(r.stdOut, "[]")
	return cmd.Run()
}

type OutRequest struct {
	Source Source `json:"source"`
	Params Params `json:"params"`
}

type Source struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Source) Valid() bool {
	return len(s.Url) != 0 &&
		len(s.Username) != 0 &&
		len(s.Password) != 0
}

type Params struct {
	Directory string `json:"directory"`
}

func (p *Params) Valid() bool {
	return len(p.Directory) != 0
}
