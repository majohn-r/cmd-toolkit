package cmd_toolkit_test

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"reflect"
	"testing"
)

func TestNewIntBounds(t *testing.T) {
	type args struct {
		v1 int
		v2 int
		v3 int
	}
	tests := map[string]struct {
		args
		want *cmdtoolkit.IntBounds
	}{
		"low, middle, high": {
			args: args{
				v1: 1,
				v2: 2,
				v3: 3,
			},
			want: &cmdtoolkit.IntBounds{
				MinValue:     1,
				DefaultValue: 2,
				MaxValue:     3,
			},
		},
		"low, high, middle": {
			args: args{
				v1: 1,
				v2: 3,
				v3: 2,
			},
			want: &cmdtoolkit.IntBounds{
				MinValue:     1,
				DefaultValue: 2,
				MaxValue:     3,
			},
		},
		"middle, low, high": {
			args: args{
				v1: 2,
				v2: 1,
				v3: 3,
			},
			want: &cmdtoolkit.IntBounds{
				MinValue:     1,
				DefaultValue: 2,
				MaxValue:     3,
			},
		},
		"middle, high, low": {
			args: args{
				v1: 2,
				v2: 3,
				v3: 1,
			},
			want: &cmdtoolkit.IntBounds{
				MinValue:     1,
				DefaultValue: 2,
				MaxValue:     3,
			},
		},
		"high, low, middle": {
			args: args{
				v1: 3,
				v2: 1,
				v3: 2,
			},
			want: &cmdtoolkit.IntBounds{
				MinValue:     1,
				DefaultValue: 2,
				MaxValue:     3,
			},
		},
		"high, middle, low": {
			args: args{
				v1: 3,
				v2: 2,
				v3: 1,
			},
			want: &cmdtoolkit.IntBounds{
				MinValue:     1,
				DefaultValue: 2,
				MaxValue:     3,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.NewIntBounds(tt.args.v1, tt.args.v2, tt.args.v3); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIntBounds() = %v, want %v", got, tt.want)
			}
		})
	}
}
