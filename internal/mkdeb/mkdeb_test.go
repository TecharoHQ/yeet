package mkdeb

import (
	"testing"

	"github.com/TecharoHQ/yeet/internal/yeettest"
	"pault.ag/go/debian/deb"
)

func TestBuild(t *testing.T) {
	fname := yeettest.BuildHello(t, Build, "1.0.0", true)

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
	yeettest.BuildHello(t, Build, ".0.0", false)
}
