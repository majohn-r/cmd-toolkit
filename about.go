package cmd_toolkit

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/majohn-r/output"
)

// the code in this file is for the "about" command, which is a common need for
// applications.

const (
	aboutCommandName = "about"
)

var (
	// these are set by InitBuildData
	appVersion        string
	buildTimestamp    string
	goVersion         string
	buildDependencies []string
	// this may be reset by SetAuthor()
	author = "Marc Johnson"
	// this should be reset by SetFirstYear()
	firstYear       = time.Now().Year()
	buildInfoReader = debug.ReadBuildInfo
)

// BuildDependencies returns information about the dependencies used to compile
// the program
func BuildDependencies() []string {
	return buildDependencies
}

// GoVersion returns the version of Go used to compile the program
func GoVersion() string {
	return goVersion
}

// InitBuildData captures information about how the program was compiled, the
// version of the program, and the timestamp for when the program was built.
func InitBuildData(version, creation string) {
	goVersion, buildDependencies = InterpretBuildData()
	appVersion = version
	buildTimestamp = creation
}

// InterpretBuildData interprets the output of calling buildInfoReader() into easily consumed forms;
// created and/or published per https://github.com/majohn-r/cmd-toolkit/issues/17
func InterpretBuildData() (goVersion string, dependencies []string) {
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

// SetFirstYear sets the first year of application development
func SetFirstYear(i int) {
	firstYear = i
}

func finalYear(o output.Bus, timestamp string) int {
	t, parseErr := time.Parse(time.RFC3339, timestamp)
	if parseErr != nil {
		o.WriteCanonicalError("The build time %q cannot be parsed: %v", timestamp, parseErr)
		o.Log(output.Error, "parse error", map[string]any{
			"error": parseErr,
			"value": timestamp,
		})
		return firstYear
	}
	return t.Year()
}

func formatBuildData() []string {
	s := make([]string, 0, 2+len(buildDependencies))
	s = append(s, BuildInformationHeader(), FormatGoVersion(goVersion))
	return append(s, FormatBuildDependencies(buildDependencies)...)
}

// BuildInformationHeader returns the canonical heading for build information.
// See https://github.com/majohn-r/cmd-toolkit/issues/17
func BuildInformationHeader() string {
	return "Build Information"
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

// FormatGoVersion returns the formatted go version; see https://github.com/majohn-r/cmd-toolkit/issues/17
func FormatGoVersion(version string) string {
	return fmt.Sprintf(" - Go version: %s", version)
}

func formatCopyright(firstYear, lastYear int, owner string) string {
	if lastYear <= firstYear {
		return fmt.Sprintf("Copyright © %d %s", firstYear, owner)
	}
	return fmt.Sprintf("Copyright © %d-%d %s", firstYear, lastYear, owner)
}

func reportAbout(o output.Bus, lines []string) {
	o.WriteConsole(strings.Join(FlowerBox(lines), "\n"))
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

type aboutCmd struct {
}

// Exec runs the command. The args parameter is ignored, and the method always
// returns true.
func (a *aboutCmd) Exec(o output.Bus, _ []string) (ok bool) {
	LogCommandStart(o, aboutCommandName, map[string]any{})
	generateAboutContent(o)
	return true
}

func generateAboutContent(o output.Bus) {
	formattedBuildData := formatBuildData()
	s := make([]string, 0, 2+len(formattedBuildData))
	name, appNameInitErr := AppName()
	if appNameInitErr != nil {
		o.Log(output.Error, "program error", map[string]any{"error": appNameInitErr})
		name = "unknown application name"
	}
	s = append(s, DecoratedAppName(name, appVersion, buildTimestamp),
		Copyright(o, firstYear, buildTimestamp, author))
	s = append(s, formattedBuildData...)
	reportAbout(o, s)
}

// DecoratedAppName returns the app name with its version and build timestamp; see https://github.com/majohn-r/cmd-toolkit/issues/17
func DecoratedAppName(applicationName, applicationVersion, timestamp string) string {
	return fmt.Sprintf("%s version %s, built on %s", applicationName, applicationVersion,
		translateTimestamp(timestamp))
}

// Copyright returns an appropriately formatted copyright statement; see https://github.com/majohn-r/cmd-toolkit/issues/17
func Copyright(o output.Bus, first int, timestamp, owner string) string {
	return formatCopyright(first, finalYear(o, timestamp), owner)
}
