package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

const coverageFile = "coverage.out"

var (
	_ = goyek.Define(goyek.Task{
		Name:  "clean",
		Usage: "delete build products",
		Action: func(a *goyek.A) {
			os.Remove(filepath.Join("..", coverageFile))
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "coverage",
		Usage: "run unit tests and produce a coverage report",
		Action: func(a *goyek.A) {
			o := makeOptions(nil)
			cmdline := fmt.Sprintf("go test -coverprofile=%s ./", coverageFile)
			if cmd.Exec(a, cmdline, o...) {
				cmdline = fmt.Sprintf("go tool cover -html=%s", coverageFile)
				cmd.Exec(a, cmdline, o...)
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "doc",
		Usage: "generate documentation",
		Action: func(a *goyek.A) {
			unifiedOutput := &bytes.Buffer{}
			cmd.Exec(a, "go doc -all", makeOptions(unifiedOutput)...)
			printOutput(unifiedOutput)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "format",
		Usage: "clean up source code formatting",
		Action: func(a *goyek.A) {
			unifiedOutput := &bytes.Buffer{}
			cmd.Exec(a, "gofmt -e -l -s -w ./", makeOptions(unifiedOutput)...)
			printOutput(unifiedOutput)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "lint",
		Usage: "run the linter on source code",
		Action: func(a *goyek.A) {
			unifiedOutput := &bytes.Buffer{}
			cmd.Exec(a, "gocritic check -enableAll ./", makeOptions(unifiedOutput)...)
			printOutput(unifiedOutput)
		},
	})

	tests = goyek.Define(goyek.Task{
		Name:  "tests",
		Usage: "run unit tests",
		Action: func(a *goyek.A) {
			unifiedOutput := &bytes.Buffer{}
			cmd.Exec(a, "go test -cover ./", makeOptions(unifiedOutput)...)
			printOutput(unifiedOutput)
		},
	})
)

func makeOptions(b *bytes.Buffer) []cmd.Option {
	var outputOptions []cmd.Option
	outputOptions = append(outputOptions, cmd.Dir(".."))
	if b != nil {
		outputOptions = append(outputOptions, cmd.Stderr(b), cmd.Stdout(b))
	}
	return outputOptions
}

func printOutput(b *bytes.Buffer) {
	output := b.String()
	if output != "" {
		fmt.Println(output)
	}
}
