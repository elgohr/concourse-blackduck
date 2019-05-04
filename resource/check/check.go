package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elgohr/blackduck-resource/shared"
	"io"
	"log"
	"os"
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
	api    shared.BlackduckApi
}

func NewRunner() Runner {
	bd := shared.NewBlackduck()
	return Runner{
		stdIn:  os.Stdin,
		stdOut: os.Stdout,
		stdErr: os.Stderr,
		api:    &bd,
	}
}

func (r *Runner) run() error {
	var input shared.Request
	if err := json.NewDecoder(r.stdIn).Decode(&input); err != nil {
		return err
	}
	if !input.Source.Valid() {
		fmt.Fprintf(r.stdOut, `[]`)
		return errors.New("source is invalid")
	}
	err := r.api.Authenticate(input.Source.Url, input.Source.Username, input.Source.Password)
	if err != nil {
		return err
	}

	project, err := r.api.GetProjectByName(input.Source.GetProjectUrl(), input.Source.Name)
	if err != nil {
		fmt.Fprintf(r.stdOut, `[]`)
		return err
	}

	versions, err := r.api.GetProjectVersions(project)
	if err != nil {
		fmt.Fprintf(r.stdOut, `[]`)
		return err
	}
	b, err := json.Marshal(versions)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(r.stdOut, string(b))
	return err
}
