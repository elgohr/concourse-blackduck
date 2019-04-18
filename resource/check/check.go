package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elgohr/blackduck-resource/shared"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
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
}

func NewRunner() Runner {
	return Runner{
		stdIn:  os.Stdin,
		stdOut: os.Stdout,
		stdErr: os.Stderr,
	}
}

func (r *Runner) run() error {
	var input shared.Request
	if err := json.NewDecoder(r.stdIn).Decode(&input); err != nil {
		return err
	}
	project, err := r.getProjectByName(input.Source.Url, input.Source.Name)
	if err != nil {
		fmt.Fprintf(r.stdOut, `[]`)
		return err
	}

	versions, err := r.getProjectVersions(project)
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

func (r *Runner) getProjectByName(baseUrl string, name string) (*Project, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "api/projects")
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var projectList ProjectList
	if err := json.NewDecoder(res.Body).Decode(&projectList); err == nil {
		for _, project := range projectList.Projects {
			if project.Name == name {
				return &project, nil
			}
		}
	} else {
		return nil, err
	}
	return nil, errors.New("no project matching the name")
}

type ProjectList struct {
	Projects []Project `json:"items"`
}

type Project struct {
	Name string `json:"name"`
	Meta Meta   `json:"_meta"`
}

type Meta struct {
	Links []Link `json:"links"`
}

func (m *Meta) GetLinkFor(key string) string {
	for _, link := range m.Links {
		if link.Rel == key {
			return link.Href
		}
	}
	return ""
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

func (r *Runner) getProjectVersions(project *Project) ([]shared.Ref, error) {
	versions := project.Meta.GetLinkFor("versions")
	res, err := http.Get(versions)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var versionList VersionList
	if err := json.NewDecoder(res.Body).Decode(&versionList); err != nil {
		return nil, err
	}
	var refs []shared.Ref
	for _, version := range versionList.Versions {
		refs = append(refs, shared.Ref{Ref: version.Name})
	}
	return refs, nil
}

type VersionList struct {
	Versions []Version `json:"items"`
}

type Version struct {
	Name string `json:"versionName"`
}
