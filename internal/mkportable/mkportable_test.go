package mkportable

import (
	"path/filepath"
	"testing"

	"github.com/TecharoHQ/yeet/internal/gpgtest"
	"github.com/TecharoHQ/yeet/internal/pkgmeta"
	"github.com/TecharoHQ/yeet/internal/yeettest"
)

func TestBuild(t *testing.T) {
	const method = "test"

	myBuild := func(p pkgmeta.Package) (string, error) {
		return build(p, method)
	}

	keyFname := filepath.Join(t.TempDir(), "foo.gpg")
	keyID, err := gpgtest.MakeKey(t.Context(), keyFname)
	if err != nil {
		t.Fatal(err)
	}

	fname := yeettest.BuildHello(t, myBuild, yeettest.BuildHelloInput{
		Version:  "1.0.0",
		KeyFname: keyFname,
		KeyID:    keyID,
		Fatal:    true,

		GOOS:   "linux",
		GOARCH: "amd64",
	})
	t.Log(fname)
}

func TestBuildError(t *testing.T) {
	const method = "test"

	myBuild := func(p pkgmeta.Package) (string, error) {
		return build(p, method)
	}

	yeettest.BuildHello(t, myBuild, yeettest.BuildHelloInput{
		Version: ".0.0",
		Fatal:   false,
	})
}
