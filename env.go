package cmd_toolkit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	// regular expressions for detecting environment variable references ($VAR or %VAR%)
	unixPattern    = regexp.MustCompile(`[$][a-zA-Z_]+[a-zA-Z0-9_]*`)
	windowsPattern = regexp.MustCompile(`%[a-zA-Z_]+[a-zA-Z0-9_]*%`)
)

type byLength []string // used for sorting environment variable references

// EnvVarMemento encapsulates information about an environment variable: its name, its value, and whether it was set
// when accessed
type EnvVarMemento struct {
	name  string
	value string
	set   bool
}

func createAppSpecificPath(topDir, applicationName string) (string, error) {
	if !isLegalApplicationName(applicationName) {
		return "", fmt.Errorf("application name %q is not valid", applicationName)
	}
	return filepath.Join(topDir, applicationName), nil
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

// NewEnvVarMemento creates a new instance of EnvVarMemento based on the state of the
// environment variable 'name'
func NewEnvVarMemento(name string) *EnvVarMemento {
	s := &EnvVarMemento{name: name}
	if value, varDefined := os.LookupEnv(name); varDefined {
		s.value = value
		s.set = true
	}
	return s
}

func findReferences(s string) []string {
	squishDuplicates := func(s []string) []string {
		found := map[string]bool{}
		for _, name := range s {
			found[name] = true
		}
		keys := make([]string, len(found))
		index := 0
		for key := range found {
			keys[index] = key
			index++
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

// Len returns the length of the byLength instance - in other words, the number of strings
func (bl byLength) Len() int {
	return len(bl)
}

// Less is a function that returns true if the length of string 'j' is less than that of
// string 'i'
func (bl byLength) Less(i, j int) bool {
	return len(bl[i]) > len(bl[j])
}

// Swap swaps the ith and jth elements
func (bl byLength) Swap(i, j int) {
	bl[i], bl[j] = bl[j], bl[i]
}

// Restore restores the environment variable encapsulated by the EnvVarMemento instance to
// its original state
func (mem *EnvVarMemento) Restore() {
	switch mem.set {
	case true:
		_ = os.Setenv(mem.name, mem.value)
	case false:
		_ = os.Unsetenv(mem.name)
	}
}
