package cmd_toolkit

import (
	"fmt"
	"reflect"
)

type errorCode int

const (
	_            errorCode = iota
	userError              // user did something silly
	programError           // program code error
	systemError            // unexpected system call failures, such as file not found
)

var (
	strStatusMap = map[errorCode]string{
		userError:    "user error",
		programError: "programming error",
		systemError:  "system call failed",
	}
)

// ExitError encapsulates the command in which the error occurred, and the kind of error
// that terminated the command
type ExitError struct {
	errorCode
	command string
}

// NewExitUserError creates an ExitError suitable for a user error to have terminated the
// command
func NewExitUserError(cmd string) *ExitError {
	return &ExitError{command: cmd, errorCode: userError}
}

// NewExitProgrammingError creates an ExitError suitable for a programming error to have
// terminated the command
func NewExitProgrammingError(cmd string) *ExitError {
	return &ExitError{command: cmd, errorCode: programError}
}

// NewExitSystemError creates an ExitError suitable for a system call failure to have terminated
// the command
func NewExitSystemError(cmd string) *ExitError {
	return &ExitError{command: cmd, errorCode: systemError}
}

// Error generates a description of what happened
func (e *ExitError) Error() string {
	return fmt.Sprintf("command %q terminated with an error (%s)", e.command, strStatusMap[e.errorCode])
}

// Status returns an int suitable to pass to os.Exit
func (e *ExitError) Status() int {
	return int(e.errorCode)
}

// ToErrorInterface translates a nil ExitError instances to a nil error instances; non-nil
// ExitError instances are returned as is
func ToErrorInterface(e *ExitError) error {
	if reflect.ValueOf(e).IsNil() {
		return nil
	}
	return e
}
