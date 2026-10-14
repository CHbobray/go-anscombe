package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var quiet = log.New(io.Discard, "", 0)

func TestRunQuartet(t *testing.T) {
	var out bytes.Buffer
	if err := run(&out, quiet, options{reps: 10, mcSize: 100}); err != nil {
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
	o := options{csvPath: filepath.Join("..", "..", "data", "anscombe_set1.csv"), mcSize: 100}
	if err := run(&out, quiet, o); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "CSV data       3.000091     0.500091") {
		t.Errorf("CSV fit does not match Set I:\n%s", out.String())
	}
}

func TestRunMonteCarlo(t *testing.T) {
	var out bytes.Buffer
	if err := run(&out, quiet, options{mcRuns: 2000, mcSize: 50, seed: 3}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "Monte Carlo: 2000 samples of n = 50") {
		t.Fatalf("output missing Monte Carlo header:\n%s", got)
	}

	// The average slope should be close to the true 0.5, and about 95% of
	// the confidence intervals should contain it.
	var meanSlope, sd, coverage float64
	line := got[strings.Index(got, "mean slope"):]
	if _, err := fmt.Sscanf(line, "mean slope %f (true 0.5), SD %f, 95%% CI coverage %f%%", &meanSlope, &sd, &coverage); err != nil {
		t.Fatalf("cannot read results line %q: %v", line, err)
	}
	if math.Abs(meanSlope-0.5) > 0.005 {
		t.Errorf("mean slope = %.5f, want about 0.5", meanSlope)
	}
	if coverage < 93 || coverage > 97 {
		t.Errorf("coverage = %.1f%%, want about 95%%", coverage)
	}
}

func TestLogAndProfilesAreWritten(t *testing.T) {
	dir := t.TempDir()
	o := options{
		reps:       100,
		mcSize:     100,
		logPath:    filepath.Join(dir, "run.log"),
		cpuProfile: filepath.Join(dir, "cpu.prof"),
		memProfile: filepath.Join(dir, "mem.prof"),
	}
	if err := runWithExtras(io.Discard, o); err != nil {
		t.Fatal(err)
	}
	logText, err := os.ReadFile(o.logPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"start:", "fit Set I", "timing:", "finished successfully"} {
		if !strings.Contains(string(logText), want) {
			t.Errorf("log missing %q:\n%s", want, logText)
		}
	}
	for _, p := range []string{o.cpuProfile, o.memProfile} {
		if info, err := os.Stat(p); err != nil || info.Size() == 0 {
			t.Errorf("profile %s missing or empty", p)
		}
	}
}

func TestParseFlagsRejectsBadInput(t *testing.T) {
	cases := map[string][]string{
		"negative reps":    {"-reps", "-1"},
		"non-number reps":  {"-reps", "lots"},
		"negative mc":      {"-mc", "-5"},
		"sample too small": {"-n", "2"},
		"stray argument":   {"5000"},
		"unknown flag":     {"-speed", "fast"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseFlags(args); err == nil {
				t.Errorf("parseFlags(%q) accepted bad input", args)
			}
		})
	}
	if o, err := parseFlags([]string{"-reps", "50", "-mc", "10", "-n", "20"}); err != nil || o.reps != 50 || o.mcRuns != 10 || o.mcSize != 20 {
		t.Errorf("valid flags: got %+v, %v", o, err)
	}
}

func TestRunRejectsBadFiles(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	cases := map[string]string{
		"missing file": filepath.Join(dir, "nope.csv"),
		"non-numeric":  write("bad.csv", "x,y\n1,2\nthree,4\n"),
		"one column":   write("narrow.csv", "1\n2\n3\n"),
		"too few rows": write("short.csv", "x,y\n1,2\n2,3\n"),
		"constant x":   write("flat.csv", "x,y\n4,1\n4,2\n4,3\n"),
	}
	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			if err := run(io.Discard, quiet, options{csvPath: path, mcSize: 100}); err == nil {
				t.Error("expected an error")
			}
		})
	}
}
