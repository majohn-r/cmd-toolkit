package cmd_toolkit

import (
	"testing"

	"github.com/majohn-r/output"
)

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
