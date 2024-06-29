package cmd_toolkit

import (
	"flag"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

// TODO: when there are no consumers of AddCommandData, delete this function
func TestAddCommandData(t *testing.T) {
	savedDescriptions := descriptions
	defer func() {
		descriptions = savedDescriptions
	}()
	tests := map[string]struct {
		name string
		d    *CommandDescription
	}{
		"typical": {
			name: "myCommand",
			d: &CommandDescription{
				IsDefault: true,
				Initializer: func(b output.Bus, c *Configuration, fs *flag.FlagSet) (CommandProcessor, bool) {
					return nil, false
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			descriptions = map[string]*CommandDescription{}
			AddCommandData(tt.name, tt.d)
			if got := len(descriptions); got != 1 {
				t.Errorf("AddCommandData() got %d want 1", got)
			}
			if got, descriptionFound := descriptions[tt.name]; !descriptionFound {
				t.Errorf("AddCommandData() could not find %q", tt.name)
			} else if !reflect.DeepEqual(got, tt.d) {
				t.Errorf("AddCommandData() retrieved %v want %v", got, tt.d)
			}
		})
	}
}

func TestLogCommandStart(t *testing.T) {
	type args struct {
		name string
		m    map[string]any
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"bad map": {
			args:            args{name: "nasty command", m: nil},
			WantedRecording: output.WantedRecording{Log: "level='info' command='nasty command' msg='executing command'\n"},
		},
		"empty map": {
			args:            args{name: "niceCommand", m: map[string]any{}},
			WantedRecording: output.WantedRecording{Log: "level='info' command='niceCommand' msg='executing command'\n"},
		},
		"busy map": {
			args: args{
				name: "", // note, this is ignored because the map contains a "command" entry
				m: map[string]any{
					"command": "BusyCommand",
					"-flag1":  "value1",
					"-flag2":  25,
				},
			},
			WantedRecording: output.WantedRecording{Log: "level='info' -flag1='value1' -flag2='25' command='BusyCommand' msg='executing command'\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			LogCommandStart(o, tt.args.name, tt.args.m)
			o.Report(t, "LogCommandStart()", tt.WantedRecording)
		})
	}
}

type dummyCommand struct{}

func (d *dummyCommand) Exec(_ output.Bus, _ []string) bool {
	return true
}

func TestProcessCommand(t *testing.T) {
	originalApplicationPath := applicationPath
	originalDescriptions := descriptions
	originalFileSystem := fileSystem
	defer func() {
		applicationPath = originalApplicationPath
		descriptions = originalDescriptions
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	tests := map[string]struct {
		preTest     func()
		args        []string
		wantCmd     bool
		wantCmdArgs []string
		wantOk      bool
		output.WantedRecording
	}{
		"fail to get configuration file": {
			preTest: func() {
				applicationPath = "badConfigFile"
				_ = fileSystem.Mkdir(applicationPath, StdDirPermissions)
				fileName := filepath.Join(applicationPath, defaultConfigFileName)
				_ = afero.WriteFile(fileSystem, fileName, []byte{1, 2, 3}, StdFilePermissions) // this will not read well as YAML
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"badConfigFile\\\\defaults.yaml\" is not well-formed YAML: yaml: control characters are not allowed.\n",
				Log:   "level='error' directory='badConfigFile' error='yaml: control characters are not allowed' fileName='defaults.yaml' msg='cannot unmarshal yaml content'\n",
			},
		},
		"non-existent configuration file, no commands registered": {
			preTest: func() {
				applicationPath = "noConfigFile"
				_ = fileSystem.Mkdir(applicationPath, StdDirPermissions)
				descriptions = map[string]*CommandDescription{}
			},
			WantedRecording: output.WantedRecording{
				Error: "A programming error has occurred - there are no commands registered!\n",
				Log: "" +
					"level='info' directory='noConfigFile' fileName='defaults.yaml' msg='file does not exist'\n" +
					"level='error'  msg='no commands registered'\n",
			},
		},
		"non-existent configuration file, bad command initialization": {
			preTest: func() {
				applicationPath = "noConfigFile"
				_ = fileSystem.Mkdir(applicationPath, StdDirPermissions)
				descriptions = map[string]*CommandDescription{
					"about": {
						Initializer: func(_ output.Bus, _ *Configuration, _ *flag.FlagSet) (CommandProcessor, bool) {
							return nil, false
						},
					},
				}
			},
			WantedRecording: output.WantedRecording{Log: "level='info' directory='noConfigFile' fileName='defaults.yaml' msg='file does not exist'\n"},
		},
		"success": {
			preTest: func() {
				applicationPath = "noConfigFile"
				_ = fileSystem.Mkdir(applicationPath, StdDirPermissions)
				descriptions = map[string]*CommandDescription{
					"about": {
						Initializer: func(_ output.Bus, _ *Configuration, _ *flag.FlagSet) (CommandProcessor, bool) {
							return &dummyCommand{}, true
						},
					},
				}
			},
			args:            []string{"cmd", "-flag1", "-flag2"},
			wantCmd:         true,
			wantCmdArgs:     []string{"-flag1", "-flag2"},
			wantOk:          true,
			WantedRecording: output.WantedRecording{Log: "level='info' directory='noConfigFile' fileName='defaults.yaml' msg='file does not exist'\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			o := output.NewRecorder()
			gotCmd, gotCmdArgs, gotOk := processCommand(o, tt.args)
			if (gotCmd != nil) != tt.wantCmd {
				t.Errorf("processCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotCmdArgs, tt.wantCmdArgs) {
				t.Errorf("processCommand() gotCmdArgs = %v, want %v", gotCmdArgs, tt.wantCmdArgs)
			}
			if gotOk != tt.wantOk {
				t.Errorf("processCommand() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "processCommand()", tt.WantedRecording)
		})
	}
}

// TODO: delete when there are no external callers
func TestReportNothingToDo(t *testing.T) {
	type args struct {
		cmd    string
		fields map[string]any
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"sample": {
			args: args{cmd: "someCommand", fields: map[string]any{"-flag1": "foo", "-flag2": 43}},
			WantedRecording: output.WantedRecording{
				Error: "You disabled all functionality for the command \"someCommand\".\n",
				Log:   "level='error' -flag1='foo' -flag2='43' msg='the user disabled all functionality'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			ReportNothingToDo(o, tt.args.cmd, tt.args.fields)
			o.Report(t, "ReportNothingToDo()", tt.WantedRecording)
		})
	}
}

func Test_determineDefaultCommand(t *testing.T) {
	savedDescriptions := descriptions
	defer func() {
		descriptions = savedDescriptions
	}()
	tests := map[string]struct {
		descriptions       map[string]*CommandDescription
		c                  *Configuration
		wantDefaultCommand string
		wantOk             bool
		output.WantedRecording
	}{
		"configured good default": {
			descriptions: map[string]*CommandDescription{"about": {}},
			c: newConfiguration(output.NewNilBus(), map[string]any{
				"default": "about",
			}),
			wantDefaultCommand: "about",
			wantOk:             true,
		},
		"configured bad default": {
			descriptions: map[string]*CommandDescription{"about": {}},
			c: newConfiguration(output.NewNilBus(), map[string]any{
				"default": "help",
			}),
			WantedRecording: output.WantedRecording{
				Error: "The configuration file specifies \"help\" as the default command. There is no such command.\n",
				Log:   "level='error' command='help' msg='invalid default command'\n",
			},
		},
		"no configured default, no commands defined": {
			descriptions: map[string]*CommandDescription{},
			c:            EmptyConfiguration(),
			WantedRecording: output.WantedRecording{
				Error: "A programming error has occurred - there are no commands registered!\n",
				Log:   "level='error'  msg='no commands registered'\n",
			},
		},
		"no configured default, exactly one command defined": {
			descriptions:       map[string]*CommandDescription{"help": {}},
			c:                  EmptyConfiguration(),
			wantDefaultCommand: "help",
			wantOk:             true,
		},
		"no configured default, multiple commands defined, no default": {
			descriptions: map[string]*CommandDescription{
				"help":  {},
				"about": {},
			},
			c: EmptyConfiguration(),
			WantedRecording: output.WantedRecording{
				Error: "A programming error has occurred - none of the defined commands is defined as the default command.\n",
				Log:   "level='error' commands='[about help]' msg='No default command'\n",
			},
		},
		"no configured default, multiple commands defined, one default": {
			descriptions: map[string]*CommandDescription{
				"help":  {IsDefault: true},
				"about": {},
			},
			c:                  EmptyConfiguration(),
			wantDefaultCommand: "help",
			wantOk:             true,
		},
		"no configured default, multiple commands defined, multiple defaults": {
			descriptions: map[string]*CommandDescription{
				"help":  {IsDefault: true},
				"about": {IsDefault: true},
				"other": {},
			},
			c: EmptyConfiguration(),
			WantedRecording: output.WantedRecording{
				Error: "A programming error has occurred - multiple commands ([about help]) are defined as default commands.\n",
				Log:   "level='error' commands='[about help]' msg='multiple default commands'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			descriptions = tt.descriptions
			o := output.NewRecorder()
			gotDefaultCommand, gotOk := determineDefaultCommand(o, tt.c)
			if gotDefaultCommand != tt.wantDefaultCommand {
				t.Errorf("determineDefaultCommand() gotDefaultCommand = %v, want %v", gotDefaultCommand, tt.wantDefaultCommand)
			}
			if gotOk != tt.wantOk {
				t.Errorf("determineDefaultCommand() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "determineDefaultCommand()", tt.WantedRecording)
		})
	}
}

func Test_describedCommandNames(t *testing.T) {
	savedDescriptions := descriptions
	defer func() {
		descriptions = savedDescriptions
	}()
	tests := map[string]struct {
		defaultCommand string
		descriptions   map[string]*CommandDescription
		want           []string
	}{
		"no default": {
			descriptions: map[string]*CommandDescription{
				"someCommand": {},
				"about":       {},
				"help":        {},
			},
			want: []string{"about", "help", "someCommand"},
		},
		"with default": {
			defaultCommand: "help",
			descriptions: map[string]*CommandDescription{
				"someCommand": {},
				"about":       {},
				"help":        {},
			},
			want: []string{"about", "help (default)", "someCommand"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			descriptions = tt.descriptions
			if got := describedCommandNames(tt.defaultCommand); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("describedCommandNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_selectCommand(t *testing.T) {
	savedDescriptions := descriptions
	defer func() {
		descriptions = savedDescriptions
	}()
	type args struct {
		defaultCmd string
		c          *Configuration
		args       []string
	}
	tests := map[string]struct {
		descriptions map[string]*CommandDescription
		args
		wantCmd     bool
		wantCmdArgs []string
		wantOk      bool
		output.WantedRecording
	}{
		"one or more commands fail to initialize": {
			descriptions: map[string]*CommandDescription{
				"about": {
					Initializer: func(output.Bus, *Configuration, *flag.FlagSet) (CommandProcessor, bool) {
						return nil, false
					},
				},
			},
			args: args{},
		},
		"all commands initialized, no arguments": {
			descriptions: map[string]*CommandDescription{
				"about": {
					Initializer: func(output.Bus, *Configuration, *flag.FlagSet) (CommandProcessor, bool) {
						return &dummyCommand{}, true
					},
				},
			},
			args:        args{defaultCmd: "about", c: EmptyConfiguration(), args: []string{}},
			wantCmd:     true,
			wantCmdArgs: []string{},
			wantOk:      true,
		},
		"all commands initialized, one argument": {
			descriptions: map[string]*CommandDescription{
				"about": {
					Initializer: func(output.Bus, *Configuration, *flag.FlagSet) (CommandProcessor, bool) {
						return &dummyCommand{}, true
					},
				},
			},
			args:        args{defaultCmd: "about", c: EmptyConfiguration(), args: []string{"cmd"}},
			wantCmd:     true,
			wantCmdArgs: []string{},
			wantOk:      true,
		},
		"all commands initialized, first real argument is a flag": {
			descriptions: map[string]*CommandDescription{
				"about": {
					Initializer: func(output.Bus, *Configuration, *flag.FlagSet) (CommandProcessor, bool) {
						return &dummyCommand{}, true
					},
				},
			},
			args:        args{defaultCmd: "about", c: EmptyConfiguration(), args: []string{"cmd", "-flag1=true", "-flag2=14"}},
			wantCmd:     true,
			wantCmdArgs: []string{"-flag1=true", "-flag2=14"},
			wantOk:      true,
		},
		"all commands initialized, first real argument is a non-existent command": {
			descriptions: map[string]*CommandDescription{
				"about": {
					Initializer: func(output.Bus, *Configuration, *flag.FlagSet) (CommandProcessor, bool) {
						return &dummyCommand{}, true
					},
				},
			},
			args: args{defaultCmd: "about", c: EmptyConfiguration(), args: []string{"cmd", "nonCommand", "-flag2=14"}},
			WantedRecording: output.WantedRecording{
				Error: "There is no command named \"nonCommand\"; valid commands include [about (default)].\n",
				Log:   "level='error' command='nonCommand' commands='[about (default)]' msg='unrecognized command'\n",
			},
		},
		"all commands initialized, first real argument is a real command": {
			descriptions: map[string]*CommandDescription{
				"about": {
					Initializer: func(output.Bus, *Configuration, *flag.FlagSet) (CommandProcessor, bool) {
						return &dummyCommand{}, true
					},
				},
				"help": {
					Initializer: func(output.Bus, *Configuration, *flag.FlagSet) (CommandProcessor, bool) {
						return &dummyCommand{}, true
					},
				},
			},
			args:        args{defaultCmd: "about", c: EmptyConfiguration(), args: []string{"cmd", "help", "-flag2=14"}},
			wantCmd:     true,
			wantCmdArgs: []string{"-flag2=14"},
			wantOk:      true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			descriptions = tt.descriptions
			o := output.NewRecorder()
			gotCmd, gotCmdArgs, gotOk := selectCommand(o, tt.args.defaultCmd, tt.args.c, tt.args.args)
			if (gotCmd != nil) != tt.wantCmd {
				t.Errorf("selectCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotCmdArgs, tt.wantCmdArgs) {
				t.Errorf("selectCommand() gotCmdArgs = %v, want %v", gotCmdArgs, tt.wantCmdArgs)
			}
			if gotOk != tt.wantOk {
				t.Errorf("selectCommand() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "selectCommand()", tt.WantedRecording)
		})
	}
}
