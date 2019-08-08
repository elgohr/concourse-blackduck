package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elgohr/concourse-blackduck/out/interpreter"
	"github.com/elgohr/concourse-blackduck/shared"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	runner := NewRunner()
	if err := runner.run(); err != nil {
		log.Fatalln(err)
	}
}

type Runner struct {
	stdIn    io.Reader
	stdOut   io.Writer
	stdErr   io.Writer
	path     string
	agentDir string
	exec     func(name string, arg ...string) *exec.Cmd
}

func NewRunner() Runner {
	return Runner{
		stdIn:    os.Stdin,
		stdOut:   os.Stdout,
		stdErr:   os.Stderr,
		path:     os.Args[1],
		agentDir: "/opt/resource",
		exec:     exec.Command,
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

	agentJar, _ := filepath.Glob(r.agentDir + "/synopsys-detect-*.jar")
	if len(agentJar) != 1 {
		return errors.New("could not find the scanner, please open an issue on Github")
	}

	cmd := r.exec("java", getArguments(agentJar[0], input)...)
	cmd.Dir = r.path + "/" + input.Params.Directory
	cmd.Stderr = r.stdErr
	buf := bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(&buf, r.stdErr)

	if err := cmd.Run(); err != nil {
		return err
	}

	outContent := buf.String()
	response, err := interpreter.NewResponse(outContent)
	b, marshErr := json.Marshal(response)
	if marshErr != nil {
		return marshErr
	}
	fmt.Fprintf(r.stdOut, string(b))
	return err
}

func getArguments(agentJar string, input shared.Request) []string {
	args := []string{
		"-jar",
		agentJar,
		"--blackduck.url=" + input.Source.Url,
		"--detect.project.name=" + input.Source.Name,
		"--blackduck.username=" + input.Source.Username,
		"--blackduck.password=" + input.Source.Password,
	}
	if input.Source.Insecure {
		args = append(args, "--blackduck.trust.cert=true")
	}
	return args
}
