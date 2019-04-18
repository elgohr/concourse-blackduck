package interpreter

import (
	"errors"
	"strings"
)

type Response struct {
	Id       Ref        `json:"version"`
	MetaData []MetaData `json:"metadata"`
}

type Ref struct {
	Ref string `json:"ref"`
}

type MetaData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func NewResponse(logResponse string) (Response, error) {
	id, name, version, url, err := extractAttributes(logResponse)
	return Response{
		Id: Ref{Ref: id},
		MetaData: []MetaData{
			{Name: "name", Value: name},
			{Name: "version", Value: version},
			{Name: "url", Value: url},
		},
	}, err
}

func extractAttributes(logResponse string) (id string, name string, version string, url string, err error) {
	lines := strings.Split(logResponse, "\n")
	for _, line := range lines {
		if strings.Contains(line, "To see your results, follow the URL") {
			url = extractFromLine(line)
			id = extractIdFromUrl(url)
		} else if strings.Contains(line, "Project version") {
			version = extractFromLine(line)
		} else if strings.Contains(line, "Project name") {
			name = extractFromLine(line)
		} else if strings.Contains(line, "Overall Status") {
			status := extractFromLine(line)
			if status != "SUCCESS" {
				err = errors.New(status)
			}
		}
	}
	return
}

func extractFromLine(line string) string {
	return strings.Split(line, ": ")[1]
}

func extractIdFromUrl(url string) string {
	urlParts := strings.Split(url, "/")
	for i, p := range urlParts {
		if p == "versions" {
			return urlParts[i+1]
		}
	}
	return ""
}
