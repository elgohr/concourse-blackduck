package interpreter_test

import (
	"github.com/elgohr/concourse-blackduck/out/interpreter"
	"io/ioutil"
	"os"
	"testing"
)

var (
	blackduckResponse      string
	blackduckErrorResponse string
)

func TestMain(m *testing.M) {
	b, err := ioutil.ReadFile("../testdata/blackduckResponse.txt")
	if err != nil {
		panic(err)
	}
	blackduckResponse = string(b)
	b, err = ioutil.ReadFile("../testdata/blackduckError.txt")
	if err != nil {
		panic(err)
	}
	blackduckErrorResponse = string(b)
	os.Exit(m.Run())
}

func TestExtractsUrl(t *testing.T) {
	e := "https://my.host/api/projects/d6aed8bb-0b9a-46a2-a1ce-60101939eb10/versions/d1530c19-1541-443f-8a5e-ea4e17c856a8/components"
	containsMetaDate(t, blackduckResponse, "url", e)
}

func TestUrlIsEmptyWhenNotFound(t *testing.T) {
	containsMetaDate(t, "", "url", "")
}

func TestExtractsVersion(t *testing.T) {
	e := "1.0.0"
	containsMetaDate(t, blackduckResponse, "version", e)
}

func TestVersionIsEmptyWhenNotFound(t *testing.T) {
	containsMetaDate(t, "", "version", "")
}

func TestExtractsName(t *testing.T) {
	e := "presentations"
	containsMetaDate(t, blackduckResponse, "name", e)
}

func TestNameIsEmptyWhenNotFound(t *testing.T) {
	containsMetaDate(t, "", "name", "")
}

func TestExtractsId(t *testing.T) {
	e := "d1530c19-1541-443f-8a5e-ea4e17c856a8"
	l, _ := interpreter.NewResponse(blackduckResponse)
	if l.Id.Ref != e {
		t.Errorf("Expected id to be %v, but was %v", e, l.Id)
	}
}

func TestIdIsEmptyWhenNotFound(t *testing.T) {
	l, _ := interpreter.NewResponse("")
	if l.Id.Ref != "" {
		t.Errorf("Expected id to be empty, but was %v", l.Id)
	}
}

func TestReturnsNoErrorWhenTheOverallStatusWasSuccess(t *testing.T) {
	_, e := interpreter.NewResponse(blackduckResponse)
	if e != nil {
		t.Error(e)
	}
}

func TestReturnsAnErrorWhenTheOverallStatusWasNotSuccess(t *testing.T) {
	_, e := interpreter.NewResponse(blackduckErrorResponse)
	if e != nil && e.Error() != "FAILURE_DETECTOR" {
		t.Errorf("Expected error to be FAILURE_DETECTOR, but was %v", e.Error())
	}
}

func containsMetaDate(t *testing.T, response string, name string, value string) {
	var contained bool
	e, _ := interpreter.NewResponse(response)
	for _, date := range e.MetaData {
		if date.Name == name {
			if date.Value != value {
				t.Errorf("Expected %v to be %v, but was %v", name, value, date.Value)
			}
			contained = true
		}
	}
	if !contained {
		t.Error(name + " was not in metadata")
	}
}
