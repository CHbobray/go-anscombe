// Command anscombe fits a simple linear regression to each Anscombe Quartet
// data set (or to a CSV file of x,y pairs), prints the results, and times
// repeated fits so Go can be compared with Python and R. It can also run a
// Monte Carlo benchmark on synthetic data, write a log, and save profiles.
//
// Usage:
//
//	anscombe                       fit the four Anscombe sets, time 10,000 reps,
//	                               and run 10,000 Monte Carlo simulations
//	anscombe -reps 100000          choose the number of timing repetitions
//	anscombe -csv data.csv         fit your own data (two numeric columns: x,y)
//	anscombe -mc 10000 -n 100      Monte Carlo benchmark on synthetic data
//	anscombe -log run.log          also write a log of the run
//	anscombe -cpuprofile cpu.prof  save CPU and/or memory profiles (-memprofile)
//	anscombe -h                    show help
package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/CHbobray/go-anscombe/anscombe"
)

// options holds the validated command-line settings.
type options struct {
	reps       int
	csvPath    string
	mcRuns     int
	mcSize     int
	seed       int64
	logPath    string
	cpuProfile string
	memProfile string
}

// True model used by the Monte Carlo benchmark. It mirrors the Anscombe
// line (y = 3 + 0.5x) with noise similar to the quartet's residual SE.
const (
	mcIntercept = 3.0
	mcSlope     = 0.5
	mcSigma     = 1.24
)

func main() {
	opts, err := parseFlags(os.Args[1:])
	if err == nil {
		err = runWithExtras(os.Stdout, opts)
	}
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, "error:", err)
		fmt.Fprintln(os.Stderr, "Run with -h to see the available options.")
	}

	// When double-clicked on Windows the console closes at once; keep it open.
	if runtime.GOOS == "windows" && len(os.Args) == 1 {
		fmt.Print("\nPress Enter to exit...")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	}
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		os.Exit(1)
	}
}

// parseFlags reads and validates the command line, returning a clear
// message for any value the program cannot use.
func parseFlags(args []string) (options, error) {
	var o options
	fs := flag.NewFlagSet("anscombe", flag.ContinueOnError)
	fs.IntVar(&o.reps, "reps", 10000, "number of times to repeat the fits for timing (0 skips timing)")
	fs.StringVar(&o.csvPath, "csv", "", "optional CSV file with two numeric columns (x,y); a header row is allowed")
	fs.IntVar(&o.mcRuns, "mc", 10000, "number of Monte Carlo simulations on synthetic data (0 skips the simulation)")
	fs.IntVar(&o.mcSize, "n", 100, "observations per Monte Carlo sample (at least 3)")
	fs.Int64Var(&o.seed, "seed", 1, "random seed for the Monte Carlo simulation, so results can be repeated")
	fs.StringVar(&o.logPath, "log", "", "write a timestamped log of the run to this file")
	fs.StringVar(&o.cpuProfile, "cpuprofile", "", "write a CPU profile to this file (view with: go tool pprof)")
	fs.StringVar(&o.memProfile, "memprofile", "", "write a memory profile to this file (view with: go tool pprof)")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "anscombe: simple linear regression in Go on the Anscombe Quartet or your own data.\n\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return o, err
	}
	switch {
	case fs.NArg() > 0:
		return o, fmt.Errorf("unexpected argument %q (options start with a dash, e.g. -reps 5000)", fs.Arg(0))
	case o.reps < 0:
		return o, errors.New("-reps must be zero or more")
	case o.mcRuns < 0:
		return o, errors.New("-mc must be zero or more")
	case o.mcSize < 3:
		return o, errors.New("-n must be at least 3 (a regression line needs three points to estimate its error)")
	}
	return o, nil
}

// runWithExtras sets up the optional log and profiles around run.
func runWithExtras(stdout io.Writer, o options) error {
	logger := log.New(io.Discard, "", 0)
	if o.logPath != "" {
		f, err := os.Create(o.logPath)
		if err != nil {
			return fmt.Errorf("creating log file: %w", err)
		}
		defer f.Close()
		logger = log.New(f, "", log.LstdFlags|log.Lmicroseconds)
		logger.Printf("start: go %s %s/%s, %d CPUs, options %+v",
			runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), o)
	}

	if o.cpuProfile != "" {
		f, err := os.Create(o.cpuProfile)
		if err != nil {
			return fmt.Errorf("creating CPU profile: %w", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return fmt.Errorf("starting CPU profile: %w", err)
		}
		defer pprof.StopCPUProfile()
		logger.Printf("CPU profile: %s", o.cpuProfile)
	}

	err := run(stdout, logger, o)

	if o.memProfile != "" && err == nil {
		err = writeMemProfile(o.memProfile)
		logger.Printf("memory profile: %s", o.memProfile)
	}
	if err != nil {
		logger.Printf("finished with error: %v", err)
	} else {
		logger.Printf("finished successfully")
	}
	return err
}

func writeMemProfile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating memory profile: %w", err)
	}
	defer f.Close()
	runtime.GC() // get up-to-date allocation statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("writing memory profile: %w", err)
	}
	return nil
}

// run does the program's work: fit, time, and optionally simulate.
func run(w io.Writer, logger *log.Logger, o options) error {
	sets := anscombe.Quartet()
	if o.csvPath != "" {
		ds, err := readCSV(o.csvPath)
		if err != nil {
			return err
		}
		sets = []anscombe.DataSet{ds}
		logger.Printf("loaded %d rows from %s", len(ds.X), o.csvPath)
	}

	if err := printFits(w, logger, sets); err != nil {
		return err
	}
	if o.reps > 0 {
		if err := timeFits(w, logger, sets, o.reps); err != nil {
			return err
		}
	}
	if o.mcRuns > 0 {
		return monteCarlo(w, logger, o)
	}
	return nil
}

func printFits(w io.Writer, logger *log.Logger, sets []anscombe.DataSet) error {
	fmt.Fprintf(w, "%-10s %12s %12s %12s %12s %10s\n", "Set", "Intercept", "Slope", "Int SE", "Slope SE", "R^2")
	for _, ds := range sets {
		r, err := anscombe.Fit(ds.X, ds.Y)
		if err != nil {
			return fmt.Errorf("%s: %w", ds.Name, err)
		}
		fmt.Fprintf(w, "%-10s %12.6f %12.6f %12.6f %12.6f %10.6f\n",
			ds.Name, r.Intercept, r.Slope, r.InterceptSE, r.SlopeSE, r.RSquared)
		logger.Printf("fit %s: intercept %.6f, slope %.6f, R^2 %.6f", ds.Name, r.Intercept, r.Slope, r.RSquared)
	}
	return nil
}

// timeFits fits every set reps times and reports time and memory per fit.
func timeFits(w io.Writer, logger *log.Logger, sets []anscombe.DataSet, reps int) error {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	start := time.Now()
	for i := 0; i < reps; i++ {
		for _, ds := range sets {
			if _, err := anscombe.Fit(ds.X, ds.Y); err != nil {
				return err
			}
		}
	}
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)

	fits := float64(reps * len(sets))
	perFit := elapsed.Seconds() / fits * 1e6
	fmt.Fprintf(w, "\nGo: %d reps x %d sets in %.3f s (%.2f microseconds per fit)\n",
		reps, len(sets), elapsed.Seconds(), perFit)
	fmt.Fprintf(w, "Memory: %.0f bytes and %.1f allocations per fit; heap in use %.2f MB; %.1f MB obtained from OS\n",
		float64(after.TotalAlloc-before.TotalAlloc)/fits,
		float64(after.Mallocs-before.Mallocs)/fits,
		float64(after.HeapInuse)/1e6,
		float64(after.Sys)/1e6)
	logger.Printf("timing: %d fits in %v (%.2f us/fit)", int(fits), elapsed, perFit)
	return nil
}

// monteCarlo draws o.mcRuns synthetic samples from a known line, fits each,
// and reports how well the estimates recover the truth and how long it took.
func monteCarlo(w io.Writer, logger *log.Logger, o options) error {
	rng := rand.New(rand.NewSource(o.seed))
	tCrit := tCritical95(o.mcSize - 2)

	var sumSlope, sumSqSlope float64
	covered := 0
	start := time.Now()
	for i := 0; i < o.mcRuns; i++ {
		ds := anscombe.Synthetic(rng, o.mcSize, mcIntercept, mcSlope, mcSigma)
		r, err := anscombe.Fit(ds.X, ds.Y)
		if err != nil {
			return fmt.Errorf("Monte Carlo run %d: %w", i+1, err)
		}
		sumSlope += r.Slope
		sumSqSlope += r.Slope * r.Slope
		if math.Abs(r.Slope-mcSlope) <= tCrit*r.SlopeSE {
			covered++
		}
	}
	elapsed := time.Since(start)

	runs := float64(o.mcRuns)
	meanSlope := sumSlope / runs
	sdSlope := math.Sqrt(math.Max(sumSqSlope/runs-meanSlope*meanSlope, 0))
	perRun := elapsed.Seconds() / runs * 1e6

	fmt.Fprintf(w, "\nMonte Carlo: %d samples of n = %d from y = %.1f + %.1fx + N(0, %.2f^2), seed %d\n",
		o.mcRuns, o.mcSize, mcIntercept, mcSlope, mcSigma, o.seed)
	fmt.Fprintf(w, "  mean slope %.5f (true %.1f), SD %.5f, 95%% CI coverage %.1f%%\n",
		meanSlope, mcSlope, sdSlope, 100*float64(covered)/runs)
	fmt.Fprintf(w, "  Go: %.3f s total (%.2f microseconds per simulate + fit)\n", elapsed.Seconds(), perRun)
	logger.Printf("monte carlo: %d runs, n=%d, mean slope %.5f, coverage %.3f, %.2f us/run",
		o.mcRuns, o.mcSize, meanSlope, float64(covered)/runs, perRun)
	return nil
}

// tCritical95 returns the two-sided 95% critical value of Student's t.
// It uses exact table values for small df and the normal approximation
// with a Cornish-Fisher correction above that (error under 0.003).
func tCritical95(df int) float64 {
	table := []float64{0, 12.706, 4.303, 3.182, 2.776, 2.571, 2.447, 2.365, 2.306, 2.262, 2.228}
	if df < len(table) {
		return table[df]
	}
	const z = 1.959964
	d := float64(df)
	return z + (z*z*z+z)/(4*d) + (5*math.Pow(z, 5)+16*z*z*z+3*z)/(96*d*d)
}

// readCSV loads two numeric columns from a CSV file. A first row that does
// not parse as numbers is treated as a header and skipped.
func readCSV(path string) (anscombe.DataSet, error) {
	f, err := os.Open(path)
	if err != nil {
		return anscombe.DataSet{}, fmt.Errorf("cannot open CSV file: %w", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return anscombe.DataSet{}, fmt.Errorf("reading %s: %w", path, err)
	}

	ds := anscombe.DataSet{Name: "CSV data"}
	for i, row := range rows {
		if len(row) < 2 {
			return ds, fmt.Errorf("%s line %d: need two columns, found %d", path, i+1, len(row))
		}
		x, errX := strconv.ParseFloat(strings.TrimSpace(row[0]), 64)
		y, errY := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if errX != nil || errY != nil {
			if i == 0 {
				continue // header row
			}
			return ds, fmt.Errorf("%s line %d: %q, %q are not both numbers", path, i+1, row[0], row[1])
		}
		ds.X = append(ds.X, x)
		ds.Y = append(ds.Y, y)
	}
	if len(ds.X) < 3 {
		return ds, fmt.Errorf("%s: found %d data rows, need at least 3", path, len(ds.X))
	}
	return ds, nil
}
