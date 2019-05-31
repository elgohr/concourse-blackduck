package main

import (
	"encoding/json"
	"fmt"
	"github.com/elgohr/concourse-blackduck/shared"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	runner := NewRunner()
	if err := runner.run(); err != nil {
		fmt.Fprintf(runner.stdOut, `{}`)
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
		return errors.Wrap(err, "Decode")
	}
	if !input.Source.Valid() {
		return errors.New("source is invalid")
	}

	project, err := r.api.GetProjectByName(input.Source)
	if err != nil {
		return errors.Wrap(err, "GetProjectByName")
	}

	versions, err := r.api.GetProjectVersions(input.Source, project)
	if err != nil {
		return errors.Wrap(err, "GetProjectVersions")
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
				return errors.Wrap(err, "Marshal")
			}
			err = ioutil.WriteFile("latest_version.json", output, 0644)
			if err != nil {
				return errors.Wrap(err, "WriteFile")
			}
		}
	}

	if len(output) != 0 {
		_, err = fmt.Fprintf(r.stdOut, string(output))
	} else {
		_, err = fmt.Fprintf(r.stdOut, "{}")
	}
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
