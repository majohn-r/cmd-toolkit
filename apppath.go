package cmd_toolkit

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	"github.com/majohn-r/output"
)

// The code in this file deals with the application path, which is where configuration files specific to the
// application exist. The application path is found by getting the xdg.CONFIG_HOME value; that value is assumed to be a
// writable directory, and a subdirectory whose name is the application name is looked for, and, if missing, is
// created.

var (
	applicationPath      string
	applicationNameRegex = regexp.MustCompile("^[._a-zA-Z][._a-zA-Z0-9-]+$")
)

// ApplicationPath returns the path to application-specific configuration data (typically %HOME%\AppData\Local\appName)
func ApplicationPath() string {
	return applicationPath
}

// InitApplicationPath ensures that the application path exists
func InitApplicationPath(o output.Bus, applicationName string) bool {
	value := xdg.ConfigHome
	if value == "" {
		// should not be possible unless Windows has lost its mind, in which case, you got major problems!
		o.Log(output.Error, "not set or defined", map[string]any{
			"environmentVariable":  "XDG_CONFIG_HOME",
			"Windows known folder": "localAppData",
		})
		o.ErrorPrintf(
			"Files used by %s cannot be read or written because the configuration home directory is not known.\n",
			applicationName,
		)
		o.ErrorPrintln("What to do:")
		o.ErrorPrintln("Define XDG_CONFIG_HOME, giving it a value that is a directory path, " +
			"typically %HOMEPATH%\\AppData\\Local.")
		return false
	}
	if err := Mkdir(value); err != nil {
		o.Log(output.Error, "directory check failed", map[string]any{
			"error":    err,
			"fileName": value,
		})
		o.ErrorPrintf(
			"The configuration home directory value %q is not a directory, nor can it be created as a directory.\n",
			value,
		)
		o.ErrorPrintln("What to do:")
		o.ErrorPrintln("The value of XDG_CONFIG_HOME should be a directory path, " +
			"typically %HOMEPATH%\\AppData\\Local.")
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

// AppName returns the name of the application as specified on the application's command line
func AppName() string {
	if len(os.Args) < 1 {
		return "" // this sucks, but we just don't know!
	}
	// get just the name of the app, don't care about the rest of the path
	path := os.Args[0]
	filename := filepath.Base(path)
	// special handling for file names beginning with "."
	switch {
	case filename == "." || filename == "..":
		return ""
	case strings.HasPrefix(filename, "."):
		prefix := ""
		base := filename
		for strings.HasPrefix(base, ".") {
			prefix += "."
			base = base[1:]
		}
		return prefix + base[:len(base)-len(filepath.Ext(base))]
	default:
		return filename[:len(filename)-len(filepath.Ext(filename))]
	}
}
