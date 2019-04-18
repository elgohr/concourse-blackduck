package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
}

func TestQueriesForTheLatestVersions(t *testing.T) {
	var (
		calledProjects  bool
		calledVersions  bool
		projectResponse []byte
	)
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/api/projects" {
			calledProjects = true
			w.Write(projectResponse)
		} else if r.RequestURI == "/api/projects/01883e41-d4c9-420a-b41b-0ddcaadda2b5/versions" {
			calledVersions = true
			b, err := ioutil.ReadFile("testdata/versions.json")
			if err != nil {
				t.Error(err)
			}
			w.Write(b)
		}

	}))
	defer h.Close()

	b, err := ioutil.ReadFile("testdata/projects.json")
	if err != nil {
		t.Error(err)
	}
	projectResponse = []byte(fmt.Sprintf(string(b), h.URL))

	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
	}

	stdIn.WriteString(fmt.Sprintf(`{
			"source": {
    			"url": "%v",
				"username": "username",
    			"password": "password",
				"name": "project1"
  			}
		}`, h.URL))

	if err := r.run(); err != nil {
		t.Error(err)
	}

	expRes := `[{"ref":"0.1.1"}]`
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
				Got:   %v`, expRes, stdOut.String())
	}
	if !calledProjects {
		t.Error("Didn't call the Blackduck api for projects")
	}

	if !calledVersions {
		t.Error("Didn't call the Blackduck api for versions")
	}
}

func TestErrorsWhenInputIsCorrupted(t *testing.T) {
	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
	}

	stdIn.WriteString(`{:]`)

	if err := r.run(); err == nil {
		t.Error("Should have errored, but didn't")
	}
}

func TestErrorsWhenProjectCouldNotBeFound(t *testing.T) {
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer h.Close()

	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
	}

	stdIn.WriteString(fmt.Sprintf(`{
			"source": {
    			"url": "%v",
				"username": "username",
    			"password": "password",
				"name": "not_here"
  			}
		}`, h.URL))

	if err := r.run(); err == nil {
		t.Error("Should have errored, but didn't")
	}

	expRes := `[]`
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
				Got:   %v`, expRes, stdOut.String())
	}
}

func TestErrorsWhenProjectResponseIsCorrupted(t *testing.T) {
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{]`))
	}))
	defer h.Close()

	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
	}

	stdIn.WriteString(fmt.Sprintf(`{
			"source": {
    			"url": "%v",
				"username": "username",
    			"password": "password",
				"name": "project1"
  			}
		}`, h.URL))

	if err := r.run(); err == nil {
		t.Error("Should have errored, but didn't")
	}
}

func TestErrorsWhenBlackduckCouldNotBeReachedForProjects(t *testing.T) {
	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
	}

	stdIn.WriteString(`{
			"source": {
    			"url": "http://localhost",
				"username": "username",
    			"password": "password",
				"name": "project1"
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
	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
	}

	stdIn.WriteString(`{
			"source": {
    			"url": "http:\\!",
				"username": "username",
    			"password": "password",
				"name": "project1"
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

func TestErrorsWhenBlackduckCouldNotBeReachedForVersions(t *testing.T) {
	var (
		h               *httptest.Server
		terminate       = make(chan bool, 1)
		projectResponse []byte
	)
	go func() {
		<-terminate
		h.Close()
	}()
	h = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		terminate <- true
		w.Write(projectResponse)
	}))

	b, err := ioutil.ReadFile("testdata/projects.json")
	if err != nil {
		t.Error(err)
	}
	projectResponse = []byte(fmt.Sprintf(string(b), h.URL))

	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
	}

	stdIn.WriteString(fmt.Sprintf(`{
			"source": {
    			"url": "%v",
				"username": "username",
    			"password": "password",
				"name": "project1"
  			}
		}`, h.URL))

	if err := r.run(); err == nil {
		t.Error("Should have errored, but didn't")
	}

	expRes := `[]`
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
				Got:   %v`, expRes, stdOut.String())
	}
}
