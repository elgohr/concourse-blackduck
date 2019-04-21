package main

import (
	"bytes"
	"testing"
)

func TestDoesNothing(t *testing.T) {
	stdIn := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	r := Runner{
		stdIn:  stdIn,
		stdOut: stdOut,
		stdErr: &bytes.Buffer{},
	}

	stdIn.WriteString(`{
			"source": {
    			"url": "http://blackduck",
				"username": "username",
    			"password": "password",
				"name": "project1"
  			},
  			"version": { "ref": "0.1.1-DEVELOPMENT" }
		}`)

	if err := r.run(); err != nil {
		t.Error(err)
	}

	if stdOut.String() != `[]` {
		t.Errorf("Expected empty array, but got %v", stdOut.String())
	}
}
