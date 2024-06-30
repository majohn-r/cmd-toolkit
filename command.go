package cmd_toolkit

import (
	"github.com/majohn-r/output"
)

// LogCommandStart logs the beginning of command execution; if the provided map
// contains a "command" value, then the name parameter is ignored
func LogCommandStart(o output.Bus, name string, m map[string]any) {
	commandIncluded := false
	if m == nil {
		m = map[string]any{"command": name}
		commandIncluded = true
	}
	if !commandIncluded {
		if _, commandFound := m["command"]; !commandFound {
			m["command"] = name
		}
	}
	o.Log(output.Info, "executing command", m)
}
