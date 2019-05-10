package shared

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
