package cmd_toolkit

import (
	"flag"
	"fmt"
	"runtime/debug"
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
	if b, ok := buildInfoReader(); ok {
		goVersion = b.GoVersion
		for _, d := range b.Deps {
			buildDependencies = append(buildDependencies, fmt.Sprintf("%s %s", d.Path, d.Version))
		}
	} else {
		goVersion = "unknown"
		buildDependencies = nil
	}
	appVersion = version
	buildTimestamp = creation
}

// SetAuthor is used to override the default value of author, in case someone
// besides me decides to use this library
func SetAuthor(s string) {
	author = s
}

func setFirstYear(i int) {
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
	var s []string
	s = append(s, "Build Information", fmt.Sprintf(" - Go version: %s", goVersion))
	for _, dep := range buildDependencies {
		s = append(s, fmt.Sprintf(" - Dependency: %s", dep))
	}
	return s
}

func formatCopyright(firstYear, lastYear int) string {
	if lastYear <= firstYear {
		return fmt.Sprintf("Copyright © %d %s", firstYear, author)
	}
	return fmt.Sprintf("Copyright © %d-%d %s", firstYear, lastYear, author)
}

func reportAbout(o output.Bus, lines []string) {
	max := 0
	for _, s := range lines {
		if len(s) > max {
			max = len([]rune(s))
		}
	}
	var formatted []string
	for _, s := range lines {
		b := make([]rune, max)
		i := 0
		for _, s1 := range s {
			b[i] = s1
			i++
		}
		for ; i < max; i++ {
			b[i] = ' '
		}
		formatted = append(formatted, string(b))
	}
	headerRunes := make([]rune, max)
	for i := 0; i < max; i++ {
		headerRunes[i] = '-'
	}
	header := string(headerRunes)
	o.WriteConsole("+-%s-+\n", header)
	for _, s := range formatted {
		o.WriteConsole("| %s |\n", s)
	}
	o.WriteConsole("+-%s-+\n", header)
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
	var s []string
	if name, err := AppName(); err != nil {
		o.Log(output.Error, "program error", map[string]any{"error": err})
		s = append(s,
			fmt.Sprintf("unknown application name version %s, built on %s", appVersion, translateTimestamp(buildTimestamp)),
			formatCopyright(firstYear, finalYear(o, buildTimestamp)))

	} else {
		s = append(s,
			fmt.Sprintf("%s version %s, built on %s", name, appVersion, translateTimestamp(buildTimestamp)),
			formatCopyright(firstYear, finalYear(o, buildTimestamp)))
	}
	s = append(s, formatBuildData()...)
	reportAbout(o, s)
	return true
}

func newAboutCmd(o output.Bus, _ *Configuration, _ *flag.FlagSet) (CommandProcessor, bool) {
	return &aboutCmd{}, true
}
