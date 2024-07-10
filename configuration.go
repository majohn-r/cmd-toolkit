package cmd_toolkit

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var (
	defaultConfigFileName = "defaults.yaml"
	flagPrefix            = "-"
	fileSystem            = afero.NewOsFs()
)

// FileSystem returns the current afero.Fs instance
func FileSystem() afero.Fs {
	return fileSystem
}

// AssignFileSystem sets the current afero.Fs instance and returns the original
// pre-assignment value
func AssignFileSystem(newFileSystem afero.Fs) afero.Fs {
	originalFs := fileSystem
	fileSystem = newFileSystem
	return originalFs
}

// Configuration defines the data structure for configuration information.
type Configuration struct {
	StringMap        map[string]string
	BoolMap          map[string]bool
	IntMap           map[string]int
	ConfigurationMap map[string]*Configuration
}

// IntBounds holds the bounds for an int value which has a minimum value, a
// maximum value, and a default that lies within those bounds
type IntBounds struct {
	MinValue     int
	DefaultValue int
	MaxValue     int
}

// DefaultConfigFileName retrieves the name of the configuration file that
// contains defaults for the commands
func DefaultConfigFileName() string {
	return defaultConfigFileName
}

// UnsafeSetDefaultConfigFileName sets the defaultConfigFileName variable, which is intended
// strictly for unit testing
func UnsafeSetDefaultConfigFileName(newConfigFileName string) {
	defaultConfigFileName = newConfigFileName
}

// FlagIndicator retrieves the string that indicates a command flag, typically either '-'
// or '--'
func FlagIndicator() string {
	return flagPrefix
}

// SetFlagIndicator sets the flag indicator to the specified value
func SetFlagIndicator(val string) {
	flagPrefix = val
}

// EmptyConfiguration creates an empty Configuration instance
func EmptyConfiguration() *Configuration {
	return &Configuration{
		BoolMap:          make(map[string]bool),
		IntMap:           make(map[string]int),
		StringMap:        make(map[string]string),
		ConfigurationMap: make(map[string]*Configuration),
	}
}

func newConfiguration(o output.Bus, data map[string]any) *Configuration {
	c := EmptyConfiguration()
	for key, v := range data {
		switch t := v.(type) {
		case string:
			c.StringMap[key] = t
		case bool:
			c.BoolMap[key] = t
		case int:
			c.IntMap[key] = t
		case map[string]any:
			c.ConfigurationMap[key] = newConfiguration(o, t)
		default:
			o.Log(output.Error, "unexpected value type", map[string]any{
				"key":   key,
				"value": v,
				"type":  fmt.Sprintf("%T", v),
			})
			o.WriteCanonicalError("The key %q, with value '%v', has an unexpected type %T", key, v, v)
			c.StringMap[key] = fmt.Sprintf("%v", v)
		}
	}
	return c
}

// NewIntBounds creates an instance of IntBounds, sorting the provided value into
// reasonable fields
func NewIntBounds(v1, v2, v3 int) *IntBounds {
	v := []int{v1, v2, v3}
	sort.Ints(v)
	return &IntBounds{
		MinValue:     v[0],
		DefaultValue: v[1],
		MaxValue:     v[2],
	}
}

// ReadConfigurationFile reads defaults.yaml from the specified path and returns
// a pointer to a cooked Configuration instance; if there is no such file, then
// an empty Configuration is returned and ok is true
func ReadConfigurationFile(o output.Bus) (*Configuration, bool) {
	c := EmptyConfiguration()
	path := ApplicationPath()
	file := filepath.Join(path, defaultConfigFileName)
	exists, fileError := verifyDefaultConfigFileExists(o, file)
	if fileError != nil {
		return c, false
	}
	if !exists {
		return c, true
	}
	// only probable error circumvented by verifyFileExists failure
	rawYaml, _ := afero.ReadFile(fileSystem, file)
	data := map[string]any{}
	fileError = yaml.Unmarshal(rawYaml, &data)
	if fileError != nil {
		o.Log(output.Error, "cannot unmarshal yaml content", map[string]any{
			"directory": path,
			"fileName":  defaultConfigFileName,
			"error":     fileError,
		})
		o.WriteCanonicalError("The configuration file %q is not well-formed YAML: %v", file, fileError)
		o.WriteCanonicalError("What to do:\nDelete the file %q from %q and restart the application", defaultConfigFileName, path)
		return c, false
	}
	c = newConfiguration(o, data)
	o.Log(output.Info, "read configuration file", map[string]any{
		"directory": path,
		"fileName":  defaultConfigFileName,
		"value":     c,
	})
	return c, true
}

func reportInvalidConfigurationData(o output.Bus, s string, e error) {
	o.WriteCanonicalError("The configuration file %q contains an invalid value for %q: %v", defaultConfigFileName, s, e)
	o.Log(output.Error, "invalid content in configuration file", map[string]any{
		"section": s,
		"error":   e,
	})
}

func verifyDefaultConfigFileExists(o output.Bus, path string) (exists bool, err error) {
	var f fs.FileInfo
	f, err = fileSystem.Stat(path)
	switch {
	case err == nil:
		if f.IsDir() {
			o.Log(output.Error, "file is a directory", map[string]any{
				"directory": filepath.Dir(path),
				"fileName":  filepath.Base(path),
			})
			o.WriteCanonicalError("The configuration file %q is a directory", path)
			o.WriteCanonicalError("What to do:\nDelete the directory %q from %q and restart the application", filepath.Base(path), filepath.Dir(path))
			err = fmt.Errorf("file exists but is a directory")
		} else {
			exists = true
		}
	case errors.Is(err, afero.ErrFileNotFound):
		o.Log(output.Info, "file does not exist", map[string]any{
			"directory": filepath.Dir(path),
			"fileName":  filepath.Base(path),
		})
		err = nil
	}
	return
}

func (c *Configuration) String() string {
	s := make([]string, 0, 4)
	if len(c.BoolMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.BoolMap))
	}
	if len(c.IntMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.IntMap))
	}
	if len(c.StringMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.StringMap))
	}
	if len(c.ConfigurationMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.ConfigurationMap))
	}
	return strings.Join(s, ", ")
}

// BoolDefault returns a boolean value for a specified key
func (c *Configuration) BoolDefault(key string, defaultValue bool) (bool, error) {
	if value, valueDefined := c.BoolMap[key]; valueDefined {
		return value, nil
	}
	if value, valueDefined := c.IntMap[key]; valueDefined {
		switch value {
		case 0:
			return false, nil
		case 1:
			return true, nil
		default:
			// note: deliberately imitating flags behavior when parsing an
			// invalid boolean
			return defaultValue, fmt.Errorf("invalid boolean value \"%d\" for %s%s: parse error", value, FlagIndicator(), key)
		}
	}
	// True values may be specified as "t", "T", "true", "TRUE", or "True"
	// False values may be specified as "f", "F", "false", "FALSE", or "False"
	value, valueDefined := c.StringMap[key]
	if !valueDefined {
		return defaultValue, nil
	}
	rawValue, dereferenceErr := DereferenceEnvVar(value)
	if dereferenceErr != nil {
		return defaultValue, fmt.Errorf("invalid boolean value %q for %s%s: %v", value, FlagIndicator(), key, dereferenceErr)
	}
	cookedValue, e := strconv.ParseBool(rawValue)
	if e != nil {
		// note: deliberately imitating flags behavior when parsing
		// an invalid boolean
		return defaultValue, fmt.Errorf("invalid boolean value %q for %s%s: parse error", value, FlagIndicator(), key)
	}
	return cookedValue, nil
}

// IntDefault returns a default value for a specified key, which may or may not
// be defined in the Configuration instance
func (c *Configuration) IntDefault(key string, b *IntBounds) (int, error) {
	if value, foundKey := c.IntMap[key]; foundKey {
		return b.constrainedValue(value), nil
	}
	value, foundKey := c.StringMap[key]
	if !foundKey {
		return b.DefaultValue, nil
	}
	rawValue, dereferenceErr := DereferenceEnvVar(value)
	if dereferenceErr != nil {
		return b.DefaultValue, fmt.Errorf("invalid value %q for flag %s%s: %v", rawValue, FlagIndicator(), key, dereferenceErr)
	}
	cookedValue, e := strconv.Atoi(rawValue)
	if e != nil {
		// note: deliberately imitating flags behavior when parsing an
		// invalid int
		return b.DefaultValue, fmt.Errorf("invalid value %q for flag %s%s: parse error", rawValue, FlagIndicator(), key)
	}
	return b.constrainedValue(cookedValue), nil
}

// StringDefault returns a string value for a specified key
func (c *Configuration) StringDefault(key, defaultValue string) (string, error) {
	var dereferencedDefault string
	var dereferenceErr error
	if dereferencedDefault, dereferenceErr = DereferenceEnvVar(defaultValue); dereferenceErr != nil {
		return "", fmt.Errorf("invalid value %q for flag %s%s: %v", defaultValue, FlagIndicator(), key, dereferenceErr)
	}
	value, found := c.StringMap[key]
	if !found {
		return dereferencedDefault, nil
	}
	var dereferencedValue string
	if dereferencedValue, dereferenceErr = DereferenceEnvVar(value); dereferenceErr != nil {
		return "", fmt.Errorf("invalid value %q for flag %s%s: %v", value, FlagIndicator(), key, dereferenceErr)
	}
	return dereferencedValue, nil
}

// stringValue returns the definition of the specified key and whether the value
// is defined
func (c *Configuration) stringValue(key string) (value string, found bool) {
	value, found = c.StringMap[key]
	return
}

// SubConfiguration returns a specified sub-configuration
func (c *Configuration) SubConfiguration(key string) *Configuration {
	if configuration, found := c.ConfigurationMap[key]; found {
		return configuration
	}
	return EmptyConfiguration()
}

func (b *IntBounds) constrainedValue(value int) (i int) {
	switch {
	case value < b.MinValue:
		i = b.MinValue
	case value > b.MaxValue:
		i = b.MaxValue
	default:
		i = value
	}
	return
}
