package main

import (
	"github.com/goyek/goyek/v2"
	tools_build "github.com/majohn-r/tools-build"
)

const coverageFile = "coverage.out"

var (
	clean = goyek.Define(goyek.Task{
		Name:  "clean",
		Usage: "delete build products",
		Action: func(a *goyek.A) {
			tools_build.Clean([]string{coverageFile})
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "coverage",
		Usage: "run unit tests and produce a coverage report",
		Action: func(a *goyek.A) {
			tools_build.GenerateCoverageReport(a, coverageFile)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "doc",
		Usage: "generate documentation",
		Action: func(a *goyek.A) {
			tools_build.GenerateDocumentation(a, []string{"build"})
		},
	})

	format = goyek.Define(goyek.Task{
		Name:  "format",
		Usage: "clean up source code formatting",
		Action: func(a *goyek.A) {
			tools_build.Format(a)
		},
	})

	lint = goyek.Define(goyek.Task{
		Name:  "lint",
		Usage: "run the linter on source code",
		Action: func(a *goyek.A) {
			tools_build.Lint(a)
		},
	})

	nilaway = goyek.Define(goyek.Task{
		Name:  "nilaway",
		Usage: "run nilaway on source code",
		Action: func(a *goyek.A) {
			tools_build.NilAway(a)
		},
	})

	tests = goyek.Define(goyek.Task{
		Name:  "tests",
		Usage: "run unit tests",
		Action: func(a *goyek.A) {
			tools_build.UnitTests(a)
		},
	})

	vulnCheck = goyek.Define(goyek.Task{
		Name:  "vulnCheck",
		Usage: "run vulnerability check on source code",
		Action: func(a *goyek.A) {
			tools_build.VulnerabilityCheck(a)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "preCommit",
		Usage: "run all pre-commit tasks",
		Deps:  goyek.Deps{clean, lint, nilaway, format, vulnCheck, tests},
	})
)
