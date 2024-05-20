package cmd_toolkit

import (
	"os"

	"github.com/majohn-r/output"
)

// The code in this file deals with the application path, which is where data
// files (such as configuration files) specific to the application exist. The
// application path is found by looking up the environment variable 'APPDATA';
// that value is assumed to be a writable directory, and a subdirectory whose
// name is the application name is looked for, and, if missing, is created.

var applicationPath string

// ApplicationDataEnvVarName is the name of the environment variable used to
// read/write application-specific files that are intended to have some
// permanence.
const ApplicationDataEnvVarName = "APPDATA"

// ApplicationPath returns the path to application-specific data (%APPDATA%\appname)
func ApplicationPath() string {
	return applicationPath
}

// InitApplicationPath ensures that the application path exists
func InitApplicationPath(o output.Bus) bool {
	value, varDefined := os.LookupEnv(ApplicationDataEnvVarName)
	if !varDefined {
		o.Log(output.Error, "not set", map[string]any{"environmentVariable": ApplicationDataEnvVarName})
		return false
	}
	dir, pathErr := CreateAppSpecificPath(value)
	if pathErr != nil {
		o.Log(output.Error, "program error", map[string]any{"error": pathErr})
		return false
	}
	// Mkdir does nothing and succeeds if applicationPath is an existing
	// directory
	if mkdirErr := Mkdir(dir); mkdirErr != nil {
		WriteDirectoryCreationError(o, dir, mkdirErr)
		o.Log(output.Error, "cannot create directory", map[string]any{
			"directory": dir,
			"error":     mkdirErr,
		})
		return false
	}
	applicationPath = dir
	return true
}

// SetApplicationPath is used to set applicationPath to a known value; intent is for use in tesing scenarios
func SetApplicationPath(s string) (previous string) {
	previous = applicationPath
	applicationPath = s
	return
}
