package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/elgohr/concourse-blackduck/shared"
	"github.com/elgohr/concourse-blackduck/shared/sharedfakes"
	"os"
	"testing"
	"time"
)

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
	now := time.Now()
	fakeBlackduckApi.GetProjectVersionsReturns([]shared.Version{{Updated: now}}, nil)

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

	expRes := fmt.Sprintf(`[{"ref":"%v"}]`, now)
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
					Got:   %v`, expRes, stdOut.String())
	}
	source := fakeBlackduckApi.GetProjectByNameArgsForCall(0)
	pu := "http://blackduck"
	if source.Url != pu {
		t.Errorf("Expected api to be called with projectUrl %v, but was called with %v", pu, source.Url)
	}
	pt := "project1"
	if source.Name != pt {
		t.Errorf("Expected api to be called with projectName %v, but was called with %v", pt, source.Name)
	}
}

func TestErrorsWhenProjectCouldNotBeFound(t *testing.T) {
	stdIn, _, fakeBlackduckApi, r := setup()

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
}

func TestErrorsWhenTheProvidedUrlIsInvalid(t *testing.T) {
	stdIn, _, fakeBlackduckApi, r := setup()

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
}

func TestErrorsWhenGetProjectByNameReturnsAnError(t *testing.T) {
	stdIn, _, fakeBlackduckApi, r := setup()

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

	msg := "GetProjectByName: " + expError.Error()
	if err := r.run(); err.Error() != msg {
		t.Errorf("Should return %v, but was %v", msg, err.Error())
	}
}

func TestErrorsWhenGetProjectVersionReturnsAnError(t *testing.T) {
	stdIn, _, fakeBlackduckApi, r := setup()

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

	msg := "GetProjectVersions: " + expError.Error()
	if err := r.run(); err.Error() != msg {
		t.Errorf("Should return %v, but was %v", msg, err.Error())
	}
}
