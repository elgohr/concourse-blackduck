package shared_test

import (
	"github.com/elgohr/concourse-blackduck/shared"
	"testing"
)

func TestIsValidWhenAllMandatoryPropertiesAreFilled(t *testing.T) {
	s := shared.Source{
		Url:"http://url",
		Username:"user",
		Password:"password",
		Name:"name",
	}
	if !s.Valid() {
		t.Error("Should be valid, but wasn't")
	}
}

func TestIsInvalidWhenUrlIsMissing(t *testing.T) {
	s := shared.Source{
		Username:"user",
		Password:"password",
		Name:"name",
	}
	if s.Valid() {
		t.Error("Should be invalid, but wasn't")
	}
}

func TestIsInvalidWhenUsernameIsMissing(t *testing.T) {
	s := shared.Source{
		Url:"http://url",
		Password:"password",
		Name:"name",
	}
	if s.Valid() {
		t.Error("Should be invalid, but wasn't")
	}
}

func TestIsInvalidWhenPasswordIsMissing(t *testing.T) {
	s := shared.Source{
		Url:"http://url",
		Username:"user",
		Name:"name",
	}
	if s.Valid() {
		t.Error("Should be invalid, but wasn't")
	}
}

func TestIsValidWhenTokenIsMissingButUsernameAndPasswordArePresent(t *testing.T) {
	s := shared.Source{
		Url:"http://url",
		Username:"user",
		Password:"password",
		Name:"name",
	}
	if !s.Valid() {
		t.Error("Should be valid, but wasn't")
	}
}

func TestIsInvalidWhenNameIsMissing(t *testing.T) {
	s := shared.Source{
		Url:"http://url",
		Username:"user",
		Password:"password",
	}
	if s.Valid() {
		t.Error("Should be invalid, but wasn't")
	}
}

func TestIsInvalidWhenUrlIsNoUrl(t *testing.T) {
	s := shared.Source{
		Url:"no_url",
		Username:"user",
		Password:"password",
		Name:"name",
	}
	if s.Valid() {
		t.Error("Should be invalid, but wasn't")
	}
}

func TestReturnsProjectUrl(t *testing.T) {
	s := shared.Source{
		Url:"http://url",
		Username:"user",
		Password:"password",
		Name:"name",
	}
	url := s.GetProjectUrl()
	if url != "http://url/api/projects?q=name:name" {
		t.Errorf("Should've appended URL with the project url, but was %v", url)
	}
}

func TestProjectUrlIsOkWithExtraSlash(t *testing.T) {
	s := shared.Source{
		Url:"http://url/",
		Username:"user",
		Password:"password",
		Name:"name",
	}
	url := s.GetProjectUrl()
	if url != "http://url/api/projects?q=name:name" {
		t.Errorf("Should've appended URL with the project url, but was %v", url)
	}
}

func TestProjectUrlIsEmptyWhenUrlWasSetIncorrectly(t *testing.T) {
	s := shared.Source{
		Url:"no_url",
		Username:"user",
		Password:"password",
		Name:"name",
	}
	url := s.GetProjectUrl()
	if url != "" {
		t.Errorf("Should've been empty, but was %v", url)
	}
}
