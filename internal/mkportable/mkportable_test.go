package mkportable

import (
	"path/filepath"
	"testing"

	"github.com/TecharoHQ/yeet/internal/gpgtest"
	"github.com/TecharoHQ/yeet/internal/pkgmeta"
	"github.com/TecharoHQ/yeet/internal/yeettest"
)

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

func TestAllMethods(t *testing.T) {
	keyFname := filepath.Join(t.TempDir(), "foo.gpg")
	keyID, err := gpgtest.MakeKey(t.Context(), keyFname)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name    string
		builder func(pkgmeta.Package) (string, error)
	}{
		{"confext", Confext},
		{"portable", Portable},
		{"sysext", Sysext},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fname := yeettest.BuildHello(t, tt.builder, yeettest.BuildHelloInput{
				Version:  "1.0.0",
				KeyFname: keyFname,
				KeyID:    keyID,
				Fatal:    true,

				GOOS:   "linux",
				GOARCH: "amd64",
			})
			t.Log(fname)
		})
	}
}
