package cmd_toolkit

import "testing"

func TestIntBounds_constrainedValue(t *testing.T) {
	type fields struct {
		MinValue     int
		DefaultValue int
		MaxValue     int
	}
	tests := map[string]struct {
		fields fields
		value  int
		wantI  int
	}{
		"low": {
			fields: fields{
				MinValue:     -2,
				DefaultValue: 10,
				MaxValue:     45,
			},
			value: -3,
			wantI: -2,
		},
		"middle": {
			fields: fields{
				MinValue:     -2,
				DefaultValue: 10,
				MaxValue:     45,
			},
			value: -1,
			wantI: -1,
		},
		"high": {
			fields: fields{
				MinValue:     -2,
				DefaultValue: 10,
				MaxValue:     45,
			},
			value: 46,
			wantI: 45,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := &IntBounds{
				MinValue:     tt.fields.MinValue,
				DefaultValue: tt.fields.DefaultValue,
				MaxValue:     tt.fields.MaxValue,
			}
			if gotI := b.constrainedValue(tt.value); gotI != tt.wantI {
				t.Errorf("constrainedValue() = %v, want %v", gotI, tt.wantI)
			}
		})
	}
}
