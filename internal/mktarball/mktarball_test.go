package mktarball

import (
	"testing"

	"github.com/TecharoHQ/yeet/internal/yeettest"
)

func TestBuild(t *testing.T) {
	yeettest.BuildHello(t, Build)
}
