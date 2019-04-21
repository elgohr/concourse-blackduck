package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	runner := NewRunner()
	if err := runner.run(); err != nil {
		log.Fatalln(err)
	}
}

type Runner struct {
	stdIn  io.Reader
	stdOut io.Writer
	stdErr io.Writer
}

func NewRunner() Runner {
	return Runner{
		stdIn:  os.Stdin,
		stdOut: os.Stdout,
		stdErr: os.Stderr,
	}
}

func (r *Runner) run() error {
	fmt.Fprintf(r.stdOut, `[]`)
	return nil
}
