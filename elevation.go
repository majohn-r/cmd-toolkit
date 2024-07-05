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
	// Scanf is the standard function for reading and scanning a line from stdin; it is used by
	// the configured exit function to allow the user control over when to close the terminal
	// window
	Scanf = fmt.Scanf
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

// ElevationControl determines whether the current process can run with elevated privileges
type ElevationControl struct {
	adminPermitted   bool
	elevated         bool
	envVarName       string
	stderrRedirected bool
	stdinRedirected  bool
	stdoutRedirected bool
}

// NewElevationControl creates a new instance of ElevationControl that does not use an environment variable to determine whether execution with elevated privileges is desired
func NewElevationControl() *ElevationControl {
	return &ElevationControl{
		adminPermitted:   true,
		elevated:         processIsElevated(),
		envVarName:       "",
		stderrRedirected: stderrState(),
		stdinRedirected:  stdinState(),
		stdoutRedirected: stdoutState(),
	}
}

// NewElevationControlWithEnvVar creates a new instance of ElevationControl that uses an
// environment variable to determine whether execution with elevated privileges is desired
func NewElevationControlWithEnvVar(envVarName string, defaultEnvVarValue bool) *ElevationControl {
	return &ElevationControl{
		adminPermitted:   environmentPermits(envVarName, defaultEnvVarValue),
		elevated:         processIsElevated(),
		envVarName:       envVarName,
		stderrRedirected: stderrState(),
		stdinRedirected:  stdinState(),
		stdoutRedirected: stdoutState(),
	}
}

// ConfigureExit sets up a new exit function that calls the original exit function, if the
// process is running with elevated privileges; otherwise, it returns the original exit function
func (ec *ElevationControl) ConfigureExit(oldExitFn func(int)) func(int) {
	returnFunc := oldExitFn
	if ec.elevated {
		originalExit := oldExitFn
		returnFunc = func(code int) {
			fmt.Printf("Exiting with exit code %d\n", code)
			var name string
			fmt.Printf("Press enter to close the window...\n")
			_, _ = Scanf("%s", &name)
			originalExit(code)
		}
	}
	return returnFunc
}

// Log logs the ElevationControl state
func (ec *ElevationControl) Log(o output.Bus, level output.Level) {
	o.Log(level, "elevation state", map[string]any{
		"elevated":             ec.elevated,
		"admin_permission":     ec.adminPermitted,
		"stderr_redirected":    ec.stderrRedirected,
		"stdin_redirected":     ec.stdinRedirected,
		"stdout_redirected":    ec.stdoutRedirected,
		"environment_variable": ec.envVarName,
	})
}

// Status returns a slice of status data suitable to display to the user
func (ec *ElevationControl) Status(appName string) []string {
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

// WillRunElevated checks whether the process can run with elevated privileges, and if so,
// attempts to do so
func (ec *ElevationControl) WillRunElevated() bool {
	if ec.canElevate() {
		// https://github.com/majohn-r/mp3repair/issues/157 if privileges can be
		// elevated successfully, return true, else assume user declined and
		// return false.
		return runElevated()
	}
	return false
}

func (ec *ElevationControl) canElevate() bool {
	if ec.elevated {
		return false // already there, so, no
	}
	if ec.redirected() {
		return false // redirection will be lost, so, no
	}
	return ec.adminPermitted // do what the environment variable says
}

func (ec *ElevationControl) describeRedirection() string {
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

func (ec *ElevationControl) redirected() bool {
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

func processIsElevated() bool {
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
