package shared

import (
	"net/url"
	"path"
)

type Request struct {
	Source Source `json:"source"`
	Params Params `json:"params"`
}

type Source struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Insecure bool   `json:"insecure"`
}

func (s *Source) Valid() bool {
	_, err := url.ParseRequestURI(s.Url)
	return len(s.Username) != 0 &&
		len(s.Password) != 0 &&
		len(s.Name) != 0 &&
		err == nil
}

func (s *Source) GetProjectUrl() string {
	u, err := url.ParseRequestURI(s.Url)
	if err != nil {
		return ""
	}
	u.Path = path.Join(u.Path, "api/projects")
	return u.String()
}
