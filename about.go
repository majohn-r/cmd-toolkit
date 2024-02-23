package cmd_toolkit

import (
	"flag"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/majohn-r/output"
)

func init() {
	AddCommandData(aboutCommandName, &CommandDescription{IsDefault: false, Initializer: newAboutCmd})
}

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
	firstYear                                       = time.Now().Year()
	buildInfoReader func() (*debug.BuildInfo, bool) = debug.ReadBuildInfo
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

// https://github.com/majohn-r/cmd-toolkit/issues/16
func InterpretBuildData() (version string, dependencies []string) {
	if b, ok := buildInfoReader(); ok && b != nil {
		version = b.GoVersion
		for _, d := range b.Deps {
			dependencies = append(dependencies, fmt.Sprintf("%s %s", d.Path, d.Version))
		}
	} else {
		version = "unknown"
	}
	return
}

// SetAuthor is used to override the default value of author, in case someone
// besides me decides to use this library
func SetAuthor(s string) {
	author = s
}

// SetFirstYear sets the first year of application development
func SetFirstYear(i int) {
	firstYear = i
}

func finalYear(o output.Bus, timestamp string) int {
	var y = firstYear
	if t, err := time.Parse(time.RFC3339, timestamp); err != nil {
		o.WriteCanonicalError("The build time %q cannot be parsed: %v", timestamp, err)
		o.Log(output.Error, "parse error", map[string]any{
			"error": err,
			"value": timestamp,
		})
	} else {
		y = t.Year()
	}
	return y
}

func formatBuildData() []string {
	s := []string{}
	s = append(s, BuildInformationHeader(), FormatGoVersion(goVersion))
	return append(s, FormatBuildDependencies(buildDependencies)...)
}

// https://github.com/majohn-r/cmd-toolkit/issues/16
func BuildInformationHeader() string {
	return "Build Information"
}

// https://github.com/majohn-r/cmd-toolkit/issues/16
func FormatBuildDependencies(dependencies []string) []string {
	formatted := make([]string, len(dependencies))
	index := 0
	for _, dep := range dependencies {
		formatted[index] = fmt.Sprintf(" - Dependency: %s", dep)
		index++
	}
	return formatted
}

// https://github.com/majohn-r/cmd-toolkit/issues/16
func FormatGoVersion(version string) string {
	return fmt.Sprintf(" - Go version: %s", version)
}

func formatCopyright(firstYear, lastYear int, owner string) string {
	if lastYear <= firstYear {
		return fmt.Sprintf("Copyright © %d %s", firstYear, author)
	}
	return fmt.Sprintf("Copyright © %d-%d %s", firstYear, lastYear, author)
}

func reportAbout(o output.Bus, lines []string) {
	o.WriteConsole(strings.Join(FlowerBox(lines), "\n"))
}

// https://github.com/majohn-r/cmd-toolkit/issues/16
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
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("Monday, January 2 2006, 15:04:05 MST")
}

type aboutCmd struct {
}

// Exec runs the command. The args parameter is ignored, and the method always
// returns true.
func (a *aboutCmd) Exec(o output.Bus, args []string) (ok bool) {
	LogCommandStart(o, aboutCommandName, map[string]any{})
	GenerateAboutContent(o)
	return true
}

// GenerateAboutContent writes 'about' content in a pretty format
func GenerateAboutContent(o output.Bus) {
	s := []string{}
	if name, err := AppName(); err != nil {
		o.Log(output.Error, "program error", map[string]any{"error": err})
		s = append(s,
			DecoratedAppName("unknown application name", appVersion, buildTimestamp),
			Copyright(o, firstYear, buildTimestamp, author))

	} else {
		s = append(s, DecoratedAppName(name, appVersion, buildTimestamp),
			Copyright(o, firstYear, buildTimestamp, author))
	}
	s = append(s, formatBuildData()...)
	reportAbout(o, s)
}

// https://github.com/majohn-r/cmd-toolkit/issues/16
func DecoratedAppName(applicationName, applicationVersion, timestamp string) string {
	return fmt.Sprintf("%s version %s, built on %s", applicationName, applicationVersion,
		translateTimestamp(timestamp))
}

// https://github.com/majohn-r/cmd-toolkit/issues/16
func Copyright(o output.Bus, first int, timestamp, owner string) string {
	return formatCopyright(first, finalYear(o, buildTimestamp), owner)
}

func newAboutCmd(o output.Bus, _ *Configuration, _ *flag.FlagSet) (CommandProcessor, bool) {
	return &aboutCmd{}, true
}
