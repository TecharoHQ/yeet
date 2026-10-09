package main

import (
	"errors"
	"runtime"
	"testing"

	"github.com/TecharoHQ/yeet/internal/filter"
	"github.com/TecharoHQ/yeet/internal/pkgmeta"
)

func TestPackageBuilder(t *testing.T) {
	otherArch := "arm64"
	if runtime.GOARCH == "arm64" {
		otherArch = "amd64"
	}

	for _, tt := range []struct {
		name      string
		expr      string
		method    string
		linuxOnly bool
		pkg       pkgmeta.Package
		wantBuilt bool
	}{
		{
			name:      "no filter builds the package",
			method:    "deb",
			linuxOnly: true,
			pkg:       pkgmeta.Package{Platform: "linux", Goarch: "amd64"},
			wantBuilt: true,
		},
		{
			name:      "matching filter builds the package",
			expr:      `goos == "linux" && method == "deb" && goarch == "amd64"`,
			method:    "deb",
			linuxOnly: true,
			pkg:       pkgmeta.Package{Platform: "linux", Goarch: "amd64"},
			wantBuilt: true,
		},
		{
			name:      "method mismatch skips the package",
			expr:      `method == "deb"`,
			method:    "rpm",
			linuxOnly: true,
			pkg:       pkgmeta.Package{Platform: "linux", Goarch: "amd64"},
		},
		{
			name:   "goos mismatch skips the package",
			expr:   `goos == "linux"`,
			method: "tarball",
			pkg:    pkgmeta.Package{Platform: "windows", Goarch: "amd64"},
		},
		{
			name:      "goarch mismatch skips the package",
			expr:      `goarch == "amd64"`,
			method:    "deb",
			linuxOnly: true,
			pkg:       pkgmeta.Package{Platform: "linux", Goarch: "ppc64le"},
		},
		{
			name:      "empty platform filters as linux",
			expr:      `goos == "linux"`,
			method:    "tarball",
			pkg:       pkgmeta.Package{Goarch: "amd64"},
			wantBuilt: true,
		},
		{
			name:      "empty goarch filters as host goarch",
			expr:      `goarch == "` + runtime.GOARCH + `"`,
			method:    "tarball",
			pkg:       pkgmeta.Package{Platform: "linux"},
			wantBuilt: true,
		},
		{
			name:   "empty goarch does not match another goarch",
			expr:   `goarch == "` + otherArch + `"`,
			method: "tarball",
			pkg:    pkgmeta.Package{Platform: "linux"},
		},
		{
			name:      "linux-only method skips windows without a filter",
			method:    "deb",
			linuxOnly: true,
			pkg:       pkgmeta.Package{Platform: "windows", Goarch: "amd64"},
		},
		{
			name:      "method that is not linux-only builds windows",
			method:    "tarball",
			pkg:       pkgmeta.Package{Platform: "windows", Goarch: "amd64"},
			wantBuilt: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f, err := filter.New(tt.expr)
			if err != nil {
				t.Fatalf("filter.New(%q): %v", tt.expr, err)
			}

			const builtPath = "var/built.pkg"
			built := false
			fn := func(pkgmeta.Package) (string, error) {
				built = true
				return builtPath, nil
			}
			if tt.linuxOnly {
				fn = linuxOnly(fn)
			}
			build := packageBuilder(f, tt.method, fn)

			got := build(tt.pkg)

			if built != tt.wantBuilt {
				t.Logf("want: %v", tt.wantBuilt)
				t.Logf("got:  %v", built)
				t.Error("build function called state is wrong")
			}

			want := ""
			if tt.wantBuilt {
				want = builtPath
			}
			if got != want {
				t.Logf("want: %q", want)
				t.Logf("got:  %q", got)
				t.Error("wrong output path")
			}
		})
	}
}

func TestPackageBuilderPanicsOnBuildError(t *testing.T) {
	f, err := filter.New("")
	if err != nil {
		t.Fatal(err)
	}

	wantErr := errors.New("build failed")
	build := packageBuilder(f, "deb", func(pkgmeta.Package) (string, error) {
		return "", wantErr
	})

	defer func() {
		r := recover()
		err, ok := r.(error)
		if !ok || !errors.Is(err, wantErr) {
			t.Logf("want: %v", wantErr)
			t.Logf("got:  %v", r)
			t.Error("got wrong panic value")
		}
	}()

	build(pkgmeta.Package{Platform: "linux", Goarch: "amd64"})
}
