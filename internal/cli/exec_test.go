package cli

import (
	"reflect"
	"testing"
)

func TestSplitExecArgs(t *testing.T) {
	isProfile := func(s string) bool { return s == "dev" || s == "vertex" }

	tests := []struct {
		name      string
		args      []string
		wantNames []string
		wantRest  []string
		wantErr   bool
	}{
		{
			name:      "single profile, no separator",
			args:      []string{"dev", "python", "app.py"},
			wantNames: []string{"dev"},
			wantRest:  []string{"python", "app.py"},
		},
		{
			name:      "single profile with separator",
			args:      []string{"dev", "--", "python", "app.py"},
			wantNames: []string{"dev"},
			wantRest:  []string{"python", "app.py"},
		},
		{
			name:      "multiple profiles with separator",
			args:      []string{"dev", "vertex", "--", "python", "app.py"},
			wantNames: []string{"dev", "vertex"},
			wantRest:  []string{"python", "app.py"},
		},
		{
			name:    "multiple profiles without separator is rejected",
			args:    []string{"dev", "vertex", "python"},
			wantErr: true,
		},
		{
			name:    "missing command after separator",
			args:    []string{"dev", "--"},
			wantErr: true,
		},
		{
			name:    "missing command in shorthand",
			args:    []string{"dev"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, rest, err := splitExecArgs(tt.args, isProfile)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got names=%v rest=%v", names, rest)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(names, tt.wantNames) {
				t.Errorf("names = %v, want %v", names, tt.wantNames)
			}
			if !reflect.DeepEqual(rest, tt.wantRest) {
				t.Errorf("rest = %v, want %v", rest, tt.wantRest)
			}
		})
	}
}
