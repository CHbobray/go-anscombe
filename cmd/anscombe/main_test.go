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
	for _, want := range []string{
		"Set I: y = 3.0001 + 0.5001x", // R-style summary block
		"x               0.5001      0.1179     4.241    0.00217 **",
		"Multiple R-squared: 0.6665,  Adjusted R-squared: 0.6295",
		"F-statistic: 17.99 on 1 and 9 DF,  p-value: 0.00217",
		"[  0.2334,   0.7668]", // 95% CI, as in statsmodels' summary
		"Side by side", "Set IV", "0.666542",
		"Go: 10 reps x 4 sets", "Memory:",
	} {
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
	if !strings.Contains(out.String(), "CSV data      9.000    7.501     3.000091     0.500091") {
		t.Errorf("CSV fit does not match Set I:\n%s", out.String())
	}
}

func TestRunDuetWithPlot(t *testing.T) {
	plot := filepath.Join(t.TempDir(), "duet.svg")
	var out bytes.Buffer
	if err := run(&out, quiet, options{dataName: "duet", svgPath: plot, mcSize: 100}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Duet A: y = 2.5177", "Duet B: y = 2.5215", "Plot written to"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q:\n%s", want, out.String())
		}
	}
	svg, err := os.ReadFile(plot)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(svg), "<svg") || strings.Count(string(svg), "<circle") != 24 {
		t.Errorf("plot is not an SVG with 24 points")
	}
}

func TestStars(t *testing.T) {
	cases := map[float64]string{0.0001: "***", 0.005: "**", 0.03: "*", 0.07: ".", 0.5: ""}
	for p, want := range cases {
		if got := stars(p); got != want {
			t.Errorf("stars(%g) = %q, want %q", p, got, want)
		}
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
		"unknown data set": {"-data", "trio"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseFlags(args); err == nil {
				t.Errorf("parseFlags(%q) accepted bad input", args)
			}
		})
	}
	o, err := parseFlags([]string{"-reps", "50", "-mc", "10", "-n", "20", "-data", "duet"})
	if err != nil || o.reps != 50 || o.mcRuns != 10 || o.mcSize != 20 || o.dataName != "duet" {
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
