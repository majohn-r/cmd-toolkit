package cmd_toolkit

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

func TestElevationControl_canElevate(t *testing.T) {
	type fields struct {
		adminPermitted   bool
		elevated         bool
		stderrRedirected bool
		stdinRedirected  bool
		stdoutRedirected bool
	}
	tests := map[string]struct {
		fields fields
		want   bool
	}{
		"00000": {
			fields: fields{
				adminPermitted:   false,
				elevated:         false,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			want: false,
		},
		"00001": {
			fields: fields{
				adminPermitted:   false,
				elevated:         false,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			want: false,
		},
		"00010": {
			fields: fields{
				adminPermitted:   false,
				elevated:         false,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			want: false,
		},
		"00011": {
			fields: fields{
				adminPermitted:   false,
				elevated:         false,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			want: false,
		},
		"00100": {
			fields: fields{
				adminPermitted:   false,
				elevated:         false,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			want: false,
		},
		"00101": {
			fields: fields{
				adminPermitted:   false,
				elevated:         false,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			want: false,
		},
		"00110": {
			fields: fields{
				adminPermitted:   false,
				elevated:         false,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			want: false,
		},
		"00111": {
			fields: fields{
				adminPermitted:   false,
				elevated:         false,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			want: false,
		},
		"01000": {
			fields: fields{
				adminPermitted:   false,
				elevated:         true,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			want: false,
		},
		"01001": {
			fields: fields{
				adminPermitted:   false,
				elevated:         true,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			want: false,
		},
		"01010": {
			fields: fields{
				adminPermitted:   false,
				elevated:         true,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			want: false,
		},
		"01011": {
			fields: fields{
				adminPermitted:   false,
				elevated:         true,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			want: false,
		},
		"01100": {
			fields: fields{
				adminPermitted:   false,
				elevated:         true,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			want: false,
		},
		"01101": {
			fields: fields{
				adminPermitted:   false,
				elevated:         true,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			want: false,
		},
		"01110": {
			fields: fields{
				adminPermitted:   false,
				elevated:         true,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			want: false,
		},
		"01111": {
			fields: fields{
				adminPermitted:   false,
				elevated:         true,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			want: false,
		},
		"10000": {
			fields: fields{
				adminPermitted:   true,
				elevated:         false,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			want: true,
		},
		"10001": {
			fields: fields{
				adminPermitted:   true,
				elevated:         false,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			want: false,
		},
		"10010": {
			fields: fields{
				adminPermitted:   true,
				elevated:         false,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			want: false,
		},
		"10011": {
			fields: fields{
				adminPermitted:   true,
				elevated:         false,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			want: false,
		},
		"10100": {
			fields: fields{
				adminPermitted:   true,
				elevated:         false,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			want: false,
		},
		"10101": {
			fields: fields{
				adminPermitted:   true,
				elevated:         false,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			want: false,
		},
		"10110": {
			fields: fields{
				adminPermitted:   true,
				elevated:         false,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			want: false,
		},
		"10111": {
			fields: fields{
				adminPermitted:   true,
				elevated:         false,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			want: false,
		},
		"11000": {
			fields: fields{
				adminPermitted:   true,
				elevated:         true,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			want: false,
		},
		"11001": {
			fields: fields{
				adminPermitted:   true,
				elevated:         true,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			want: false,
		},
		"11010": {
			fields: fields{
				adminPermitted:   true,
				elevated:         true,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			want: false,
		},
		"11011": {
			fields: fields{
				adminPermitted:   true,
				elevated:         true,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			want: false,
		},
		"11100": {
			fields: fields{
				adminPermitted:   true,
				elevated:         true,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			want: false,
		},
		"11101": {
			fields: fields{
				adminPermitted:   true,
				elevated:         true,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			want: false,
		},
		"11110": {
			fields: fields{
				adminPermitted:   true,
				elevated:         true,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			want: false,
		},
		"11111": {
			fields: fields{
				adminPermitted:   true,
				elevated:         true,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ec := &elevationControl{
				adminPermitted:   tt.fields.adminPermitted,
				elevated:         tt.fields.elevated,
				stderrRedirected: tt.fields.stderrRedirected,
				stdinRedirected:  tt.fields.stdinRedirected,
				stdoutRedirected: tt.fields.stdoutRedirected,
			}
			if got := ec.canElevate(); got != tt.want {
				t.Errorf("canElevate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElevationControl_redirected(t *testing.T) {
	type fields struct {
		stderrRedirected bool
		stdinRedirected  bool
		stdoutRedirected bool
	}
	tests := map[string]struct {
		fields          fields
		want            bool
		wantDescription string
	}{
		"000": {
			fields: fields{
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			want:            false,
			wantDescription: "",
		},
		"001": {
			fields: fields{
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			want:            true,
			wantDescription: "stdout has been redirected",
		},
		"010": {
			fields: fields{
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			want:            true,
			wantDescription: "stdin has been redirected",
		},
		"011": {
			fields: fields{
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			want:            true,
			wantDescription: "stdin and stdout have been redirected",
		},
		"100": {
			fields: fields{
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			want:            true,
			wantDescription: "stderr has been redirected",
		},
		"101": {
			fields: fields{
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			want:            true,
			wantDescription: "stderr and stdout have been redirected",
		},
		"110": {
			fields: fields{
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			want:            true,
			wantDescription: "stderr and stdin have been redirected",
		},
		"111": {
			fields: fields{
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			want:            true,
			wantDescription: "stderr, stdin, and stdout have been redirected",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ec := &elevationControl{
				stderrRedirected: tt.fields.stderrRedirected,
				stdinRedirected:  tt.fields.stdinRedirected,
				stdoutRedirected: tt.fields.stdoutRedirected,
			}
			if got := ec.redirected(); got != tt.want {
				t.Errorf("redirected() = %v, want %v", got, tt.want)
			}
			if got := ec.describeRedirection(); got != tt.wantDescription {
				t.Errorf("describeRedirection() = %v, want %v", got, tt.wantDescription)
			}
		})
	}
}

func Test_environmentPermits(t *testing.T) {
	const varName = "MY_APP_CARES"
	envVarMemento := NewEnvVarMemento(varName)
	defer envVarMemento.Restore()
	tests := map[string]struct {
		preTest      func()
		defaultValue bool
		want         bool
	}{
		"var undefined, default false": {
			preTest: func() {
				_ = os.Unsetenv(varName)
			},
			defaultValue: false,
			want:         false,
		},
		"var undefined, default true": {
			preTest: func() {
				_ = os.Unsetenv(varName)
			},
			defaultValue: true,
			want:         true,
		},
		"var set to 'true'": {
			preTest: func() {
				_ = os.Setenv(varName, "true")
			},
			defaultValue: false,
			want:         true,
		},
		"var set to 'false'": {
			preTest: func() {
				_ = os.Setenv(varName, "false")
			},
			defaultValue: true,
			want:         false,
		},
		"var set to garbage, default false": {
			preTest: func() {
				_ = os.Setenv(varName, "junk")
			},
			defaultValue: false,
			want:         false,
		},
		"var set to garbage, default true": {
			preTest: func() {
				_ = os.Setenv(varName, "garbage")
			},
			defaultValue: true,
			want:         true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			if got := environmentPermits(varName, tt.defaultValue); got != tt.want {
				t.Errorf("environmentPermits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeArguments(t *testing.T) {
	tests := map[string]struct {
		args []string
		want string
	}{
		"no args": {
			args: nil,
			want: "",
		},
		"one arg": {
			args: []string{"myApp"},
			want: "",
		},
		"two args": {
			args: []string{"myApp", "arg1"},
			want: "arg1",
		},
		"three args": {
			args: []string{"myApp", "arg1", "arg2"},
			want: "arg1 arg2",
		},
		"four args": {
			args: []string{"myApp", "arg1", "arg2", "arg3"},
			want: "arg1 arg2 arg3",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := mergeArguments(tt.args); got != tt.want {
				t.Errorf("mergeArguments() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_redirectedDescriptor(t *testing.T) {
	originalIsTerminal := IsTerminal
	originalIsCygwinTerminal := IsCygwinTerminal
	defer func() {
		IsTerminal = originalIsTerminal
		IsCygwinTerminal = originalIsCygwinTerminal
	}()
	tests := map[string]struct {
		terminalFunc       func(uintptr) bool
		cygwinTerminalFunc func(uintptr) bool
		fd                 uintptr
		want               bool
	}{
		"is terminal": {
			terminalFunc:       func(_ uintptr) bool { return true },
			cygwinTerminalFunc: func(_ uintptr) bool { return false },
			fd:                 os.Stdin.Fd(),
			want:               false,
		},
		"is cygwin terminal": {
			terminalFunc:       func(_ uintptr) bool { return false },
			cygwinTerminalFunc: func(_ uintptr) bool { return true },
			fd:                 os.Stderr.Fd(),
			want:               false,
		},
		"is both?": {
			terminalFunc:       func(_ uintptr) bool { return true },
			cygwinTerminalFunc: func(_ uintptr) bool { return true },
			fd:                 os.Stdout.Fd(),
			want:               false,
		},
		"is neither": {
			terminalFunc:       func(_ uintptr) bool { return false },
			cygwinTerminalFunc: func(_ uintptr) bool { return false },
			fd:                 os.Stderr.Fd(),
			want:               true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			IsTerminal = tt.terminalFunc
			IsCygwinTerminal = tt.cygwinTerminalFunc
			if got := redirectedDescriptor(tt.fd); got != tt.want {
				t.Errorf("redirectedDescriptor() = %v, want %v", got, tt.want)
			}
			if got := stderrState(); got != tt.want {
				t.Errorf("stderrState() = %v, want %v", got, tt.want)
			}
			if got := stdinState(); got != tt.want {
				t.Errorf("stdinState() = %v, want %v", got, tt.want)
			}
			if got := stdoutState(); got != tt.want {
				t.Errorf("stdoutState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runElevated(t *testing.T) {
	originalShellExecute := ShellExecute
	defer func() {
		ShellExecute = originalShellExecute
	}()
	testError := fmt.Errorf("test error")
	tests := map[string]struct {
		preTest    func()
		wantError  error
		wantStatus bool
	}{
		"fail": {
			preTest: func() {
				ShellExecute = func(windows.Handle, *uint16, *uint16, *uint16, *uint16, int32) error {
					return testError
				}
			},
			wantStatus: false,
			wantError:  testError,
		},
		"success": {
			preTest: func() {
				ShellExecute = func(windows.Handle, *uint16, *uint16, *uint16, *uint16, int32) error {
					return nil
				}
			},
			wantStatus: true,
			wantError:  nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			gotError, gotStatus := runElevated()
			if gotStatus != tt.wantStatus || !errors.Is(gotError, tt.wantError) {
				t.Errorf("runElevated() = %q:%v, want %q:%v", gotError, gotStatus, tt.wantError, tt.wantStatus)
			}
		})
	}
}

func TestElevationControlImplemented(t *testing.T) {
	var ec any
	ec = &elevationControl{}
	if _, ok := ec.(ElevationControl); !ok {
		t.Errorf("&elevationControl does not implement ElevationControl")
	}
}
