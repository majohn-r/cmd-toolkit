package main

import (
	"github.com/goyek/goyek/v3"
	toolsbuild "github.com/majohn-r/tools-build"
)

const (
	coverageFile           = "coverage.out"
	taskClean              = "clean"
	taskCoverage           = "coverage"
	taskDeadCode           = "deadcode"
	taskDoc                = "doc"
	taskFix                = "fix"
	taskFormat             = "format"
	taskLint               = "lint"
	taskNilAway            = "nilaway"
	taskTests              = "tests"
	taskUpdateDependencies = "updateDependencies"
	taskVulnCheck          = "vulnCheck"
)

var (
	clean = goyek.Define(goyek.Task{
		Name:  taskClean,
		Usage: "delete build products",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskClean) {
				toolsbuild.Clean([]string{coverageFile})
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  taskCoverage,
		Usage: "run unit tests and produce a coverage report",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskCoverage) {
				toolsbuild.GenerateCoverageReport(a, coverageFile)
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  taskDeadCode,
		Usage: "run deadcode analysis",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskDeadCode) {
				toolsbuild.Deadcode(a)
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  taskDoc,
		Usage: "generate documentation",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskDoc) {
				toolsbuild.GenerateDocumentation(a, []string{"build"})
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  taskFix,
		Usage: "run go fix",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskFix) {
				toolsbuild.GoFix(a)
			}
		},
	})

	format = goyek.Define(goyek.Task{
		Name:  taskFormat,
		Usage: "clean up source code formatting",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskFormat) {
				toolsbuild.Format(a)
			}
		},
	})

	lint = goyek.Define(goyek.Task{
		Name:  taskLint,
		Usage: "run the linter on source code",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskLint) {
				toolsbuild.Lint(a)
			}
		},
	})

	nilaway = goyek.Define(goyek.Task{
		Name:  taskNilAway,
		Usage: "run nilaway on source code",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskNilAway) {
				toolsbuild.NilAway(a)
			}
		},
	})

	tests = goyek.Define(goyek.Task{
		Name:  taskTests,
		Usage: "run unit tests",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskTests) {
				toolsbuild.UnitTests(a)
			}
		},
	})

	updateDependencies = goyek.Define(goyek.Task{
		Name:  taskUpdateDependencies,
		Usage: "update dependencies",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskUpdateDependencies) {
				toolsbuild.UpdateDependencies(a)
			}
		},
	})

	vulnCheck = goyek.Define(goyek.Task{
		Name:  taskVulnCheck,
		Usage: "run vulnerability check on source code",
		Action: func(a *goyek.A) {
			if !toolsbuild.TaskDisabled(taskVulnCheck) {
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
