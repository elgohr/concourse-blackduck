package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/elgohr/blackduck-resource/shared"
	"github.com/elgohr/blackduck-resource/shared/sharedfakes"
	"io/ioutil"
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

func clean(t *testing.T) {
	if err := os.Remove("latest_version.json"); err != nil {
		t.Error(err)
	}
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
	fakeBlackduckApi.GetProjectVersionsReturns([]shared.Version{
		{
			Updated: now,
			Name:    "TEST_VERSION",
			Phase:   "TEST_PHASE",
		},
	}, nil)

	stdIn.WriteString(fmt.Sprintf(`{
				"source": {
	    			"url": "http://blackduck",
					"username": "username",
	    			"password": "password",
					"name": "project1"
	  			},
				"version": {"ref": "%v"}
			}`, now))

	if err := r.run(); err != nil {
		t.Error(err)
	}

	expRes := fmt.Sprintf(`{"version":{"ref":"%v"},"metadata":[{"name":"versionName","value":"TEST_VERSION"},{"name":"phase","value":"TEST_PHASE"},{"name":"settingUpdatedAt","value":"%v"}]}`, now, now)
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
           Got:      %v`, expRes, stdOut.String())
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
	b, err := ioutil.ReadFile("latest_version.json")
	if err != nil {
		t.Error(err)
	}
	if string(b)!= expRes {
		t.Errorf(`Expected: %v
           Got:      %v`, expRes, string(b))
	}
	clean(t)
}

func TestReturnsEmptyArrayWhenNoLatestVersions(t *testing.T) {
	stdIn, stdOut, fakeBlackduckApi, r := setup()

	fakeProject := shared.Project{Name: "TEST_PROJECT"}
	fakeBlackduckApi.GetProjectByNameReturns(&fakeProject, nil)
	now := time.Now()
	fakeBlackduckApi.GetProjectVersionsReturns([]shared.Version{}, nil)

	stdIn.WriteString(fmt.Sprintf(`{
				"source": {
	    			"url": "http://blackduck",
					"username": "username",
	    			"password": "password",
					"name": "project1"
	  			},
				"version": {"ref": "%v"}
			}`, now))

	if err := r.run(); err != nil {
		t.Error(err)
	}

	if stdOut.String() != "{}" {
		t.Errorf(`Expected {}, got %v`, stdOut.String())
	}
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
