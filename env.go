package cmd_toolkit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	// the name of the application
	appName string
	// regular expressions for detecting environment variable references ($VAR or %VAR%)
	unixPattern    = regexp.MustCompile(`[$][a-zA-Z_]+[a-zA-Z0-9_]*`)
	windowsPattern = regexp.MustCompile(`%[a-zA-Z_]+[a-zA-Z0-9_]*%`)
)

type byLength []string // used for sorting environment variable references

type envVarMemento struct {
	name  string
	value string
	set   bool
}

// AppName retrieves the name of the application
func AppName() (string, error) {
	if appName == "" {
		return "", errors.New("app name has not been initialized")
	}
	return appName, nil
}

// CreateAppSpecificPath creates a path string for an app-related directory
func CreateAppSpecificPath(topDir string) (string, error) {
	s, appNameInitErr := AppName()
	if appNameInitErr != nil {
		return "", appNameInitErr
	}
	return filepath.Join(topDir, s), nil
}

// DereferenceEnvVar scans a string for environment variable references, looks
// up the values of those environment variables, and replaces the references
// with their values. If one or more of the referenced environment variables are
// undefined, they are all reported in the error return
func DereferenceEnvVar(s string) (string, error) {
	refs := findReferences(s)
	if len(refs) == 0 {
		return s, nil
	}
	missing := make([]string, 0, len(refs))
	for _, ref := range refs {
		var envVar string
		switch {
		case strings.HasPrefix(ref, "$"):
			envVar = ref[1:]
		default:
			envVar = ref[1 : len(ref)-1]
		}
		value, varDefined := os.LookupEnv(envVar)
		switch varDefined {
		case true:
			s = strings.ReplaceAll(s, ref, value)
		case false:
			missing = append(missing, envVar)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return "", fmt.Errorf("missing environment variables: %v", missing)
	}
	return s, nil
}

func newEnvVarMemento(name string) *envVarMemento {
	s := &envVarMemento{name: name}
	if value, varDefined := os.LookupEnv(name); varDefined {
		s.value = value
		s.set = true
	}
	return s
}

// SetAppName sets the name of the application, returning an error if the name
// has already been set to a different value or if it is being set to an empty
// string
func SetAppName(s string) error {
	if s == "" {
		return errors.New("cannot initialize app name with an empty string")
	}
	if appName != "" && appName != s {
		return fmt.Errorf("app name has already been initialized: %s", appName)
	}
	appName = s
	return nil
}

func findReferences(s string) []string {
	squishDuplicates := func(s []string) []string {
		found := map[string]bool{}
		for _, name := range s {
			found[name] = true
		}
		keys := make([]string, 0, len(found))
		for key := range found {
			keys = append(keys, key)
		}
		return keys
	}
	matches := squishDuplicates(unixPattern.FindAllString(s, -1))
	// unix-style variable references can easily be confused: $MYAPP and
	// $MYAPP_USER are both valid, and it would be unfortunate if a string
	// containing both of them dereferenced the shorter one first. So, we sort
	// them from longest to shortest, and that's the order in which they'll be
	// dereferenced
	if len(matches) > 1 {
		sort.Sort(byLength(matches))
	}
	// but windows-style variable references, which begin and end with '%', do
	// not suffer from the same issue - %MYAPP% and %MYAPP_USER% are not going to
	// clobber each other, regardless of the order in which they are
	// dereferenced - so they don't need to be sorted
	windowsMatches := squishDuplicates(windowsPattern.FindAllString(s, -1))
	sort.Strings(windowsMatches) // sorted alphabetically for determinism in testing
	matches = append(matches, windowsMatches...)
	return matches
}

func (bl byLength) Len() int {
	return len(bl)
}

func (bl byLength) Less(i, j int) bool {
	return len(bl[i]) > len(bl[j])
}

func (bl byLength) Swap(i, j int) {
	bl[i], bl[j] = bl[j], bl[i]
}

// Restore resets a saved environment variable to its original state
func (mem *envVarMemento) Restore() {
	switch mem.set {
	case true:
		_ = os.Setenv(mem.name, mem.value)
	case false:
		_ = os.Unsetenv(mem.name)
	}
}
