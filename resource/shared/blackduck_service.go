package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const ProjectCacheName = "./project.cache"

//go:generate counterfeiter . BlackduckApi
type BlackduckApi interface {
	GetProjectByName(projectUrl string, name string) (*Project, error)
	GetProjectVersions(project *Project) ([]Ref, error)
}

type Blackduck struct {}

func (b *Blackduck) GetProjectByName(projectUrl string, name string) (*Project, error) {
	if cache, cached := projectIsCached(); cached {
		var cachedProject Project
		if err := json.Unmarshal(cache, &cachedProject); err == nil {
			return &cachedProject, nil
		}
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

func (b *Blackduck) GetProjectVersions(project *Project) ([]Ref, error) {
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
	var refs []Ref
	for _, version := range sortVersionsChronologically(versionList) {
		versionRef := fmt.Sprintf("%v-%v", version.Name, version.Phase)
		refs = append(refs, Ref{Ref: versionRef})
	}
	return refs, nil
}

func (b *Blackduck) GetAuthenticationToken(blackBaseDuckUrl string, blackDuckUsername string, blackDuckPassword string) (string, error) {

	const urlLoginPostfix = "/j_spring_security_check"
	const formUserName= "j_username"
	const formPassword = "j_password"

	res, err := http.PostForm(blackBaseDuckUrl+urlLoginPostfix,
		url.Values{formUserName: {blackDuckUsername}, formPassword: {blackDuckPassword}})

	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if isStatusCodeNotOk(res) {
		return "", errors.New("authentication failed")
	}

	cookieHeader := res.Header.Get("Set-Cookie")

	cookieHeaderEntries := strings.Split(cookieHeader, ";")

	var token string
	for _, cookieHeaderEntry := range cookieHeaderEntries {
		trimmedCookieHeaderEntry := strings.TrimSpace(cookieHeaderEntry)
		const tokenPrefix = "AUTHORIZATION_BEARER="
		if strings.HasPrefix(trimmedCookieHeaderEntry, tokenPrefix) {
			token = strings.Replace(trimmedCookieHeaderEntry, tokenPrefix, "", 1)
			return token, nil
		}
	}

	return "", errors.New("token not found")
}

func isStatusCodeNotOk(res *http.Response) bool {
	return res.StatusCode >= 300
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

func sortVersionsChronologically(versionList VersionList) (versions []Version) {
	versions = versionList.Versions
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Updated.Before(versions[j].Updated)
	})
	return
}
