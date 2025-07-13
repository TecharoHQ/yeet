package mktarball

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"testing"

	"github.com/TecharoHQ/yeet/internal"
	"github.com/TecharoHQ/yeet/internal/yeettest"
)

func TestBuild(t *testing.T) {
	yeettest.BuildHello(t, Build, "1.0.0", "", "", true)
}

func TestBuildError(t *testing.T) {
	yeettest.BuildHello(t, Build, ".0.0", "", "", false)
}

func TestTimestampsNotZero(t *testing.T) {
	pkg := yeettest.BuildHello(t, Build, "1.0.0", "", "", true)

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
