// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confyg

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

// Test that reading and then writing the golden files
// does not change their output.
func TestPrintGolden(t *testing.T) {
	outs, err := filepath.Glob("testdata/*.golden")
	if err != nil {
		t.Fatal(err)
	}
	for _, out := range outs {
		testPrint(t, out, out)
	}
}

// testPrint is a helper for testing the printer.
// It reads the file named in, reformats it, and compares
// the result to the file named out.
func testPrint(t *testing.T, in, out string) {
	data, err := os.ReadFile(in)
	if err != nil {
		t.Error(err)
		return
	}

	golden, err := os.ReadFile(out)
	if err != nil {
		t.Error(err)
		return
	}

	base := "testdata/" + filepath.Base(in)
	f, err := parse(in, data)
	if err != nil {
		t.Error(err)
		return
	}

	ndata := Format(f)

	normalizedGolden := normalizeNewlines(golden)
	normalizedOutput := normalizeNewlines(ndata)

	if !bytes.Equal(normalizedOutput, normalizedGolden) {
		t.Errorf("formatted %s incorrectly: diff shows -golden, +ours", base)
		tdiff(t, string(normalizedGolden), string(normalizedOutput))
		return
	}
}

// diff returns the output of running diff on b1 and b2.
func diff(b1, b2 []byte) (data []byte, err error) {
	ud := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(b1)),
		B:        difflib.SplitLines(string(b2)),
		FromFile: "golden",
		ToFile:   "ours",
		Context:  3,
	}

	text, err := difflib.GetUnifiedDiffString(ud)
	if err != nil {
		return nil, err
	}

	return []byte(text), nil
}

// tdiff logs the diff output to t.Error.
func tdiff(t *testing.T, a, b string) {
	data, err := diff([]byte(a), []byte(b))
	if err != nil {
		t.Error(err)
		return
	}
	t.Error(string(data))
}

func normalizeNewlines(in []byte) []byte {
	if len(in) == 0 {
		return in
	}

	out := bytes.ReplaceAll(in, []byte("\r\n"), []byte("\n"))
	return bytes.ReplaceAll(out, []byte("\r"), []byte("\n"))
}
