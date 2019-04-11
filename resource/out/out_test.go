package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

func TestStartsBlackduck(t *testing.T) {
	stdIn := &bytes.Buffer{}
	targetUrl := "https://BLACKDUCK"
	username := "USERNAME"
	password := "PASSWORD"
	stdIn.WriteString(fmt.Sprintf(`{
			"source": {
    			"url": "%v",
				"username": "%v",
    			"password": "%v"
  			},
			"params": {
				"directory": "."
			}
		}`, targetUrl, username, password))

	var called bool
	r := Runner{
		stdIn:  stdIn,
		stdOut: &bytes.Buffer{},
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
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
		},
	}

	if err := r.run(); err != nil {
		t.Error(err)
	}
	if !called {
		t.Error("Blackduck wasn't started")
	}
}

func TestSetsTheWorkingDirectoryToTheProvidedSource(t *testing.T) {
	stdIn := &bytes.Buffer{}
	command := exec.Command("true")

	r := Runner{
		stdIn:  stdIn,
		stdOut: &bytes.Buffer{},
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
			return command
		},
	}

	directory := "."

	stdIn.WriteString(fmt.Sprintf(`{
			"source": {
    			"url": "https://BLACKDUCK",
				"username": "username",
    			"password": "password"
  			},
			"params": {
				"directory": "%v"
			}
		}`, directory))

	if err := r.run(); err != nil {
		t.Error(err)
	}
	if command.Dir != directory {
		t.Error("Working dir was not set correctly")
	}
}

func TestAddsLoggingToSubProcess(t *testing.T) {
	stdIn := &bytes.Buffer{}
	command := exec.Command("true")
	r := Runner{
		stdIn:  stdIn,
		stdOut: &bytes.Buffer{},
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
			return command
		},
	}

	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
				"username": "username",
    			"password": "password"
  			},
			"params": {
				"directory": "."
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
	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("true")
		},
	}

	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
				"username": "username",
    			"password": "password"
  			},
			"params": {
				"directory": "."
			}
		}`)

	if err := r.run(); err != nil {
		t.Error(err)
	}

	if stdOut.String() != `[]` {
		t.Errorf("Expected empty array, but got %v", stdOut.String())
	}
}

func TestErrorsWhenUrlWasNotConfigured(t *testing.T) {
	stdIn := &bytes.Buffer{}
	var called bool
	r := Runner{
		stdIn:  stdIn,
		stdOut: &bytes.Buffer{},
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
			called = true
			return exec.Command("true")
		},
	}

	stdIn.WriteString(`{
			"source": {
				"username": "username",
    			"password": "password"
  			}
		}`)

	if err := r.run(); err != nil {
		e := "missing mandatory source field"
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
	stdIn := &bytes.Buffer{}
	var called bool
	r := Runner{
		stdIn:  stdIn,
		stdOut: &bytes.Buffer{},
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
			called = true
			return exec.Command("true")
		},
	}

	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
    			"password": "password"
  			}
		}`)

	if err := r.run(); err != nil {
		e := "missing mandatory source field"
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
	stdIn := &bytes.Buffer{}
	var called bool
	r := Runner{
		stdIn:  stdIn,
		stdOut: &bytes.Buffer{},
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
			called = true
			return exec.Command("true")
		},
	}

	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
				"username": "username"
  			}
		}`)

	if err := r.run(); err != nil {
		e := "missing mandatory source field"
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

func TestErrorsWhenDirectoryIsMissing(t *testing.T) {
	stdIn := &bytes.Buffer{}
	var called bool
	r := Runner{
		stdIn:  stdIn,
		stdOut: &bytes.Buffer{},
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
			called = true
			return exec.Command("true")
		},
	}

	stdIn.WriteString(`{
			"source": {
				"url": "https://BLACKDUCK",
				"username": "username",
    			"password": "password"
  			},
			"params": {}
		}`)

	if err := r.run(); err != nil {
		e := "missing mandatory params field"
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
	stdIn := &bytes.Buffer{}
	var called bool
	r := Runner{
		stdIn:  stdIn,
		stdOut: &bytes.Buffer{},
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
			called = true
			return exec.Command("true")
		},
	}

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
