package shared_test

import (
	"github.com/elgohr/blackduck-resource/shared"
	"testing"
)

func TestIsValidWhenDirectoryIsFilled(t *testing.T) {
	p := shared.Params{
		Directory:"directory",
	}
	if !p.Valid() {
		t.Error("Should be valid, but wasn't")
	}
}

func TestIsInvalidWhenDirectoryIsNotFilled(t *testing.T) {
	p := shared.Params{}
	if p.Valid() {
		t.Error("Should be invalid, but wasn't")
	}
}
