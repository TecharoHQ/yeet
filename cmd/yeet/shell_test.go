package main

import "testing"

func TestBuildShellCommand(t *testing.T) {
	type args struct {
		literals []string
		exprs    []any
	}

	for _, tt := range []struct {
		name   string
		input  args
		output string
	}{
		{
			name: "basic true",
			input: args{
				literals: []string{"true"},
			},
			output: "true",
		},
		{
			name: "with args",
			input: args{
				literals: []string{"go build -o ", ""},
				exprs:    []any{"./var/anubis"},
			},
			output: `go build -o ./var/anubis`,
		},
		{
			name: "with escaped args",
			input: args{
				literals: []string{"go build -o ", ""},
				exprs:    []any{`$OUT`},
			},
			output: `go build -o '$OUT'`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := buildShellCommand(tt.input.literals, tt.input.exprs...)
			if result != tt.output {
				t.Errorf("wanted %q but got %q", tt.output, result)
			}
		})
	}
}

func TestRunShellCommand(t *testing.T) {
	_, err := runShellCommand(t.Context(), []string{"true"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunShellCommandFails(t *testing.T) {
	_, err := runShellCommand(t.Context(), []string{"false"})
	if err == nil {
		t.Fatal("false should have failed but did not")
	}
}
