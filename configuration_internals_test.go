package cmd_toolkit

import (
	"github.com/majohn-r/output"
	"reflect"
	"testing"
)

func Test_newConfiguration(t *testing.T) {
	tests := map[string]struct {
		data map[string]any
		want *Configuration
		output.WantedRecording
	}{
		"unrecognized type": {
			data: map[string]any{
				"boolean":     true,
				"integer":     12,
				"string":      "hello",
				"problematic": 1.234,
			},
			want: &Configuration{
				BoolMap:          map[string]bool{"boolean": true},
				ConfigurationMap: map[string]*Configuration{},
				IntMap:           map[string]int{"integer": 12},
				StringMap:        map[string]string{"string": "hello", "problematic": "1.234"},
			},
			WantedRecording: output.WantedRecording{
				Error: "The key \"problematic\", with value '1.234', has an unexpected type float64.\n",
				Log:   "level='error' key='problematic' type='float64' value='1.234' msg='unexpected value type'\n",
			},
		},
		"no unrecognized types": {
			data: map[string]any{
				"boolean": true,
				"integer": 12,
				"string":  "hello",
				"complex": map[string]any{
					"another boolean": false,
					"another integer": 13,
					"another string":  "hi!",
				},
			},
			want: &Configuration{
				BoolMap: map[string]bool{"boolean": true},
				ConfigurationMap: map[string]*Configuration{
					"complex": {
						BoolMap:          map[string]bool{"another boolean": false},
						ConfigurationMap: map[string]*Configuration{},
						IntMap:           map[string]int{"another integer": 13},
						StringMap:        map[string]string{"another string": "hi!"},
					},
				},
				IntMap:    map[string]int{"integer": 12},
				StringMap: map[string]string{"string": "hello"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := newConfiguration(o, tt.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newConfiguration() = %v, want %v", got, tt.want)
			}
			o.Report(t, "newConfiguration()", tt.WantedRecording)
		})
	}
}

func TestConfiguration_stringValue(t *testing.T) {
	tests := map[string]struct {
		c         *Configuration
		key       string
		wantValue string
		wantOk    bool
	}{
		"missing": {
			c:   EmptyConfiguration(),
			key: "s",
		},
		"found": {
			c: &Configuration{
				StringMap: map[string]string{"s": "hello"},
			},
			key:       "s",
			wantValue: "hello",
			wantOk:    true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotValue, gotOk := tt.c.stringValue(tt.key)
			if gotValue != tt.wantValue {
				t.Errorf("Configuration.stringValue() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Configuration.stringValue() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
