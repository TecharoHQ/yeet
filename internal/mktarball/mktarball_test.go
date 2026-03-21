package mktarball

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/TecharoHQ/yeet/internal"
	"github.com/TecharoHQ/yeet/internal/yeettest"
)

func TestBuild(t *testing.T) {
	yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
		Version: "1.0.0",
		Fatal:   true,
		GOOS:    "linux",
		GOARCH:  "amd64",
	})
}

func TestBuildError(t *testing.T) {
	yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
		Version: ".0.0",
		Fatal:   false,
		GOOS:    "linux",
		GOARCH:  "amd64",
	})
}

func TestTimestampsNotZero(t *testing.T) {
	pkg := yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
		Version: "1.0.0",
		Fatal:   true,
		GOOS:    "linux",
		GOARCH:  "amd64",
	})

	fin, err := os.Open(pkg)
	if err != nil {
		t.Fatal(err)
	}
	defer fin.Close()

	gzr, err := gzip.NewReader(fin)
	if err != nil {
		t.Fatal(err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return
		case err != nil:
			t.Fatal(err)
		}

		expect := internal.SourceEpoch()

		t.Run(header.Name, func(t *testing.T) {
			header := header
			if !header.ModTime.Equal(expect) {
				t.Errorf("file has wrong timestamp %s, wanted: %s", header.ModTime, expect)
			}
		})
	}
}

func TestWindowsBuildProducesZip(t *testing.T) {
	pkg := yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
		Version: "1.0.0",
		GOOS:    "windows",
		GOARCH:  "amd64",
		Fatal:   true,
	})

	if !strings.HasSuffix(pkg, ".zip") {
		t.Fatalf("expected .zip extension, got %s", pkg)
	}

	zr, err := zip.OpenReader(pkg)
	if err != nil {
		t.Fatalf("can't open zip: %v", err)
	}
	defer zr.Close()

	expect := internal.SourceEpoch()

	for _, f := range zr.File {
		t.Run(f.Name, func(t *testing.T) {
			if !f.Modified.Equal(expect) {
				t.Errorf("file has wrong timestamp %s, wanted: %s", f.Modified, expect)
			}
		})
	}
}
