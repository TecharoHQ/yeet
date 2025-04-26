package mktarball

import (
	"testing"

	"github.com/TecharoHQ/yeet/internal/yeettest"
)

func TestBuild(t *testing.T) {
	yeettest.BuildHello(t, Build, "1.0.0", true)
}

func TestBuildError(t *testing.T) {
	yeettest.BuildHello(t, Build, ".0.0", false)
}
