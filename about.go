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
	dependencies = make([]string, len(buildInfo.Deps))
	for k, d := range buildInfo.Deps {
		dependencies[k] = fmt.Sprintf("%s %s", d.Path, d.Version)
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
	for index, dep := range dependencies {
		formatted[index] = fmt.Sprintf(" - Dependency: %s", dep)
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

type boxChars interface {
	upperLeftCorner() rune
	upperRightCorner() rune
	lowerLeftCorner() rune
	lowerRightCorner() rune
	verticalLine() rune
	horizontalLine() rune
}

type asciiBoxChars struct{}

func (asciiBoxChars) upperLeftCorner() rune  { return '+' }
func (asciiBoxChars) upperRightCorner() rune { return '+' }
func (asciiBoxChars) lowerLeftCorner() rune  { return '+' }
func (asciiBoxChars) lowerRightCorner() rune { return '+' }
func (asciiBoxChars) verticalLine() rune     { return '|' }
func (asciiBoxChars) horizontalLine() rune   { return '-' }

type curvedBoxChars struct{}

func (curvedBoxChars) upperLeftCorner() rune  { return '╭' }
func (curvedBoxChars) upperRightCorner() rune { return '╮' }
func (curvedBoxChars) lowerLeftCorner() rune  { return '╰' }
func (curvedBoxChars) lowerRightCorner() rune { return '╯' }
func (curvedBoxChars) verticalLine() rune     { return '│' }
func (curvedBoxChars) horizontalLine() rune   { return '─' }

type doubleLineBoxChars struct{}

func (doubleLineBoxChars) upperLeftCorner() rune  { return '╔' }
func (doubleLineBoxChars) upperRightCorner() rune { return '╗' }
func (doubleLineBoxChars) lowerLeftCorner() rune  { return '╚' }
func (doubleLineBoxChars) lowerRightCorner() rune { return '╝' }
func (doubleLineBoxChars) verticalLine() rune     { return '║' }
func (doubleLineBoxChars) horizontalLine() rune   { return '═' }

type heavyLineBoxChars struct{}

func (heavyLineBoxChars) upperLeftCorner() rune  { return '┏' }
func (heavyLineBoxChars) upperRightCorner() rune { return '┓' }
func (heavyLineBoxChars) lowerLeftCorner() rune  { return '┗' }
func (heavyLineBoxChars) lowerRightCorner() rune { return '┛' }
func (heavyLineBoxChars) verticalLine() rune     { return '┃' }
func (heavyLineBoxChars) horizontalLine() rune   { return '━' }

type lightLineBoxChars struct{}

func (lightLineBoxChars) upperLeftCorner() rune  { return '┌' }
func (lightLineBoxChars) upperRightCorner() rune { return '┐' }
func (lightLineBoxChars) lowerLeftCorner() rune  { return '└' }
func (lightLineBoxChars) lowerRightCorner() rune { return '┘' }
func (lightLineBoxChars) verticalLine() rune     { return '│' }
func (lightLineBoxChars) horizontalLine() rune   { return '─' }

// FlowerBoxStyle specifies a style of drawing flower box borders
type FlowerBoxStyle uint32

const (
	// ASCIIFlowerBox uses ASCII characters ('+', '+', '+', '+', '-', and '|')
	ASCIIFlowerBox = iota
	// CurvedFlowerBox is uses light lines rounded corners ('╭', '╮', '╰', '╯', '─', and '│')
	CurvedFlowerBox
	// DoubleLinedFlowerBox uses double line characters ('╔', '╗', '╚', '╝', '═', and '║')
	DoubleLinedFlowerBox
	// HeavyLinedFlowerBox uses heavy lined characters ('┏', '┓', '┗', '┛', '━', and '┃')
	HeavyLinedFlowerBox
	// LightLinedFlowerBox uses heavy lined characters ('┌', '┐', '└', '┘', '─', and '│')
	LightLinedFlowerBox
)

func getBoxChars(style FlowerBoxStyle) boxChars {
	switch style {
	case CurvedFlowerBox:
		return curvedBoxChars{}
	case DoubleLinedFlowerBox:
		return doubleLineBoxChars{}
	case HeavyLinedFlowerBox:
		return heavyLineBoxChars{}
	case LightLinedFlowerBox:
		return lightLineBoxChars{}
	default:
		return asciiBoxChars{}
	}
}

// StyledFlowerBox draws a box around the provided slice of strings in a specified style
func StyledFlowerBox(lines []string, style FlowerBoxStyle) []string {
	maxRunesPerLine := 0
	for _, s := range lines {
		maxRunesPerLine = max(maxRunesPerLine, len([]rune(s)))
	}
	bc := getBoxChars(style)
	headerRunes := make([]rune, maxRunesPerLine+4)
	headerRunes[0] = bc.upperLeftCorner()
	for i := 1; i < maxRunesPerLine+3; i++ {
		headerRunes[i] = bc.horizontalLine()
	}
	headerRunes[maxRunesPerLine+3] = bc.upperRightCorner()
	// size: 2 for horizontal lines + 1 for empty string at the end + 1 per line
	formattedLines := make([]string, 3+len(lines))
	formattedLines[0] = string(headerRunes)
	for index, s := range lines {
		formattedLines[index+1] = fmt.Sprintf("%c %s%*s %c", bc.verticalLine(), s, maxRunesPerLine-len([]rune(s)), "", bc.verticalLine())
	}
	footerRunes := make([]rune, maxRunesPerLine+4)
	footerRunes[0] = bc.lowerLeftCorner()
	for i := 1; i < maxRunesPerLine+3; i++ {
		footerRunes[i] = bc.horizontalLine()
	}
	footerRunes[maxRunesPerLine+3] = bc.lowerRightCorner()
	formattedLines[len(lines)+1] = string(footerRunes)
	formattedLines[len(lines)+2] = ""
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
