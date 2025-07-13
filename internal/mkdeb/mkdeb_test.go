package mkdeb

import (
	"path/filepath"
	"testing"

	"github.com/TecharoHQ/yeet/internal/gpgtest"
	"github.com/TecharoHQ/yeet/internal/yeettest"
	"pault.ag/go/debian/deb"
)

func TestBuild(t *testing.T) {
	keyFname := filepath.Join(t.TempDir(), "foo.gpg")
	keyID, err := gpgtest.MakeKey(t.Context(), keyFname)
	if err != nil {
		t.Fatal(err)
	}

	fname := yeettest.BuildHello(t, Build, "1.0.0", keyFname, keyID, true)

	debFile, close, err := deb.LoadFile(fname)
	if err != nil {
		t.Fatalf("failed to load deb file: %v", err)
	}
	defer close()

	if debFile.Control.Version.Empty() {
		t.Error("version is empty")
	}
}

func TestBuildError(t *testing.T) {
	yeettest.BuildHello(t, Build, ".0.0", "", "", false)
}
