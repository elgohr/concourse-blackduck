package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestConstructsRunnerCorrectly(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"programBin", "path-to-sources"}
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
	if r.path != "path-to-sources" {
		t.Error("Expected the path to come from program args")
	}
	if r.agentDir != "/opt/resource" {
		t.Errorf("Expected the agent to be searched in /opt/resource, but was %v", r.agentDir)
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

	dir, mockFileName := prepareMockAgentFile(t)
	var called bool
	r := Runner{
		stdIn:    stdIn,
		stdOut:   &bytes.Buffer{},
		stdErr:   &bytes.Buffer{},
		agentDir: dir,
		exec: func(command string, arg ...string) *exec.Cmd {
			called = true
			if command != "java" {
				t.Errorf("Should have started java, but started %v", command)
			}
			expectedArgs := []string{
				"-jar",
				dir + "/" + mockFileName,
				"--blackduck.url=" + targetUrl,
				"--detect.project.name=" + name,
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

func TestLoadsCertificateFromBlackduckWhenConfiguredInsecure(t *testing.T) {
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
				"name": "%v",
				"insecure": true
  			},
			"params": {
				"directory": "/"
			}
		}`, targetUrl, username, password, name))

	dir, mockFileName := prepareMockAgentFile(t)
	var called bool
	r := Runner{
		stdIn:    stdIn,
		stdOut:   &bytes.Buffer{},
		stdErr:   &bytes.Buffer{},
		agentDir: dir,
		exec: func(command string, arg ...string) *exec.Cmd {
			called = true
			if command != "java" {
				t.Errorf("Should have started java, but started %v", command)
			}
			expectedArgs := []string{
				"-jar",
				dir + "/" + mockFileName,
				"--blackduck.url=" + targetUrl,
				"--detect.project.name=" + name,
				"--blackduck.username=" + username,
				"--blackduck.password=" + password,
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
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	args := []string{"program", pwd}
	stdIn := &bytes.Buffer{}
	command := exec.Command("true")

	dir, _ := prepareMockAgentFile(t)
	r := Runner{
		stdIn:    stdIn,
		stdOut:   &bytes.Buffer{},
		stdErr:   &bytes.Buffer{},
		path:     args[1],
		agentDir: dir,
		exec: func(name string, arg ...string) *exec.Cmd {
			return command
		},
	}

	directory := "/"

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
	expPath := pwd + "/" + directory
	if command.Dir != expPath {
		t.Errorf("Working dir was %v, but should be %v", command.Dir, expPath)
	}
}

func TestAddsLoggingToSubProcess(t *testing.T) {
	stdIn := &bytes.Buffer{}
	command := exec.Command("echo", "hi")
	stdErrBuf := &bytes.Buffer{}
	dir, _ := prepareMockAgentFile(t)
	r := Runner{
		stdIn:    stdIn,
		stdOut:   &bytes.Buffer{},
		stdErr:   stdErrBuf,
		agentDir: dir,
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
				"directory": "/"
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
	dir, _ := prepareMockAgentFile(t)

	r := Runner{
		stdIn:    stdIn,
		stdOut:   stdOut,
		stdErr:   &bytes.Buffer{},
		agentDir: dir,
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
				"directory": "/"
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
	dir, _ := prepareMockAgentFile(t)

	r := Runner{
		stdIn:    stdIn,
		stdOut:   stdOut,
		stdErr:   &bytes.Buffer{},
		agentDir: dir,
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
				"directory": "/"
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

func TestUsesLatestAgentVersionAvailable(t *testing.T) {
	stdIn := &bytes.Buffer{}
	targetUrl := "https://BLACKDUCK"
	username := "USERNAME"
	password := "PASSWORD"
	name := "project1"
	dir, mockFileName := prepareMockAgentFile(t)
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
		stdIn:    stdIn,
		stdOut:   &bytes.Buffer{},
		stdErr:   &bytes.Buffer{},
		agentDir: dir,
		exec: func(command string, arg ...string) *exec.Cmd {
			called = true
			if command != "java" {
				t.Errorf("Should have started java, but started %v", command)
			}
			expectedArgs := []string{
				"-jar",
				dir + "/" + mockFileName,
				"--blackduck.url=" + targetUrl,
				"--detect.project.name=" + name,
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

func TestErrorsWhenNoAgentIsAvailable(t *testing.T) {
	stdIn := &bytes.Buffer{}
	targetUrl := "https://BLACKDUCK"
	username := "USERNAME"
	password := "PASSWORD"
	name := "project1"
	dir, err := ioutil.TempDir("", "test-agent-dir")
	if err != nil {
		t.Error(err)
	}
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
		stdIn:    stdIn,
		stdOut:   &bytes.Buffer{},
		stdErr:   &bytes.Buffer{},
		agentDir: dir,
		exec: func(command string, arg ...string) *exec.Cmd {
			called = true
			return exec.Command("true")
		},
	}

	errMsg := "could not find the scanner, please open an issue on Github"
	if err := r.run(); err.Error() != errMsg {
		t.Errorf("Should have errored with %v, but was %v", errMsg, err)
	}
	if called {
		t.Error("Blackduck was trieed to be called, but wasn't there")
	}
}

func TestErrorsWhenTheAgentDirectoryIsNotPresent(t *testing.T) {
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
		stdIn:    stdIn,
		stdOut:   &bytes.Buffer{},
		stdErr:   &bytes.Buffer{},
		agentDir: "/not_here",
		exec: func(command string, arg ...string) *exec.Cmd {
			called = true
			return exec.Command("true")
		},
	}

	errMsg := "could not find the scanner, please open an issue on Github"
	if err := r.run(); err.Error() != errMsg {
		t.Errorf("Should have errored with %v, but was %v", errMsg, err)
	}
	if called {
		t.Error("Blackduck was trieed to be called, but wasn't there")
	}
}

func prepareMockAgentFile(t *testing.T) (string, string) {
	dir, err := ioutil.TempDir("", "test-agent-dir")
	if err != nil {
		t.Error(err)
	}
	mockFileName := "synopsys-detect-5.4.99.jar"
	mockAgentFile := filepath.Join(dir, mockFileName)
	if err := ioutil.WriteFile(mockAgentFile, []byte(""), 0666); err != nil {
		log.Fatal(err)
	}
	return dir, mockFileName
}
