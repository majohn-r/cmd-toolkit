package cmd_toolkit

import (
	"fmt"
	"github.com/majohn-r/output"
	"github.com/mattn/go-isatty"
	"golang.org/x/sys/windows"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// the vars declared below exist to make it possible for unit tests to thoroughly exercise the
// functionality in this file

var (
	// IsCygwinTerminal determines whether a particular file descriptor (e.g., os.Stdin.Fd()) is a
	// Cygwin terminal
	IsCygwinTerminal = isatty.IsCygwinTerminal
	// IsTerminal determines whether a particular file descriptor (e.g., os.Stdin.Fd()) is a
	// terminal
	IsTerminal = isatty.IsTerminal
	// GetCurrentProcessToken gets the windows token representing the current process
	GetCurrentProcessToken = windows.GetCurrentProcessToken
	// ShellExecute is the windows function that runs the specified process with elevated
	// privileges
	ShellExecute = windows.ShellExecute
	// IsElevated determines whether a specified windows token represents a process running with
	// elevated privileges
	IsElevated = windows.Token.IsElevated
)

// ElevationControl defines behavior for code pertaining to running a process with elevated
// privileges
type ElevationControl interface {
	// Log logs the elevationControl state
	Log(output.Bus, output.Level)
	// Status returns a slice of status data suitable to display to the user
	Status(string) []string
	// WillRunElevated checks whether the process can run with elevated privileges, and if so,
	// attempts to do so
	WillRunElevated() bool
}

type elevationControl struct {
	adminPermitted   bool
	elevated         bool
	envVarName       string
	stderrRedirected bool
	stdinRedirected  bool
	stdoutRedirected bool
}

// NewElevationControl creates a new instance of elevationControl that does not use an environment variable to
// determine whether execution with elevated privileges is desired
func NewElevationControl() ElevationControl {
	return &elevationControl{
		adminPermitted:   true,
		elevated:         ProcessIsElevated(),
		envVarName:       "",
		stderrRedirected: stderrState(),
		stdinRedirected:  stdinState(),
		stdoutRedirected: stdoutState(),
	}
}

// NewElevationControlWithEnvVar creates a new instance of elevationControl that uses an
// environment variable to determine whether execution with elevated privileges is desired
func NewElevationControlWithEnvVar(envVarName string, defaultEnvVarValue bool) ElevationControl {
	return &elevationControl{
		adminPermitted:   environmentPermits(envVarName, defaultEnvVarValue),
		elevated:         ProcessIsElevated(),
		envVarName:       envVarName,
		stderrRedirected: stderrState(),
		stdinRedirected:  stdinState(),
		stdoutRedirected: stdoutState(),
	}
}

// Log is the reference implementation of the ElevationControl function
func (ec *elevationControl) Log(o output.Bus, level output.Level) {
	o.Log(level, "elevation state", map[string]any{
		"elevated":             ec.elevated,
		"admin_permission":     ec.adminPermitted,
		"stderr_redirected":    ec.stderrRedirected,
		"stdin_redirected":     ec.stdinRedirected,
		"stdout_redirected":    ec.stdoutRedirected,
		"environment_variable": ec.envVarName,
	})
}

// Status is the reference implementation of the ElevationControl function
func (ec *elevationControl) Status(appName string) []string {
	results := make([]string, 0, 3)
	if ec.elevated {
		results = append(results, fmt.Sprintf("%s is running with elevated privileges", appName))
		return results
	}
	results = append(results, fmt.Sprintf("%s is not running with elevated privileges", appName))
	if ec.redirected() {
		results = append(results, ec.describeRedirection())
	}
	if !ec.adminPermitted {
		results = append(results, fmt.Sprintf("The environment variable %s evaluates as false", ec.envVarName))
	}
	return results
}

// WillRunElevated is the reference implementation of the ElevationControl function
func (ec *elevationControl) WillRunElevated() bool {
	if ec.canElevate() {
		// https://github.com/majohn-r/mp3repair/issues/157 if privileges can be
		// elevated successfully, return true, else assume user declined and
		// return false.
		return runElevated()
	}
	return false
}

func (ec *elevationControl) canElevate() bool {
	if ec.elevated {
		return false // already there, so, no
	}
	if ec.redirected() {
		return false // redirection will be lost, so, no
	}
	return ec.adminPermitted // do what the environment variable says
}

func (ec *elevationControl) describeRedirection() string {
	redirectedIO := make([]string, 0, 3)
	if ec.stderrRedirected {
		redirectedIO = append(redirectedIO, "stderr")
	}
	if ec.stdinRedirected {
		redirectedIO = append(redirectedIO, "stdin")
	}
	if ec.stdoutRedirected {
		redirectedIO = append(redirectedIO, "stdout")
	}
	result := ""
	switch len(redirectedIO) {
	case 1:
		result = fmt.Sprintf("%s has been redirected", redirectedIO[0])
	case 2:
		result = fmt.Sprintf("%s have been redirected", strings.Join(redirectedIO, " and "))
	case 3:
		result = "stderr, stdin, and stdout have been redirected"
	}
	return result
}

func (ec *elevationControl) redirected() bool {
	return ec.stderrRedirected || ec.stdinRedirected || ec.stdoutRedirected
}

func environmentPermits(varName string, defaultValue bool) bool {
	if value, varDefined := os.LookupEnv(varName); varDefined {
		// interpret value as bool
		boolValue, parseErr := strconv.ParseBool(value)
		if parseErr == nil {
			return boolValue
		}
		fmt.Fprintf(os.Stderr, "The value %q of environment variable %q is neither true nor false\n", value, varName)
	}
	return defaultValue
}

func mergeArguments(args []string) string {
	merged := ""
	if len(args) > 1 {
		merged = strings.Join(args[1:], " ")
	}
	return merged
}

// ProcessIsElevated determines whether the current process is running with elevated privileges
func ProcessIsElevated() bool {
	t := GetCurrentProcessToken()
	return IsElevated(t)
}

func redirectedDescriptor(fd uintptr) bool {
	if !IsTerminal(fd) && !IsCygwinTerminal(fd) {
		return true
	}
	return false
}

// credit: https://gist.github.com/jerblack/d0eb182cc5a1c1d92d92a4c4fcc416c6

func runElevated() (status bool) {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := mergeArguments(os.Args)
	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)
	var showCmd int32 = syscall.SW_NORMAL
	// https://github.com/majohn-r/mp3repair/issues/157 if ShellExecute returns
	// no error, assume the user accepted admin privileges and return true
	// status
	if refusedErr := ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd); refusedErr == nil {
		status = true
	}
	return
}

func stderrState() bool {
	return redirectedDescriptor(os.Stderr.Fd())
}

func stdinState() bool {
	return redirectedDescriptor(os.Stdin.Fd())
}

func stdoutState() bool {
	return redirectedDescriptor(os.Stdout.Fd())
}
