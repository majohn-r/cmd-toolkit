package cmd_toolkit

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/majohn-r/output"
)

// Configuration defines the data structure for configuration information.
type Configuration struct {
	StringMap        map[string]string
	BoolMap          map[string]bool
	IntMap           map[string]int
	ConfigurationMap map[string]*Configuration
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
			o.ErrorPrintf("The key %q, with value '%v', has an unexpected type %T.\n", key, v, v)
			c.StringMap[key] = fmt.Sprintf("%v", v)
		}
	}
	return c
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
			return defaultValue, fmt.Errorf("invalid boolean value \"%d\" for --%s: parse error", value, key)
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
		return defaultValue, fmt.Errorf("invalid boolean value %q for --%s: %v", value, key, dereferenceErr)
	}
	cookedValue, e := strconv.ParseBool(rawValue)
	if e != nil {
		// note: deliberately imitating flags behavior when parsing
		// an invalid boolean
		return defaultValue, fmt.Errorf("invalid boolean value %q for --%s: parse error", value, key)
	}
	return cookedValue, nil
}

// IntDefault returns a default value for a specified key, which may or may not
// be defined in the Configuration instance
func (c *Configuration) IntDefault(key string, b *IntBounds) (int, error) {
	if value, foundKey := c.IntMap[key]; foundKey {
		return b.ConstrainedValue(value), nil
	}
	value, foundKey := c.StringMap[key]
	if !foundKey {
		return b.DefaultValue, nil
	}
	rawValue, dereferenceErr := DereferenceEnvVar(value)
	if dereferenceErr != nil {
		return b.DefaultValue, fmt.Errorf("invalid value %q for flag --%s: %v", rawValue, key, dereferenceErr)
	}
	cookedValue, e := strconv.Atoi(rawValue)
	if e != nil {
		// note: deliberately imitating flags behavior when parsing an
		// invalid int
		return b.DefaultValue, fmt.Errorf("invalid value %q for flag --%s: parse error", rawValue, key)
	}
	return b.ConstrainedValue(cookedValue), nil
}

// StringDefault returns a string value for a specified key
func (c *Configuration) StringDefault(key, defaultValue string) (string, error) {
	var dereferencedDefault string
	var dereferenceErr error
	if dereferencedDefault, dereferenceErr = DereferenceEnvVar(defaultValue); dereferenceErr != nil {
		return "", fmt.Errorf("invalid value %q for flag --%s: %v", defaultValue, key, dereferenceErr)
	}
	value, found := c.StringMap[key]
	if !found {
		return dereferencedDefault, nil
	}
	var dereferencedValue string
	if dereferencedValue, dereferenceErr = DereferenceEnvVar(value); dereferenceErr != nil {
		return "", fmt.Errorf("invalid value %q for flag --%s: %v", value, key, dereferenceErr)
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
