package main

import (
	"os"
	"testing"
)

func TestRunCmd(t *testing.T) {
	// Place your code here
	os.Setenv("ENV_TO_DELETE", "should_be_deleted")
	defer os.Unsetenv("ENV_TO_DELETE")

	tests := []struct {
		name       string
		cmd        []string
		env        Environment
		wantReturn int
	}{
		{
			name:       "Cmd is empty",
			cmd:        []string{},
			env:        nil,
			wantReturn: 1,
		},
		{
			name:       "Success on 'env' command",
			cmd:        []string{"env"},
			env:        nil,
			wantReturn: 0,
		},
		{
			name: "Add custom environment",
			cmd:  []string{"go", "env"},
			env: Environment{
				"CUSTOM_VAR": EnvValue{Value: "custom_value", NeedRemove: false},
			},
			wantReturn: 0,
		},
		{
			name: "Check for NeedRemove",
			cmd:  []string{"env"},
			env: Environment{
				"TEST_TO_REMOVE": EnvValue{NeedRemove: true},
			},
			wantReturn: 0,
		},
		{
			name:       "Check fo not existent command",
			cmd:        []string{"some-non-existent-command-12345"},
			env:        nil,
			wantReturn: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReturn := RunCmd(tt.cmd, tt.env)
			if gotReturn != tt.wantReturn {
				t.Errorf("RunCmd() = %v, ожидалось %v", gotReturn, tt.wantReturn)
			}
		})
	}
}
