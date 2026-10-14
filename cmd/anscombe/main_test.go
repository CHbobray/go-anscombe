package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunQuartet(t *testing.T) {
	var out bytes.Buffer
	if err := run(&out, 10, ""); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"Set I ", "Set IV", "3.000091", "0.500091", "0.666542", "Go: 10 reps x 4 sets", "Memory:"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
}

func TestRunCSVMatchesBuiltInSet(t *testing.T) {
	var out bytes.Buffer
	if err := run(&out, 0, filepath.Join("..", "..", "data", "anscombe_set1.csv")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "CSV data       3.000091     0.500091") {
		t.Errorf("CSV fit does not match Set I:\n%s", out.String())
	}
}

func TestRunRejectsBadInput(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.csv")
	if err := os.WriteFile(bad, []byte("x,y\n1,2\nthree,4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := map[string]struct {
		reps int
		path string
	}{
		"negative reps": {-1, ""},
		"missing file":  {0, filepath.Join(dir, "nope.csv")},
		"non-numeric":   {0, bad},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if err := run(&bytes.Buffer{}, c.reps, c.path); err == nil {
				t.Error("expected an error")
			}
		})
	}
}
