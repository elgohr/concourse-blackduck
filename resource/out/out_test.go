package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
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
	if reflect.TypeOf(r.exec) != reflect.TypeOf(exec.Command) {
		t.Error("Didn't set exec correctly")
	}
}

func TestStartsBlackduckWithUsernamePassword(t *testing.T) {
	stdIn := &bytes.Buffer{}
	targetUrl := "https://BLACKDUCK"
	username := "USERNAME"
	password := "PASSWORD"
	name := "project1"
	stdIn.WriteString(fmt.Sprintf(`{
			"source": {
    			"url": "%v",
				"username": "%v",
    			"password": "%v",
				"name": "%v"
  			},
			"params": {
				"directory": "."
			}
		}`, targetUrl, username, password, name))

	var called bool
	r := Runner{
		stdIn:  stdIn,
		stdOut: &bytes.Buffer{},
		stdErr: &bytes.Buffer{},
		exec: func(command string, arg ...string) *exec.Cmd {
			called = true
			if command != "java" {
				t.Errorf("Should have started java, but started %v", command)
			}
			expectedArgs := []string{
				"-jar",
				"/opt/resource/synopsys-detect-5.3.3.jar",
				"--blackduck.url=" + targetUrl,
				"--blackduck.username=" + username,
				"--blackduck.password=" + password,
				"--detect.project.name=" + name,
				"--blackduck.trust.cert=true",
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

func TestStartsBlackduckWithToken(t *testing.T) {
	stdIn := &bytes.Buffer{}
	targetUrl := "https://BLACKDUCK"
	token := "TOKEN"
	name := "project1"
	stdIn.WriteString(fmt.Sprintf(`{
			"source": {
    			"url": "%v",
				"token": "%v",
				"name": "%v"
  			},
			"params": {
				"directory": "."
			}
		}`, targetUrl, token, name))

	var called bool
	r := Runner{
		stdIn:  stdIn,
		stdOut: &bytes.Buffer{},
		stdErr: &bytes.Buffer{},
		exec: func(command string, arg ...string) *exec.Cmd {
			called = true
			if command != "java" {
				t.Errorf("Should have started java, but started %v", command)
			}
			expectedArgs := []string{
				"-jar",
				"/opt/resource/synopsys-detect-5.3.3.jar",
				"--blackduck.url=" + targetUrl,
				"--blackduck.api.token=" + token,
				"--detect.project.name=" + name,
				"--blackduck.trust.cert=true",
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
    			"password": "password",
				"name": "project1"
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
    			"password": "password",
				"name": "project1"
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
}

func TestReturnsTheVersionAndMetaDataOfTheBlackduckScan(t *testing.T) {
	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
			b, err := ioutil.ReadFile("testdata/blackduckResponse.txt")
			if err != nil {
				t.Error(err)
			}
			command := exec.Command("echo", string(b))
			return command
		},
	}

	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
				"username": "username",
    			"password": "password",
				"name": "project1"
  			},
			"params": {
				"directory": "."
			}
		}`)

	if err := r.run(); err != nil {
		t.Error(err)
	}

	expRes := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(`{
		"version": { "ref": "d1530c19-1541-443f-8a5e-ea4e17c856a8" },
		"metadata": [
			{ "name": "name", "value": "presentations" },
			{ "name": "version", "value": "1.0.0" },
			{ "name": "url", "value": "https://my.host/api/projects/d6aed8bb-0b9a-46a2-a1ce-60101939eb10/versions/d1530c19-1541-443f-8a5e-ea4e17c856a8/components" }
		]
	}`, "\n", ""), "	", ""), " ", "")
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
				Got:   %v`, expRes, stdOut.String())
	}
}

func TestErrorsWhenTheScanFails(t *testing.T) {
	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
		exec: func(name string, arg ...string) *exec.Cmd {
			b, err := ioutil.ReadFile("testdata/blackduckError.txt")
			if err != nil {
				t.Error(err)
			}
			command := exec.Command("echo", string(b))
			return command
		},
	}

	stdIn.WriteString(`{
			"source": {
    			"url": "https://BLACKDUCK",
				"username": "username",
    			"password": "password",
				"name": "project1"
  			},
			"params": {
				"directory": "."
			}
		}`)

	if err := r.run(); err == nil {
		t.Error("Expected resource to error")
	}

	expRes := fmt.Sprintf(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(`{
		"version": { "ref": "509ce50d-b7a2-4303-89bf-bde16e4b7bef" },
		"metadata": [
			{ "name": "name", "value": "accountant" },
			{ "name": "version", "value": "%v" },
			{ "name": "url", "value": "https://my.Host/api/projects/01883e41-d4c9-420a-b41b-0ddcaadda2b5/versions/509ce50d-b7a2-4303-89bf-bde16e4b7bef/components" }
		]
	}`, "\n", ""), "	", ""), " ", ""), "Default Detect Version")
	if stdOut.String() != expRes {
		t.Errorf(`Expected: %v
				Got:   %v`, expRes, stdOut.String())
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
    			"password": "password",
				"name": "project1"
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
				"username": "username",
				"name": "project1"
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
    			"password": "password",
				"name": "project1"
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
