package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elgohr/blackduck-resource/out/interpreter"
	"github.com/elgohr/blackduck-resource/shared"
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
	var input shared.Request
	if err := json.NewDecoder(r.stdIn).Decode(&input); err != nil {
		return err
	}
	if !input.Source.Valid() {
		return errors.New("missing mandatory source field")
	}
	if !input.Params.Valid() {
		return errors.New("missing mandatory params field")
	}
	var cmd *exec.Cmd
	if len(input.Source.Token) != 0 {
		cmd = r.exec(
			"java",
			"-jar",
			"/opt/resource/synopsys-detect-5.3.3.jar",
			"--blackduck.url="+input.Source.Url,
			"--blackduck.api.token="+input.Source.Token,
			"--detect.project.name="+input.Source.Name,
			"--blackduck.trust.cert=true")
	} else {
		cmd = r.exec(
			"java",
			"-jar",
			"/opt/resource/synopsys-detect-5.3.3.jar",
			"--blackduck.url="+input.Source.Url,
			"--blackduck.username="+input.Source.Username,
			"--blackduck.password="+input.Source.Password,
			"--detect.project.name="+input.Source.Name,
			"--blackduck.trust.cert=true")
	}

	cmd.Dir = input.Params.Directory
	cmd.Stderr = r.stdErr
	buf := bytes.Buffer{}
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return err
	}

	response, err := interpreter.NewResponse(buf.String())
	b, marshErr := json.Marshal(response)
	if marshErr != nil {
		return marshErr
	}
	fmt.Fprintf(r.stdOut, string(b))
	return err
}
