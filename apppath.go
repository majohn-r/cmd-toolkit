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
func InitApplicationPath(o output.Bus) (initialized bool) {
	if value, ok := os.LookupEnv(ApplicationDataEnvVarName); ok {
		if dir, err := CreateAppSpecificPath(value); err == nil {
			applicationPath = dir
			if DirExists(applicationPath) {
				initialized = true
			} else {
				if err := Mkdir(applicationPath); err == nil {
					initialized = true
				} else {
					WriteDirectoryCreationError(o, applicationPath, err)
					o.Log(output.Error, "cannot create directory", map[string]any{
						"directory": applicationPath,
						"error":     err,
					})
				}
			}
		} else {
			o.Log(output.Error, "program error", map[string]any{"error": err})
		}
	} else {
		o.Log(output.Error, "not set", map[string]any{"environmentVariable": ApplicationDataEnvVarName})
	}
	return
}

// SetApplicationPath is used to set applicationPath to a known value; intent is for use in tesing scenarios
func SetApplicationPath(s string) (previous string) {
	previous = applicationPath
	applicationPath = s
	return
}
