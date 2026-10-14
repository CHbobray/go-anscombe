// Command anscombe fits a simple linear regression to each Anscombe Quartet
// data set (or to a CSV file of x,y pairs), prints the results, and times
// repeated fits so Go can be compared with Python and R.
//
// Usage:
//
//	anscombe                  fit the four Anscombe sets, time 10,000 reps
//	anscombe -reps 100000     choose the number of timing repetitions
//	anscombe -csv data.csv    fit your own data (two numeric columns: x,y)
//	anscombe -h               show help
package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/CHbobray/go-anscombe/anscombe"
)

func main() {
	reps := flag.Int("reps", 10000, "number of times to repeat the fits for timing (0 skips timing)")
	csvPath := flag.String("csv", "", "optional CSV file with two numeric columns (x,y); a header row is allowed")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"anscombe: simple linear regression in Go on the Anscombe Quartet or your own data.\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	err := run(os.Stdout, *reps, *csvPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}

	// When double-clicked on Windows the console closes at once; keep it open.
	if runtime.GOOS == "windows" && len(os.Args) == 1 {
		fmt.Print("\nPress Enter to exit...")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	}
	if err != nil {
		os.Exit(1)
	}
}

func run(w io.Writer, reps int, csvPath string) error {
	if reps < 0 {
		return errors.New("-reps must be zero or more")
	}

	sets := anscombe.Quartet()
	if csvPath != "" {
		ds, err := readCSV(csvPath)
		if err != nil {
			return err
		}
		sets = []anscombe.DataSet{ds}
	}

	fmt.Fprintf(w, "%-10s %12s %12s %12s %12s %10s\n", "Set", "Intercept", "Slope", "Int SE", "Slope SE", "R^2")
	for _, ds := range sets {
		r, err := anscombe.Fit(ds.X, ds.Y)
		if err != nil {
			return fmt.Errorf("%s: %w", ds.Name, err)
		}
		fmt.Fprintf(w, "%-10s %12.6f %12.6f %12.6f %12.6f %10.6f\n",
			ds.Name, r.Intercept, r.Slope, r.InterceptSE, r.SlopeSE, r.RSquared)
	}

	if reps == 0 {
		return nil
	}

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

	fits := reps * len(sets)
	fmt.Fprintf(w, "\nGo: %d reps x %d sets in %.3f s (%.2f microseconds per fit)\n",
		reps, len(sets), elapsed.Seconds(), float64(elapsed.Microseconds())/float64(fits))
	fmt.Fprintf(w, "Memory: %.0f bytes and %.1f allocations per fit; heap in use %.2f MB; %.1f MB obtained from OS\n",
		float64(after.TotalAlloc-before.TotalAlloc)/float64(fits),
		float64(after.Mallocs-before.Mallocs)/float64(fits),
		float64(after.HeapInuse)/1e6,
		float64(after.Sys)/1e6)
	return nil
}

// readCSV loads two numeric columns from a CSV file. A first row that does
// not parse as numbers is treated as a header and skipped.
func readCSV(path string) (anscombe.DataSet, error) {
	f, err := os.Open(path)
	if err != nil {
		return anscombe.DataSet{}, err
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
	return ds, nil
}
