package mkportable

import (
	"path/filepath"
	"testing"

	"github.com/TecharoHQ/yeet/internal/gpgtest"
	"github.com/TecharoHQ/yeet/internal/yeettest"
)

func TestBuild(t *testing.T) {
	keyFname := filepath.Join(t.TempDir(), "foo.gpg")
	keyID, err := gpgtest.MakeKey(t.Context(), keyFname)
	if err != nil {
		t.Fatal(err)
	}

	fname := yeettest.BuildHello(t, build, yeettest.BuildHelloInput{
		Version:  "1.0.0",
		KeyFname: keyFname,
		KeyID:    keyID,
		Fatal:    true,
	})
	t.Log(fname)
}

func TestBuildError(t *testing.T) {
	yeettest.BuildHello(t, build, yeettest.BuildHelloInput{
		Version: ".0.0",
		Fatal:   false,
	})
}
