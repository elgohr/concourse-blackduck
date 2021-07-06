package shared

import (
	"crypto/tls"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	ProjectCacheName = "./project.cache"
	tokenPrefix      = "AUTHORIZATION_BEARER="
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . BlackduckApi
type BlackduckApi interface {
	GetProjectByName(source Source) (*Project, error)
	GetProjectVersions(source Source, project *Project) ([]Version, error)
}

type Blackduck struct {
	client http.Client
}

func NewBlackduck() Blackduck {
	return Blackduck{
		client: http.Client{Timeout: 30 * time.Second},
	}
}

func (b *Blackduck) GetProjectByName(source Source) (*Project, error) {
	if cache, cached := projectIsCached(); cached {
		var cachedProject Project
		if err := json.Unmarshal(cache, &cachedProject); err == nil {
			return &cachedProject, nil
		}
	}

	if source.Insecure {
		b.client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	token, err := authenticate(source)
	if err != nil {
		return nil, errors.Wrap(err, "GetProjectByName")
	}

	req, _ := http.NewRequest(http.MethodGet, source.GetProjectUrl(), nil)
	res, err := b.client.Do(authenticatedRequest(*req, token))
	if err != nil {
		return nil, errors.Wrap(errors.Wrap(err, "GetProjectByName"),"GetProjectUrl")
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
		return nil, errors.Wrap(errors.Wrap(err, "Decode"),"GetProjectByName")
	}
	return nil, errors.Wrap(errors.New("no project matching the name"), "GetProjectByName")
}

func (b *Blackduck) GetProjectVersions(source Source, project *Project) ([]Version, error) {
	versionsLink := project.Meta.GetLinkFor("versions")

	if source.Insecure {
		b.client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	token, err := authenticate(source)
	if err != nil {
		return nil, errors.Wrap(err, "GetProjectVersions")
	}

	req, _ := http.NewRequest(http.MethodGet, versionsLink, nil)
	res, err := b.client.Do(authenticatedRequest(*req, token))
	if err != nil {
		return nil, errors.Wrap(errors.Wrap(err, "GetProjectVersions"),"Versions")
	}
	defer res.Body.Close()

	var versionList VersionList
	if err := json.NewDecoder(res.Body).Decode(&versionList); err != nil {
		return nil, errors.Wrap(errors.Wrap(err, "Decode"),"GetProjectVersions")
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
		return "", errors.Wrap(err, "Authentication")
	}
	defer res.Body.Close()

	if res.StatusCode > 300 {
		return "", errors.Wrap(errors.New("authentication failed"), "Authentication")
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
	return "", errors.Wrap(errors.New("token not found"), "Authentication")
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
