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

	project, err := r.api.GetProjectByName(input.Source)
	if err != nil {
		fmt.Fprintf(r.stdOut, `[]`)
		return err
	}

	versions, err := r.api.GetProjectVersions(input.Source, project)
	if err != nil {
		fmt.Fprintf(r.stdOut, `[]`)
		return err
	}
	var refs []shared.Ref
	for _, version := range versions {
		refs = append(refs, shared.Ref{Ref: version.Updated.String()})
	}
	b, err := json.Marshal(refs)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(r.stdOut, string(b))
	return err
}
