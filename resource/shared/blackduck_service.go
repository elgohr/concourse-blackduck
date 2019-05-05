package shared

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const (
	ProjectCacheName = "./project.cache"
	tokenPrefix      = "AUTHORIZATION_BEARER="
)

//go:generate counterfeiter . BlackduckApi
type BlackduckApi interface {
	GetProjectByName(projectUrl string, name string) (*Project, error)
	GetProjectVersions(project *Project) ([]Ref, error)
	Authenticate(baseUrl string, username string, password string) error
}

type Blackduck struct {
	client http.Client
	token  string
}

func NewBlackduck() Blackduck {
	return Blackduck{
		client: http.Client{},
	}
}

func (b *Blackduck) GetProjectByName(projectUrl string, name string) (*Project, error) {
	if cache, cached := projectIsCached(); cached {
		var cachedProject Project
		if err := json.Unmarshal(cache, &cachedProject); err == nil {
			return &cachedProject, nil
		}
	}

	req, _ := http.NewRequest("GET", projectUrl, nil)
	res, err := b.client.Do(authenticatedRequest(*req, *b))
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

	req, _ := http.NewRequest("GET", versionsLink, nil)
	res, err := b.client.Do(authenticatedRequest(*req, *b))
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
		refs = append(refs, Ref{Ref: version.Updated.String()})
	}
	return refs, nil
}

func (b *Blackduck) Authenticate(baseUrl string, username string, password string) error {
	formValues := url.Values{
		"j_username": {username},
		"j_password": {password},
	}
	authUrl := baseUrl + "/j_spring_security_check"
	res, err := http.PostForm(authUrl, formValues)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 300 {
		return errors.New("authentication failed")
	}

	cookieHeader := res.Header.Get("Set-Cookie")
	cookieHeaderEntries := strings.Split(cookieHeader, ";")

	for _, cookieHeaderEntry := range cookieHeaderEntries {
		trimmedCookieHeaderEntry := strings.TrimSpace(cookieHeaderEntry)
		if strings.HasPrefix(trimmedCookieHeaderEntry, tokenPrefix) {
			b.token = strings.Replace(trimmedCookieHeaderEntry, tokenPrefix, "", 1)
			return nil
		}
	}
	return errors.New("token not found")
}

func authenticatedRequest(req http.Request, b Blackduck) (authReq *http.Request) {
	req.Header.Set("Cookie", tokenPrefix+b.token)
	return &req
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
