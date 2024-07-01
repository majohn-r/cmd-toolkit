package cmd_toolkit

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/majohn-r/output"
)

// the code in this file is for the "about" command, which is a common need for
// applications.

// InterpretBuildData interprets the output of calling buildInfoReader() into easily
// consumed forms; see https://github.com/majohn-r/cmd-toolkit/issues/17. for production
// callers, pass in debug.ReadBuildInfo
func InterpretBuildData(buildInfoReader func() (*debug.BuildInfo, bool)) (goVersion string, dependencies []string) {
	buildInfo, infoObtained := buildInfoReader()
	if !infoObtained || buildInfo == nil {
		goVersion = "unknown"
		return
	}
	goVersion = buildInfo.GoVersion
	dependencies = make([]string, 0, len(buildInfo.Deps))
	for _, d := range buildInfo.Deps {
		dependencies = append(dependencies, fmt.Sprintf("%s %s", d.Path, d.Version))
	}
	return
}

func finalYear(o output.Bus, timestamp string, initialYear int) int {
	t, parseErr := time.Parse(time.RFC3339, timestamp)
	if parseErr != nil {
		o.WriteCanonicalError("The build time %q cannot be parsed: %v", timestamp, parseErr)
		o.Log(output.Error, "parse error", map[string]any{
			"error": parseErr,
			"value": timestamp,
		})
		return initialYear
	}
	return t.Year()
}

// FormatBuildDependencies returns build dependency data formatted nicely;
// see https://github.com/majohn-r/cmd-toolkit/issues/17
func FormatBuildDependencies(dependencies []string) []string {
	formatted := make([]string, len(dependencies))
	index := 0
	for _, dep := range dependencies {
		formatted[index] = fmt.Sprintf(" - Dependency: %s", dep)
		index++
	}
	return formatted
}

// FormatGoVersion returns the formatted go version;
// see https://github.com/majohn-r/cmd-toolkit/issues/17cmdtoolkit.
func FormatGoVersion(version string) string {
	return fmt.Sprintf(" - Go version: %s", version)
}

func formatCopyright(firstYear, lastYear int, owner string) string {
	if lastYear <= firstYear {
		return fmt.Sprintf("Copyright © %d %s", firstYear, owner)
	}
	return fmt.Sprintf("Copyright © %d-%d %s", firstYear, lastYear, owner)
}

// FlowerBox draws a box around the provided slice of strings; see https://github.com/majohn-r/cmd-toolkit/issues/17
func FlowerBox(lines []string) []string {
	maxRunesPerLine := 0
	for _, s := range lines {
		maxRunesPerLine = max(maxRunesPerLine, len([]rune(s)))
	}
	headerRunes := make([]rune, maxRunesPerLine+4)
	headerRunes[0] = '+'
	for i := 1; i < maxRunesPerLine+3; i++ {
		headerRunes[i] = '-'
	}
	headerRunes[maxRunesPerLine+3] = '+'
	hLine := string(headerRunes)
	// size: 2 for horizontal lines + 1 for empty string at the end + 1 per line
	formattedLines := make([]string, 3+len(lines))
	formattedLines[0] = hLine
	index := 1
	for _, s := range lines {
		formattedLines[index] = fmt.Sprintf("| %s%*s |", s, maxRunesPerLine-len([]rune(s)), "")
		index++
	}
	formattedLines[index] = hLine
	formattedLines[index+1] = ""
	return formattedLines
}

func translateTimestamp(s string) string {
	t, parseErr := time.Parse(time.RFC3339, s)
	if parseErr != nil {
		return s
	}
	// https://github.com/majohn-r/cmd-toolkit/issues/18
	return t.Format("Monday, January 2 2006, 15:04:05 -0700")
}

// DecoratedAppName returns the app name with its version and build timestamp; see https://github.com/majohn-r/cmd-toolkit/issues/17
func DecoratedAppName(applicationName, applicationVersion, timestamp string) string {
	return fmt.Sprintf("%s version %s, built on %s", applicationName, applicationVersion,
		translateTimestamp(timestamp))
}

// Copyright returns an appropriately formatted copyright statement; see https://github.com/majohn-r/cmd-toolkit/issues/17
func Copyright(o output.Bus, first int, timestamp, owner string) string {
	return formatCopyright(first, finalYear(o, timestamp, first), owner)
}
