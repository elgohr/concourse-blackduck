package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"testing"
)

func setup(t *testing.T) (runner Runner) {
	tmpDir, err := ioutil.TempDir("", "concourse-blackduck")
	if err != nil {
		t.Error(err)
	}
	return Runner{
		stdIn:       &bytes.Buffer{},
		stdOut:      &bytes.Buffer{},
		stdErr:      &bytes.Buffer{},
		downloadDir: tmpDir,
	}
}

func TestStartsBlackduck(t *testing.T) {
	r := setup(t)

	stdIn := r.stdIn.(*bytes.Buffer)
	targetUrl := "https://BLACKDUCK"
	username := "USERNAME"
	password := "PASSWORD"
	stdIn.WriteString(fmt.Sprintf(`{
			"source": {
    			"url": "%v",
				"username": "%v",
    			"password": "%v"
  			}
		}`, targetUrl, username, password))

	var called bool
	r.exec = func(name string, arg ...string) *exec.Cmd {
		called = true
		if name != "java" {
			t.Errorf("Should have started java, but started %v", name)
		}
		expectedArgs := []string{
			"-jar",
			"synopsys-detect-5.3.3.jar",
			"--blackduck.url=" + targetUrl,
			"--blackduck.username=" + username,
			"--blackduck.password=" + password,
		}
		for i, a := range arg {
			if a != expectedArgs[i] {
				t.Errorf("Expected argument %v, but got %v", expectedArgs[i], a)
			}
		}
		return exec.Command("true")
	}

	if err := r.run(); err != nil {
		t.Error(err)
	}
	if !called {
		t.Error("Blackduck wasn't started")
	}
}

func TestSetsTheWorkingDirectoryToTheProvidedSource(t *testing.T) {
	r := setup(t)

	command := exec.Command("true")
	r.exec = func(name string, arg ...string) *exec.Cmd {
		return command
	}

	stdIn := r.stdIn.(*bytes.Buffer)
	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
				"username": "username",
    			"password": "password"
  			}
		}`)

	if err := r.run(); err != nil {
		t.Error(err)
	}
	if command.Dir != r.downloadDir {
		t.Error("Working dir was not set correctly")
	}
}

func TestAddsLoggingToSubProcess(t *testing.T) {
	r := setup(t)

	command := exec.Command("true")
	r.exec = func(name string, arg ...string) *exec.Cmd {
		return command
	}

	stdIn := r.stdIn.(*bytes.Buffer)
	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
				"username": "username",
    			"password": "password"
  			}
		}`)

	if err := r.run(); err != nil {
		t.Error(err)
	}
	if command.Stderr != r.stdErr {
		t.Error("StdErr was not added correctly")
	}
	if command.Stdout != r.stdOut {
		t.Error("Stdout was not added correctly")
	}
}

func TestReturnsValidJson(t *testing.T) {
	r := setup(t)

	r.exec = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("true")
	}

	stdIn := r.stdIn.(*bytes.Buffer)
	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
				"username": "username",
    			"password": "password"
  			}
		}`)

	if err := r.run(); err != nil {
		t.Error(err)
	}

	stdout := r.stdOut.(*bytes.Buffer)
	if stdout.String() != `[]` {
		t.Errorf("Expected empty array, but got %v", stdout.String())
	}
}

func TestErrorsWhenUrlWasNotConfigured(t *testing.T) {
	r := setup(t)

	var called bool
	r.exec = func(name string, arg ...string) *exec.Cmd {
		called = true
		return exec.Command("true")
	}

	stdIn := r.stdIn.(*bytes.Buffer)
	stdIn.WriteString(`{
			"source": {
				"username": "username",
    			"password": "password"
  			}
		}`)

	if err := r.run(); err != nil {
		e := "missing mandatory field"
		if err.Error() != e {
			t.Errorf("Expected error message '%v', but got '%v'", e, err.Error())
		}
	} else {
		t.Error("Should have errored, but didn't")
	}
	if called {
		t.Error("Should not have called Blackduck")
	}
}

func TestErrorsWhenUsernameWasNotConfigured(t *testing.T) {
	r := setup(t)

	var called bool
	r.exec = func(name string, arg ...string) *exec.Cmd {
		called = true
		return exec.Command("true")
	}

	stdIn := r.stdIn.(*bytes.Buffer)
	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
    			"password": "password"
  			}
		}`)

	if err := r.run(); err != nil {
		e := "missing mandatory field"
		if err.Error() != e {
			t.Errorf("Expected error message '%v', but got '%v'", e, err.Error())
		}
	} else {
		t.Error("Should have errored, but didn't")
	}
	if called {
		t.Error("Should not have called Blackduck")
	}
}

func TestErrorsWhenPasswordWasNotConfigured(t *testing.T) {
	r := setup(t)

	var called bool
	r.exec = func(name string, arg ...string) *exec.Cmd {
		called = true
		return exec.Command("true")
	}

	stdIn := r.stdIn.(*bytes.Buffer)
	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
				"username": "username"
  			}
		}`)

	if err := r.run(); err != nil {
		e := "missing mandatory field"
		if err.Error() != e {
			t.Errorf("Expected error message '%v', but got '%v'", e, err.Error())
		}
	} else {
		t.Error("Should have errored, but didn't")
	}
	if called {
		t.Error("Should not have called Blackduck")
	}
}

func TestErrorsWhenConfigurationJsonIsInvalid(t *testing.T) {
	r := setup(t)

	var called bool
	r.exec = func(name string, arg ...string) *exec.Cmd {
		called = true
		return exec.Command("true")
	}

	stdIn := r.stdIn.(*bytes.Buffer)
	stdIn.WriteString(`{
			"source": {
    			:
  			}
		}`)

	if err := r.run(); err != nil {
		e := "invalid character ':' looking for beginning of object key string"
		if err.Error() != e {
			t.Errorf("Expected error message '%v', but got '%v'", e, err.Error())
		}
	} else {
		t.Error("Should have errored, but didn't")
	}
	if called {
		t.Error("Should not have called Blackduck")
	}
}
