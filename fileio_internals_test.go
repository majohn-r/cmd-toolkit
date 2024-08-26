package cmd_toolkit

import (
	"errors"
	"github.com/majohn-r/output"
	"testing"
)

func Test_logUnreadableDirectory(t *testing.T) {
	type args struct {
		s string
		e error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"basic": {
			args: args{s: "directory name", e: errors.New("directory is missing")},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='error' " +
					"directory='directory name' " +
					"error='directory is missing' " +
					"msg='cannot read directory'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			logUnreadableDirectory(o, tt.args.s, tt.args.e)
			o.Report(t, "logUnreadableDirectory()", tt.WantedRecording)
		})
	}
}

func Test_writeDirectoryCreationError(t *testing.T) {
	type args struct {
		d string
		e error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"basic": {
			args: args{d: "dirName", e: errors.New("parent directory does not exist")},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"dirName\" cannot be created: parent directory does not exist.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			writeDirectoryCreationError(o, tt.args.d, tt.args.e)
			o.Report(t, "writeDirectoryCreationError()", tt.WantedRecording)
		})
	}
}
