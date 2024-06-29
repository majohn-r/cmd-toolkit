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
	sMap map[string]string
	bMap map[string]bool
	iMap map[string]int
	cMap map[string]*Configuration
}

// IntBounds holds the bounds for an int value which has a minimum value, a
// maximum value, and a default that lies within those bounds
type IntBounds struct {
	minValue     int
	defaultValue int
	maxValue     int
}

// DefaultConfigFileName retrieves the name of the configuration file that
// contains defaults for the commands
func DefaultConfigFileName() string {
	return defaultConfigFileName
}

func flagIndicator() string {
	return flagPrefix
}

// SetFlagIndicator sets the flag indicator to the specified value
func SetFlagIndicator(val string) {
	flagPrefix = val
}

// EmptyConfiguration creates an empty Configuration instance
func EmptyConfiguration() *Configuration {
	return &Configuration{
		bMap: make(map[string]bool),
		iMap: make(map[string]int),
		sMap: make(map[string]string),
		cMap: make(map[string]*Configuration),
	}
}

// NewConfiguration returns a Configuration instance populated as specified by
// the data parameter
func NewConfiguration(o output.Bus, data map[string]any) *Configuration {
	c := EmptyConfiguration()
	for key, v := range data {
		switch t := v.(type) {
		case string:
			c.sMap[key] = t
		case bool:
			c.bMap[key] = t
		case int:
			c.iMap[key] = t
		case map[string]any:
			c.cMap[key] = NewConfiguration(o, t)
		default:
			o.Log(output.Error, "unexpected value type", map[string]any{
				"key":   key,
				"value": v,
				"type":  fmt.Sprintf("%T", v),
			})
			o.WriteCanonicalError("The key %q, with value '%v', has an unexpected type %T", key, v, v)
			c.sMap[key] = fmt.Sprintf("%v", v)
		}
	}
	return c
}

// NewIntBounds creates a instance of IntBounds, sorting the provided value into
// reasonable fields
func NewIntBounds(v1, v2, v3 int) *IntBounds {
	v := []int{v1, v2, v3}
	sort.Ints(v)
	return &IntBounds{
		minValue:     v[0],
		defaultValue: v[1],
		maxValue:     v[2],
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
		return c, false
	}
	c = NewConfiguration(o, data)
	o.Log(output.Info, "read configuration file", map[string]any{
		"directory": path,
		"fileName":  defaultConfigFileName,
		"value":     c,
	})
	return c, true
}

// ReportInvalidConfigurationData handles errors found when attempting to parse
// a YAML configuration file, both logging the error and notifying the user of
// the error
func ReportInvalidConfigurationData(o output.Bus, s string, e error) {
	o.WriteCanonicalError("The configuration file %q contains an invalid value for %q: %v", defaultConfigFileName, s, e)
	o.Log(output.Error, "invalid content in configuration file", map[string]any{
		"section": s,
		"error":   e,
	})
}

// SetDefaultConfigFileName sets the name of the configuration file that
// contains defaults for the commands
func SetDefaultConfigFileName(s string) {
	defaultConfigFileName = s
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
	if len(c.bMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.bMap))
	}
	if len(c.iMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.iMap))
	}
	if len(c.sMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.sMap))
	}
	if len(c.cMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.cMap))
	}
	return strings.Join(s, ", ")
}

// BoolDefault returns a boolean value for a specified key
func (c *Configuration) BoolDefault(key string, defaultValue bool) (bool, error) {
	if value, valueDefined := c.bMap[key]; valueDefined {
		return value, nil
	}
	if value, valueDefined := c.iMap[key]; valueDefined {
		switch value {
		case 0:
			return false, nil
		case 1:
			return true, nil
		default:
			// note: deliberately imitating flags behavior when parsing an
			// invalid boolean
			return defaultValue, fmt.Errorf("invalid boolean value \"%d\" for %s%s: parse error", value, flagIndicator(), key)
		}
	}
	// True values may be specified as "t", "T", "true", "TRUE", or "True"
	// False values may be specified as "f", "F", "false", "FALSE", or "False"
	value, valueDefined := c.sMap[key]
	if !valueDefined {
		return defaultValue, nil
	}
	rawValue, dereferenceErr := DereferenceEnvVar(value)
	if dereferenceErr != nil {
		return defaultValue, fmt.Errorf("invalid boolean value %q for %s%s: %v", value, flagIndicator(), key, dereferenceErr)
	}
	cookedValue, e := strconv.ParseBool(rawValue)
	if e != nil {
		// note: deliberately imitating flags behavior when parsing
		// an invalid boolean
		return defaultValue, fmt.Errorf("invalid boolean value %q for %s%s: parse error", value, flagIndicator(), key)
	}
	return cookedValue, nil
}

// IntDefault returns a default value for a specified key, which may or may not
// be defined in the Configuration instance
func (c *Configuration) IntDefault(key string, b *IntBounds) (int, error) {
	if value, foundKey := c.iMap[key]; foundKey {
		return b.constrainedValue(value), nil
	}
	value, foundKey := c.sMap[key]
	if !foundKey {
		return b.Default(), nil
	}
	rawValue, dereferenceErr := DereferenceEnvVar(value)
	if dereferenceErr != nil {
		return b.Default(), fmt.Errorf("invalid value %q for flag %s%s: %v", rawValue, flagIndicator(), key, dereferenceErr)
	}
	cookedValue, e := strconv.Atoi(rawValue)
	if e != nil {
		// note: deliberately imitating flags behavior when parsing an
		// invalid int
		return b.Default(), fmt.Errorf("invalid value %q for flag %s%s: parse error", rawValue, flagIndicator(), key)
	}
	return b.constrainedValue(cookedValue), nil
}

// StringDefault returns a string value for a specified key
func (c *Configuration) StringDefault(key, defaultValue string) (string, error) {
	var dereferencedDefault string
	var dereferenceErr error
	if dereferencedDefault, dereferenceErr = DereferenceEnvVar(defaultValue); dereferenceErr != nil {
		return "", fmt.Errorf("invalid value %q for flag %s%s: %v", defaultValue, flagIndicator(), key, dereferenceErr)
	}
	value, found := c.sMap[key]
	if !found {
		return dereferencedDefault, nil
	}
	var dereferencedValue string
	if dereferencedValue, dereferenceErr = DereferenceEnvVar(value); dereferenceErr != nil {
		return "", fmt.Errorf("invalid value %q for flag %s%s: %v", value, flagIndicator(), key, dereferenceErr)
	}
	return dereferencedValue, nil
}

// StringValue returns the definition of the specified key and whether the value
// is defined
func (c *Configuration) StringValue(key string) (value string, found bool) {
	value, found = c.sMap[key]
	return
}

// SubConfiguration returns a specified sub-configuration
func (c *Configuration) SubConfiguration(key string) *Configuration {
	if configuration, found := c.cMap[key]; found {
		return configuration
	}
	return EmptyConfiguration()
}

// Default returns the default value for a bounded int
func (b *IntBounds) Default() int {
	return b.defaultValue
}

func (b *IntBounds) constrainedValue(value int) (i int) {
	switch {
	case value < b.minValue:
		i = b.minValue
	case value > b.maxValue:
		i = b.maxValue
	default:
		i = value
	}
	return
}
