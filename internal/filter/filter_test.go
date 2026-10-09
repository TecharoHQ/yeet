package filter

import (
	"testing"
)

func TestMatch(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		expr   string
		method string
		goos   string
		goarch string
		want   bool
	}{
		{
			name:   "empty expression matches everything",
			expr:   "",
			method: "rpm",
			goos:   "windows",
			goarch: "ppc64le",
			want:   true,
		},
		{
			name:   "all three variables match",
			expr:   `goos == "linux" && method == "deb" && goarch == "amd64"`,
			method: "deb",
			goos:   "linux",
			goarch: "amd64",
			want:   true,
		},
		{
			name:   "wrong method",
			expr:   `goos == "linux" && method == "deb" && goarch == "amd64"`,
			method: "rpm",
			goos:   "linux",
			goarch: "amd64",
			want:   false,
		},
		{
			name:   "wrong goos",
			expr:   `goos == "linux" && method == "deb" && goarch == "amd64"`,
			method: "deb",
			goos:   "windows",
			goarch: "amd64",
			want:   false,
		},
		{
			name:   "wrong goarch",
			expr:   `goos == "linux" && method == "deb" && goarch == "amd64"`,
			method: "deb",
			goos:   "linux",
			goarch: "arm64",
			want:   false,
		},
		{
			name:   "or matches second branch",
			expr:   `method == "deb" || method == "rpm"`,
			method: "rpm",
			goos:   "linux",
			goarch: "amd64",
			want:   true,
		},
		{
			name:   "in list",
			expr:   `goarch in ["amd64", "arm64"]`,
			method: "tarball",
			goos:   "linux",
			goarch: "arm64",
			want:   true,
		},
		{
			name:   "not in list",
			expr:   `goarch in ["amd64", "arm64"]`,
			method: "tarball",
			goos:   "linux",
			goarch: "ppc64le",
			want:   false,
		},
		{
			name:   "not equal",
			expr:   `goos != "windows"`,
			method: "tarball",
			goos:   "windows",
			goarch: "amd64",
			want:   false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := New(tt.expr)
			if err != nil {
				t.Fatalf("New(%q): %v", tt.expr, err)
			}

			got, err := f.Match(tt.method, tt.goos, tt.goarch)
			if err != nil {
				t.Fatalf("Match: %v", err)
			}

			if got != tt.want {
				t.Logf("want: %v", tt.want)
				t.Logf("got:  %v", got)
				t.Errorf("wrong result for %q with method=%s goos=%s goarch=%s", tt.expr, tt.method, tt.goos, tt.goarch)
			}
		})
	}
}

func TestNewInvalid(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		expr string
	}{
		{name: "syntax error", expr: `goos ==`},
		{name: "undeclared variable", expr: `gooos == "linux"`},
		{name: "wrong operand type", expr: `goos == 1`},
		{name: "string result", expr: `goos`},
		{name: "int result", expr: `1 + 1`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := New(tt.expr); err == nil {
				t.Errorf("New(%q) did not return an error", tt.expr)
			}
		})
	}
}
