package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elgohr/blackduck-resource/shared"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"time"
)

const ProjectCacheName = "./project.cache"

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
	if cache, cached := projectIsCached(); cached {
		var cachedProject Project
		if err := json.Unmarshal(cache, &cachedProject); err == nil {
			return &cachedProject, nil
		}
	}
	projectUrl, err := getProjectUrlFrom(baseUrl)
	if err != nil {
		return nil, err
	}
	res, err := http.Get(projectUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var projectList ProjectList
	if err := json.NewDecoder(res.Body).Decode(&projectList); err == nil {
		for _, project := range projectList.Projects {
			if project.Name == name {
				writeProjectToCache(project)
				return &project, nil
			}
		}
	} else {
		return nil, err
	}
	return nil, errors.New("no project matching the name")
}

func writeProjectToCache(project Project) {
	if b, err := json.Marshal(project); err == nil {
		_ = ioutil.WriteFile(ProjectCacheName, b, 0644)
	}
}

func projectIsCached() (content []byte, cached bool) {
	if content, err := ioutil.ReadFile(ProjectCacheName); err == nil {
		return content, true
	}
	return nil, false
}

func getProjectUrlFrom(baseUrl string) (projectUrl string, err error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "api/projects")
	return u.String(), nil
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
	versionsLink := project.Meta.GetLinkFor("versions")
	res, err := http.Get(versionsLink)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var versionList VersionList
	if err := json.NewDecoder(res.Body).Decode(&versionList); err != nil {
		return nil, err
	}
	var refs []shared.Ref
	for _, version := range sortVersionsChronologically(versionList) {
		versionRef := fmt.Sprintf("%v-%v", version.Name, version.Phase)
		refs = append(refs, shared.Ref{Ref: versionRef})
	}
	return refs, nil
}

func sortVersionsChronologically(versionList VersionList) []Version {
	versions := versionList.Versions
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Updated.Before(versions[j].Updated)
	})
	return versions
}

type VersionList struct {
	Versions []Version `json:"items"`
}

type Version struct {
	Name    string    `json:"versionName"`
	Phase   string    `json:"phase"`
	Updated time.Time `json:"settingUpdatedAt"`
}
