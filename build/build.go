package main

import (
	"github.com/goyek/goyek/v2"
	toolsbuild "github.com/majohn-r/tools-build"
)

const coverageFile = "coverage.out"

var (
	clean = goyek.Define(goyek.Task{
		Name:  "clean",
		Usage: "delete build products",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled("clean") {
				toolsbuild.Clean([]string{coverageFile})
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "coverage",
		Usage: "run unit tests and produce a coverage report",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled("coverage") {
				toolsbuild.GenerateCoverageReport(a, coverageFile)
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "deadcode",
		Usage: "run deadcode analysis",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled("deadcode") {
				toolsbuild.Deadcode(a)
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "doc",
		Usage: "generate documentation",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled("doc") {
				toolsbuild.GenerateDocumentation(a, []string{"build"})
			}
		},
	})

	format = goyek.Define(goyek.Task{
		Name:  "format",
		Usage: "clean up source code formatting",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled("format") {
				toolsbuild.Format(a)
			}
		},
	})

	lint = goyek.Define(goyek.Task{
		Name:  "lint",
		Usage: "run the linter on source code",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled("lint") {
				toolsbuild.Lint(a)
			}
		},
	})

	nilaway = goyek.Define(goyek.Task{
		Name:  "nilaway",
		Usage: "run nilaway on source code",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled("nilaway") {
				toolsbuild.NilAway(a)
			}
		},
	})

	tests = goyek.Define(goyek.Task{
		Name:  "tests",
		Usage: "run unit tests",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled("tests") {
				toolsbuild.UnitTests(a)
			}
		},
	})

	updateDependencies = goyek.Define(goyek.Task{
		Name:  "updateDependencies",
		Usage: "update dependencies",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled("updateDependencies") {
				toolsbuild.UpdateDependencies(a)
			}
		},
	})

	vulnCheck = goyek.Define(goyek.Task{
		Name:  "vulnCheck",
		Usage: "run vulnerability check on source code",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled("vulnCheck") {
				toolsbuild.VulnerabilityCheck(a)
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "preCommit",
		Usage: "run all pre-commit tasks",
		Deps: goyek.Deps{
			clean,
			updateDependencies,
			lint,
			nilaway,
			format,
			vulnCheck,
			tests,
		},
	})
)
