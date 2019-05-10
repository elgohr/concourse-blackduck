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
	GetProjectByName(source Source) (*Project, error)
	GetProjectVersions(source Source, project *Project) ([]Version, error)
}

type Blackduck struct {
	client http.Client
}

func NewBlackduck() Blackduck {
	return Blackduck{
		client: http.Client{},
	}
}

func (b *Blackduck) GetProjectByName(source Source) (*Project, error) {
	if cache, cached := projectIsCached(); cached {
		var cachedProject Project
		if err := json.Unmarshal(cache, &cachedProject); err == nil {
			return &cachedProject, nil
		}
	}

	token, err := authenticate(source)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", source.GetProjectUrl(), nil)
	res, err := b.client.Do(authenticatedRequest(*req, token))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var projectList ProjectList
	if err := json.NewDecoder(res.Body).Decode(&projectList); err == nil {
		for _, project := range projectList.Projects {
			if project.Name == source.Name {
				writeProjectToCache(project)
				return &project, nil
			}
		}
	} else {
		return nil, err
	}
	return nil, errors.New("no project matching the name")
}

func (b *Blackduck) GetProjectVersions(source Source, project *Project) ([]Version, error) {
	versionsLink := project.Meta.GetLinkFor("versions")

	token, err := authenticate(source)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", versionsLink, nil)
	res, err := b.client.Do(authenticatedRequest(*req, token))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var versionList VersionList
	if err := json.NewDecoder(res.Body).Decode(&versionList); err != nil {
		return nil, err
	}
	return sortVersionsChronologically(versionList), nil
}

func authenticate(source Source) (token string, err error) {
	formValues := url.Values{
		"j_username": {source.Username},
		"j_password": {source.Password},
	}
	authUrl := source.Url + "/j_spring_security_check"
	res, err := http.PostForm(authUrl, formValues)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode > 300 {
		return "", errors.New("authentication failed")
	}

	cookieHeader := res.Header.Get("Set-Cookie")
	cookieHeaderEntries := strings.Split(cookieHeader, ";")

	for _, cookieHeaderEntry := range cookieHeaderEntries {
		trimmedCookieHeaderEntry := strings.TrimSpace(cookieHeaderEntry)
		if strings.HasPrefix(trimmedCookieHeaderEntry, tokenPrefix) {
			token = strings.Replace(trimmedCookieHeaderEntry, tokenPrefix, "", 1)
			return token, nil
		}
	}
	return "", errors.New("token not found")
}

func authenticatedRequest(req http.Request, token string) (authReq *http.Request) {
	req.Header.Set("Cookie", tokenPrefix+token)
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
