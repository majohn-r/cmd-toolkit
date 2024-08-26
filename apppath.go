package cmd_toolkit

import (
	"os"
	"regexp"

	"github.com/majohn-r/output"
)

// The code in this file deals with the application path, which is where data
// files (such as configuration files) specific to the application exist. The
// application path is found by looking up the environment variable 'APPDATA';
// that value is assumed to be a writable directory, and a subdirectory whose
// name is the application name is looked for, and, if missing, is created.

var (
	applicationPath      string
	applicationNameRegex = regexp.MustCompile("^[._a-zA-Z][._a-zA-Z0-9-]+$")
)

const applicationDataEnvVarName = "APPDATA"

// ApplicationPath returns the path to application-specific data (%APPDATA%\appName)
func ApplicationPath() string {
	return applicationPath
}

// InitApplicationPath ensures that the application path exists
func InitApplicationPath(o output.Bus, applicationName string) bool {
	value, varDefined := os.LookupEnv(applicationDataEnvVarName)
	if !varDefined {
		o.Log(output.Error, "not set", map[string]any{
			"environmentVariable": applicationDataEnvVarName,
		})
		o.ErrorPrintf(
			"Files used by %s cannot be read or written because the environment variable %s has not been set.\n",
			applicationName,
			applicationDataEnvVarName,
		)
		o.ErrorPrintln("What to do:")
		o.ErrorPrintf(
			"Define %s, giving it a value that is a directory path, typically %%HOMEPATH%%\\AppData\\Roaming.\n",
			applicationDataEnvVarName,
		)
		return false
	}
	if err := Mkdir(value); err != nil {
		o.Log(output.Error, "directory check failed", map[string]any{
			"error":    err,
			"fileName": value,
		})
		o.ErrorPrintf(
			"The %s environment variable value %q is not a directory, nor can it be created as a directory.\n",
			applicationDataEnvVarName,
			value,
		)
		o.ErrorPrintln("What to do:")
		o.ErrorPrintf(
			"The value of %s should be a directory path, typically %%HOMEPATH%%\\AppData\\Roaming.\n",
			applicationDataEnvVarName,
		)
		o.ErrorPrintf("Either it should contain a subdirectory named %q.\n", applicationName)
		o.ErrorPrintln("Or, if it does not exist, it must be possible to create that subdirectory.")
		return false
	}
	dir, pathErr := createAppSpecificPath(value, applicationName)
	if pathErr != nil {
		// note: not writing anything to stderr; creating a logging path should have already caught it.
		o.Log(output.Error, "program error", map[string]any{"error": pathErr})
		return false
	}
	// Mkdir does nothing and succeeds if applicationPath is an existing
	// directory
	if mkdirErr := Mkdir(dir); mkdirErr != nil {
		writeDirectoryCreationError(o, dir, mkdirErr)
		o.Log(output.Error, "cannot create directory", map[string]any{
			"directory": dir,
			"error":     mkdirErr,
		})
		return false
	}
	applicationPath = dir
	return true
}

// SetApplicationPath is used to set applicationPath to a known value; intent is for use in testing scenarios
func SetApplicationPath(s string) (previous string) {
	previous = applicationPath
	applicationPath = s
	return
}

func isLegalApplicationName(applicationName string) bool {
	return applicationNameRegex.MatchString(applicationName)
}
