package shared_test

import (
	"github.com/elgohr/blackduck-resource/shared"
	"testing"
)

func TestIsValidWhenAllPropertiesAreFilled(t *testing.T) {
	s := shared.Source{
		Url:"url",
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
		Url:"url",
		Password:"password",
		Name:"name",
	}
	if s.Valid() {
		t.Error("Should be invalid, but wasn't")
	}
}

func TestIsInvalidWhenPasswordIsMissing(t *testing.T) {
	s := shared.Source{
		Url:"url",
		Username:"user",
		Name:"name",
	}
	if s.Valid() {
		t.Error("Should be invalid, but wasn't")
	}
}

func TestIsValidWhenTokenIsPresentButUsernameAndPasswordAreMissing(t *testing.T) {
	s := shared.Source{
		Url:"url",
		Token:"token",
		Name:"name",
	}
	if !s.Valid() {
		t.Error("Should be valid, but wasn't")
	}
}

func TestIsValidWhenTokenIsMissingButUsernameAndPasswordArePresent(t *testing.T) {
	s := shared.Source{
		Url:"url",
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
		Url:"url",
		Username:"user",
		Password:"password",
	}
	if s.Valid() {
		t.Error("Should be invalid, but wasn't")
	}
}
