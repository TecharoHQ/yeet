package main

import (
	"strings"
	"testing"
)

func TestFileGlob(t *testing.T) {
	for _, tt := range []struct {
		name      string
		pattern   string
		wantPanic bool
		contains  string // if non-empty, at least one result must contain this substring
		wantEmpty bool   // if true, expect zero results
	}{
		{
			name:     "recursive pattern finds go files",
			pattern:  "**/*.go",
			contains: ".go",
		},
		{
			name:     "non-recursive pattern includes main.go",
			pattern:  "*.go",
			contains: "main.go",
		},
		{
			name:      "no match returns empty slice",
			pattern:   "**/*.doesnotexist12345",
			wantEmpty: true,
		},
		{
			name:      "invalid pattern panics",
			pattern:   "[invalid",
			wantPanic: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic but did not get one")
					}
				}()
				fileGlob(tt.pattern)
				return
			}

			results := fileGlob(tt.pattern)

			if tt.wantEmpty {
				if len(results) != 0 {
					t.Errorf("expected empty results, got %d: %v", len(results), results)
				}
				return
			}

			if len(results) == 0 {
				t.Fatalf("expected results for pattern %q, got none", tt.pattern)
			}

			if tt.contains != "" {
				found := false
				for _, r := range results {
					if strings.Contains(r, tt.contains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected at least one result containing %q, got %v", tt.contains, results)
				}
			}
		})
	}
}

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
