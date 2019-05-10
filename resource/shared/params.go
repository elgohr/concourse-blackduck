package shared

type Params struct {
	Directory string `json:"directory"`
}

func (p *Params) Valid() bool {
	return len(p.Directory) != 0
}
