package cmd_toolkit

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/majohn-r/output"
)

type CommandDescription struct {
	IsDefault   bool
	Initializer func(output.Bus, *Configuration, *flag.FlagSet) (CommandProcessor, bool)
}

var descriptions = map[string]*CommandDescription{}

// CommandProcessor defines the functions needed to run a command
type CommandProcessor interface {
	Exec(output.Bus, []string) bool
}

func AddCommandData(name string, d *CommandDescription) {
	descriptions[name] = d
}

// LogCommandStart logs the beginning of command execution; if the provided map
// contains a "command" value, then the name parameter is ignored
func LogCommandStart(o output.Bus, name string, m map[string]any) {
	if m == nil {
		m = map[string]any{}
	}
	if _, commandFound := m["command"]; !commandFound {
		m["command"] = name
	}
	o.Log(output.Info, "executing command", m)
}

func processCommand(o output.Bus, args []string) (cmd CommandProcessor, cmdArgs []string, processed bool) {
	var c *Configuration
	if c, processed = ReadConfigurationFile(o); !processed {
		return
	}
	var defaultCmd string
	if defaultCmd, processed = determineDefaultCommand(o, c.SubConfiguration("command")); processed {
		cmd, cmdArgs, processed = selectCommand(o, defaultCmd, c, args)
	}
	return
}

// ReportNothingToDo reports a user error in which a command's parameter values
// prevent the command from doing any work; the report is made to error output
// and to the log
func ReportNothingToDo(o output.Bus, cmd string, fields map[string]any) {
	o.WriteCanonicalError("You disabled all functionality for the command %q", cmd)
	o.Log(output.Error, "the user disabled all functionality", fields)
}

func determineDefaultCommand(o output.Bus, c *Configuration) (defaultCommand string, defaultFound bool) {
	// get the default command name from configuration, if it's defined
	defaultCommand, defaultFound = c.StringValue("default")
	if defaultFound {
		_, defaultFound = descriptions[defaultCommand]
		if !defaultFound {
			o.Log(output.Error, "invalid default command", map[string]any{"command": defaultCommand})
			o.WriteCanonicalError("The configuration file specifies %q as the default command. There is no such command", defaultCommand)
			defaultCommand = ""
		}
		return
	}
	switch len(descriptions) {
	case 0:
		o.Log(output.Error, "no commands registered", nil)
		o.WriteCanonicalError("A programming error has occurred - there are no commands registered!")
		return
	case 1:
		for name := range descriptions {
			defaultCommand = name
			defaultFound = true
			return
		}
	default:
		// common case: there is more than 1 command defined
		defaultCommands := make([]string, 0, len(descriptions))
		for name, d := range descriptions {
			if d.IsDefault {
				defaultCommands = append(defaultCommands, name)
			}
		}
		switch len(defaultCommands) {
		case 0:
			o.Log(output.Error, "No default command", map[string]any{"commands": describedCommandNames("")})
			o.WriteCanonicalError("A programming error has occurred - none of the defined commands is defined as the default command.")
			defaultFound = false
		case 1:
			defaultCommand = defaultCommands[0]
			defaultFound = true
		default:
			sort.Strings(defaultCommands)
			o.WriteCanonicalError("A programming error has occurred - multiple commands (%v) are defined as default commands.", defaultCommands)
			o.Log(output.Error, "multiple default commands", map[string]any{"commands": defaultCommands})
			defaultFound = false
		}
	}
	return
}

func describedCommandNames(defaultCommand string) []string {
	names := make([]string, 0, len(descriptions))
	for name := range descriptions {
		if name == defaultCommand {
			name = fmt.Sprintf("%s (default)", name)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func selectCommand(o output.Bus, defaultCmd string, c *Configuration, args []string) (cmd CommandProcessor, cmdArgs []string, commandSelected bool) {
	m := make(map[string]CommandProcessor)
	for name, description := range descriptions {
		fSet := flag.NewFlagSet(name, flag.ContinueOnError)
		cmdProcessor, initialized := description.Initializer(o, c, fSet)
		if !initialized || cmdProcessor == nil {
			return
		}
		m[name] = cmdProcessor
	}
	if len(args) < 2 {
		// no arguments at all
		cmd = m[defaultCmd]
		cmdArgs = []string{}
		commandSelected = true
		return
	}
	firstArg := args[1]
	if strings.HasPrefix(firstArg, "-") {
		// first argument is a flag
		cmd = m[defaultCmd]
		cmdArgs = args[1:]
		commandSelected = true
		return
	}
	var found bool
	cmd, found = m[firstArg]
	if !found {
		cmd = nil
		cmdArgs = nil
		names := describedCommandNames(defaultCmd)
		o.Log(output.Error, "unrecognized command", map[string]any{"command": firstArg, "commands": names})
		o.WriteCanonicalError("There is no command named %q; valid commands include [%s]", firstArg, strings.Join(names, ", "))
		return
	}
	cmdArgs = args[2:]
	commandSelected = true
	return
}
