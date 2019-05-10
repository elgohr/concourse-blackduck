package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elgohr/blackduck-resource/shared"
	"io"
	"io/ioutil"
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
	var output []byte
	for _, v := range versions {
		if v.Updated.String() == input.Version.Ref {
			output, err = json.Marshal(Output{
				Ref: shared.Ref{
					Ref: v.Updated.String(),
				},
				Meta: []Meta{
					{Name: "versionName", Value: v.Name},
					{Name: "phase", Value: v.Phase},
					{Name: "settingUpdatedAt", Value: v.Updated.String()},
				},
			})
			if err != nil {
				return err
			}
			err = ioutil.WriteFile("latest_version.json", output, 0644)
			if err != nil {
				return err
			}
		}
	}

	_, err = fmt.Fprintf(r.stdOut, string(output))
	return err
}

type Output struct {
	Ref  shared.Ref `json:"version"`
	Meta []Meta     `json:"metadata"`
}

type Meta struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
