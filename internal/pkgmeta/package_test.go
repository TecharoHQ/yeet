package pkgmeta_test

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/TecharoHQ/yeet/internal/pkgmeta"
	"github.com/TecharoHQ/yeet/internal/yeettest"
	"github.com/goreleaser/nfpm/v2/files"
)

func TestAllCopyMethodsForEmpty(t *testing.T) {
	type testcase struct {
		name string
		fn   func(pkgmeta.Package) files.Contents
	}

	tmp := t.TempDir()

	fns := []testcase{
		{
			name: "CopyConfigFiles",
			fn:   pkgmeta.Package.CopyConfigFiles,
		},
		{
			name: "CopyDocumentation",
			fn:   pkgmeta.Package.CopyDocumentation,
		},
		{
			name: "CopyEmptyDirs",
			fn:   func(p pkgmeta.Package) files.Contents { return p.CopyEmptyDirs(0000) },
		},
		{
			name: "CopyFiles",
			fn:   pkgmeta.Package.CopyFiles,
		},
		{
			name: "CopyTree",
			fn: func(p pkgmeta.Package) files.Contents {
				path := filepath.Join(tmp, "CopyTree")

				if err := os.Mkdir(path, 0600); err != nil {
					panic(err)
				}

				cs, err := p.CopyTree(path)
				if err != nil {
					panic(err)
				}
				return cs
			},
		},
	}

	for _, tc := range fns {
		t.Run(tc.name, func(t *testing.T) {
			p := yeettest.HelloFixture(t, "", "", "", nil)
			p.EmptyDirs = nil
			p.ConfigFiles = nil
			p.Documentation = nil
			p.Files = nil

			cs := tc.fn(p)
			if len(cs) != 0 {
				t.Fatalf("expected empty output for empty input, got %v", cs)
			}
		})
	}
}

func TestFileCopyMethodsForNonEmpty(t *testing.T) {
	fns := []func(pkgmeta.Package) files.Contents{
		pkgmeta.Package.CopyConfigFiles,
		pkgmeta.Package.CopyDocumentation,
		// skip pkgmeta.Pacakge.CopyEmptyDirs
		pkgmeta.Package.CopyFiles,
		// skip pkgmeta.Package.CopyTree
	}

	for _, fn := range fns {
		t.Run(runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(), func(t *testing.T) {
			p := yeettest.HelloFixture(t, "", "", "", nil)
			p.ConfigFiles = map[string]string{"config.in": "config.out"}
			p.Documentation = map[string]string{"doc.in": "doc.out"}
			p.Files = map[string]string{"file.in": "file.out"}

			cs := fn(p)
			if len(cs) != 1 {
				t.Fatalf("expected exactly one output, got %v", cs)
			}

			src := cs[0].Source
			dst := cs[0].Destination
			srcPrefix, srcFound := strings.CutSuffix(src, ".in")
			dstPrefix, dstFound := strings.CutSuffix(filepath.Base(dst), ".out")
			if !srcFound || !dstFound || srcPrefix != dstPrefix {
				t.Fatalf("input/output mismatch, input %s, output %s", src, dst)
			}
		})
	}
}

func TestCopyEmptyDirsForNonEmpty(t *testing.T) {
	expected := "empty.d"

	p := yeettest.HelloFixture(t, "", "", "", nil)
	p.EmptyDirs = []string{expected}

	cs := p.CopyEmptyDirs(0000)

	if len(cs) != 1 {
		t.Fatalf("expected exactly one output, got %v", cs)
	}

	dst := filepath.Base(cs[0].Destination)
	if expected != dst {
		t.Fatalf("unexpected output, expected %s, got %s", expected, dst)
	}
}

func TestCopyTreeForNonEmpty(t *testing.T) {
	expected := "example.txt"
	tmp := t.TempDir()
	path := filepath.Join(tmp, expected)

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	f.Close()

	p := yeettest.HelloFixture(t, "", "", "", nil)

	cs, err := p.CopyTree(tmp)
	if err != nil {
		panic(err)
	}

	if len(cs) != 1 {
		t.Fatalf("expected exactly one output, got %v", cs)
	}

	dst := filepath.Base(cs[0].Destination)
	if expected != dst {
		t.Fatalf("unexpected output, expected %s, got %s", expected, dst)
	}

	src := filepath.Base(cs[0].Source)
	if expected != src {
		t.Fatalf("unexpected source, expected %s, got %s", expected, src)
	}
}
