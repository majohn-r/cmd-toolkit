package cmd_toolkit_test

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"testing"
)

func TestExitError_Error(t *testing.T) {
	tests := map[string]struct {
		e    *cmdtoolkit.ExitError
		want string
	}{
		"user": {
			e:    cmdtoolkit.NewExitUserError("cmdX"),
			want: `command "cmdX" terminated with an error (user error)`,
		},
		"code": {
			e:    cmdtoolkit.NewExitProgrammingError("cmdY"),
			want: `command "cmdY" terminated with an error (programming error)`,
		},
		"system": {
			e:    cmdtoolkit.NewExitSystemError("cmdZ"),
			want: `command "cmdZ" terminated with an error (system call failed)`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExitError_Status(t *testing.T) {
	tests := map[string]struct {
		e    *cmdtoolkit.ExitError
		want int
	}{
		"user": {
			e:    cmdtoolkit.NewExitUserError("cmdX"),
			want: 1,
		},
		"code": {
			e:    cmdtoolkit.NewExitProgrammingError("cmdY"),
			want: 2,
		},
		"system": {
			e:    cmdtoolkit.NewExitSystemError("cmdZ"),
			want: 3,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.e.Status(); got != tt.want {
				t.Errorf("Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToErrorInterface(t *testing.T) {
	tests := map[string]struct {
		e       *cmdtoolkit.ExitError
		wantErr bool
	}{
		"nil": {
			e:       nil,
			wantErr: false,
		},
		"user": {
			e:       cmdtoolkit.NewExitUserError("cmdX"),
			wantErr: true,
		},
		"code": {
			e:       cmdtoolkit.NewExitProgrammingError("cmdY"),
			wantErr: true,
		},
		"system": {
			e:       cmdtoolkit.NewExitSystemError("cmdZ"),
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := cmdtoolkit.ToErrorInterface(tt.e); (err != nil) != tt.wantErr {
				t.Errorf("ToErrorInterface() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
