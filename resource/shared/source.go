package shared

type Request struct {
	Source Source `json:"source"`
	Params Params `json:"params"`
}

type Source struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (s *Source) Valid() bool {
	return len(s.Url) != 0 &&
		len(s.Username) != 0 &&
		len(s.Password) != 0 &&
		len(s.Name) != 0
}
