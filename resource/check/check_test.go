package main

import (
	"bytes"
	"errors"
	"github.com/elgohr/blackduck-resource/shared"
	"github.com/elgohr/blackduck-resource/shared/sharedfakes"
	"os"
	"testing"
)

func TestConstructsRunnerCorrectly(t *testing.T) {
	r := NewRunner()
	if r.stdIn != os.Stdin {
		t.Error("Didn't set stdIn correctly")
	}
	if r.stdOut != os.Stdout {
		t.Error("Didn't set stdOut correctly")
	}
	if r.stdErr != os.Stderr {
		t.Error("Didn't set stdErr correctly")
	}
	if r.api == nil {
		t.Error("Didn't set Blackduck Api")
	}
}

func TestQueriesForTheLatestVersionsInChronologicalOrder(t *testing.T) {
	stdIn, stdOut, fakeBlackduckApi, r := setup()

	fakeProject := shared.Project{Name: "TEST_PROJECT"}
	fakeBlackduckApi.GetProjectByNameReturns(&fakeProject, nil)
	fakeRefs := []shared.Ref{{Ref: "TEST_REF"}}
	fakeBlackduckApi.GetProjectVersionsReturns(fakeRefs, nil)

	stdIn.WriteString(`{
				"source": {
	    			"url": "http://blackduck",
					"username": "username",
	    			"password": "password",
					"name": "project1"
	  			}
			}`)

	if err := r.run(); err != nil {
		t.Error(err)
	}

	expRes := `[{"ref":"TEST_REF"}]`
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
					Got:   %v`, expRes, stdOut.String())
	}
	projectUrl, projectName := fakeBlackduckApi.GetProjectByNameArgsForCall(0)
	pu := "http://blackduck/api/projects"
	if projectUrl != pu {
		t.Errorf("Expected api to be called with projectUrl %v, but was called with %v", pu, projectUrl)
	}
	pt := "project1"
	if projectName != pt {
		t.Errorf("Expected api to be called with projectName %v, but was called with %v", pt, projectName)
	}
}

func setup() (stdIn *bytes.Buffer, stdOut *bytes.Buffer, fakeBlackduckApi *sharedfakes.FakeBlackduckApi, r Runner) {
	stdIn = &bytes.Buffer{}
	stdOut = &bytes.Buffer{}
	fakeBlackduckApi = &sharedfakes.FakeBlackduckApi{}
	r = Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
		api:    fakeBlackduckApi,
	}
	return
}

func TestErrorsWhenProjectCouldNotBeFound(t *testing.T) {
	stdIn, stdOut, fakeBlackduckApi, r := setup()

	fakeBlackduckApi.GetProjectByNameReturns(nil, errors.New("no project matching the name"))
	stdIn.WriteString(`{
			"source": {
    			"url": "http://blackduck",
				"username": "username",
    			"password": "password",
				"name": "not_here"
  			}
		}`)

	if err := r.run(); err == nil {
		t.Error("Should have errored, but didn't")
	}

	expRes := `[]`
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
				Got:   %v`, expRes, stdOut.String())
	}
}

func TestErrorsWhenTheProvidedUrlIsInvalid(t *testing.T) {
	stdIn, stdOut, fakeBlackduckApi, r := setup()

	stdIn.WriteString(`{
			"source": {
    			"url": "ht!",
				"username": "username",
    			"password": "password",
				"name": "project1"
  			}
		}`)

	if err := r.run(); err == nil {
		t.Error("Should have errored, but didn't")
	}

	if fakeBlackduckApi.GetProjectByNameCallCount() > 0 {
		t.Error("Should not have been called")
	}

	if fakeBlackduckApi.GetProjectVersionsCallCount() > 0 {
		t.Error("Should not have been called")
	}

	expRes := `[]`
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
				Got:   %v`, expRes, stdOut.String())
	}
}

func TestErrorsWhenGetProjectByNameReturnsAnError(t *testing.T) {
	stdIn, stdOut, fakeBlackduckApi, r := setup()

	expError := errors.New("something bad")
	fakeBlackduckApi.GetProjectByNameReturns(nil, expError)

	stdIn.WriteString(`{
			"source": {
    			"url": "http://blackduck",
				"username": "username",
    			"password": "password",
				"name": "project1"
  			}
		}`)

	if err := r.run(); err != expError {
		t.Error("Should return GetProjectByNameError")
	}

	expRes := `[]`
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
				Got:   %v`, expRes, stdOut.String())
	}
}

func TestErrorsWhenGetProjectVersionReturnsAnError(t *testing.T) {
	stdIn, stdOut, fakeBlackduckApi, r := setup()

	fakeProject := shared.Project{Name: "TEST_PROJECT"}
	expError := errors.New("something bad")
	fakeBlackduckApi.GetProjectByNameReturns(&fakeProject, nil)
	fakeBlackduckApi.GetProjectVersionsReturns(nil, expError)

	stdIn.WriteString(`{
			"source": {
    			"url": "http://blackduck",
				"username": "username",
    			"password": "password",
				"name": "project1"
  			}
		}`)

	if err := r.run(); err != expError {
		t.Error("Should return GetProjectByNameError")
	}

	expRes := `[]`
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
				Got:   %v`, expRes, stdOut.String())
	}
}
