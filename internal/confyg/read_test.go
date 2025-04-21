// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confyg

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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

	if !bytes.Equal(ndata, golden) {
		t.Errorf("formatted %s incorrectly: diff shows -golden, +ours", base)
		tdiff(t, string(golden), string(ndata))
		return
	}
}

// diff returns the output of running diff on b1 and b2.
func diff(b1, b2 []byte) (data []byte, err error) {
	f1, err := os.CreateTemp("", "testdiff")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := os.CreateTemp("", "testdiff")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	f1.Write(b1)
	f2.Write(b2)

	data, err = exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}
	return
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
