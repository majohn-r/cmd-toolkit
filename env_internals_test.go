package cmd_toolkit

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_newEnvVarMemento(t *testing.T) {
	const varName = "VAR1"
	envVarMemento := NewEnvVarMemento(varName)
	defer envVarMemento.Restore()
	tests := map[string]struct {
		value string
		set   bool
		name  string
		want  *EnvVarMemento
	}{
		"set": {
			value: "the value",
			set:   true,
			name:  varName,
			want: &EnvVarMemento{
				name:  varName,
				value: "the value",
				set:   true,
			},
		},
		"unset": {
			name: varName,
			want: &EnvVarMemento{
				name: varName,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.set {
				_ = os.Setenv(varName, tt.value)
			} else {
				_ = os.Unsetenv(varName)
			}
			if got := NewEnvVarMemento(tt.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEnvVarMemento() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findReferences(t *testing.T) {
	tests := map[string]struct {
		s    string
		want []string
	}{
		"no references": {
			s:    "no references here, not even this: %VAR1",
			want: make([]string, 0),
		},
		"many references": {
			s:    "$VAR1 $VAR11 $VAR111 $VAR1 %VAR2% %VAR22% %VAR222% %VAR222%",
			want: []string{"$VAR111", "$VAR11", "$VAR1", "%VAR2%", "%VAR22%", "%VAR222%"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := findReferences(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findReferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_byLength_Len(t *testing.T) {
	tests := map[string]struct {
		bl   byLength
		want int
	}{
		"empty":  {want: 0},
		"plenty": {bl: byLength([]string{"a", "b", "c"}), want: 3},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.bl.Len(); got != tt.want {
				t.Errorf("byLength.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_byLength_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := map[string]struct {
		bl byLength
		args
		want bool
	}{
		"same":    {bl: byLength([]string{"$VAR1", "$VAR1"}), args: args{i: 0, j: 1}, want: false},
		"shorter": {bl: byLength([]string{"$VAR1", "$VAR11"}), args: args{i: 0, j: 1}, want: false},
		"longer":  {bl: byLength([]string{"$VAR11", "$VAR1"}), args: args{i: 0, j: 1}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.bl.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("byLength.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_byLength_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := map[string]struct {
		bl byLength
		args
		wantBl byLength
	}{
		"same": {
			bl:     byLength([]string{"$VAR1", "$VAR1"}),
			args:   args{i: 0, j: 1},
			wantBl: byLength([]string{"$VAR1", "$VAR1"}),
		},
		"different": {
			bl:     byLength([]string{"$VAR1", "$VAR11"}),
			args:   args{i: 0, j: 1},
			wantBl: byLength([]string{"$VAR11", "$VAR1"}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.bl.Swap(tt.args.i, tt.args.j)
			if !reflect.DeepEqual(tt.bl, tt.wantBl) {
				t.Errorf("byLength.Swap = %v, want %v", tt.bl, tt.wantBl)
			}
		})
	}
}

func Test_envVarMemento_restore(t *testing.T) {
	const varName = "VAR1"
	envVarMemento := NewEnvVarMemento(varName)
	defer envVarMemento.Restore()
	tests := map[string]struct {
		preValue  string
		preSet    bool
		mem       *EnvVarMemento
		wantValue string
		wantSet   bool
	}{
		"set": {
			mem:       &EnvVarMemento{name: varName, value: "the value", set: true},
			wantValue: "the value",
			wantSet:   true,
		},
		"unset": {
			preValue: "the value",
			preSet:   true,
			mem:      &EnvVarMemento{name: varName},
		},
		"overwrite": {
			preValue:  "old value",
			preSet:    true,
			mem:       &EnvVarMemento{name: varName, value: "the value", set: true},
			wantValue: "the value",
			wantSet:   true,
		},
		"redundant clear": {
			preValue: "",
			mem:      &EnvVarMemento{name: varName},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.preSet {
				_ = os.Setenv(varName, tt.preValue)
			} else {
				_ = os.Unsetenv(varName)
			}
			tt.mem.Restore()
			if gotValue, gotSet := os.LookupEnv(varName); gotValue != tt.wantValue || gotSet != tt.wantSet {
				t.Errorf("EnvVarMemento.Restore = (%q, %t) want (%q, %t)", gotValue, gotSet, tt.wantValue, tt.wantSet)
			}
		})
	}
}

func Test_createAppSpecificPath(t *testing.T) {
	tests := map[string]struct {
		applicationName string
		topDir          string
		want            string
		wantErr         bool
	}{
		"uninitialized applicationName": {
			applicationName: "",
			topDir:          "topDir",
			wantErr:         true,
		},
		"initialized applicationName": {
			applicationName: "myApp",
			topDir:          "dir",
			want:            filepath.Join("dir", "myApp"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := createAppSpecificPath(tt.topDir, tt.applicationName)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("createAppSpecificPath() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createAppSpecificPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
