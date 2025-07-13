package gpgtest

import (
	"path/filepath"
	"testing"
)

func TestMakeKey(t *testing.T) {
	fname := filepath.Join(t.TempDir(), "key.gpg")

	fp, err := MakeKey(t.Context(), fname)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(fp)
}
