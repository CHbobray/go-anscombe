# Go for Statistics: Anscombe Quartet Regression

This repository tests whether Go, using the [`github.com/montanaflynn/stats`](https://github.com/montanaflynn/stats) package, can stand in for Python and R for basic statistical work. It runs simple linear regression on the Anscombe Quartet (Anscombe 1973; Miller 2015), checks the Go results against R and Python with Go's `testing` package, and compares execution time and memory across the three languages, including a Monte Carlo benchmark.

**Short answer:** Go reproduces R's `lm()` and Python's statsmodels results to at least nine decimal places, and runs regressions far faster than the R and Python programs from Miller (2015): about 1,000 times faster than the formula-based workflow. It is a sound choice for statistics that run inside backend services. It is not yet a full replacement for R or Python for exploratory data science. The recommendation section below explains why.

## Repository layout

| Path | Role |
|---|---|
| `anscombe/data.go` | The four Anscombe data sets |
| `anscombe/regression.go` | `Fit(x, y)`: slope, intercept, standard errors, t statistic, R², residual standard error. General-purpose: works on any x/y data, not just the quartet |
| `anscombe/synthetic.go` | `Synthetic(...)`: generates data from a known line plus noise, for tests and Monte Carlo runs |
| `anscombe/*_test.go` | Unit tests against R, synthetic-data tests, benchmarks, runnable example |
| `cmd/anscombe/` | The command-line application and its tests |
| `bin/` | Ready-to-run executables: Windows `.exe`, macOS `.app` (zipped), plain macOS and Linux binaries |
| `macos/Anscombe.app/` | Source of the macOS app bundle (launcher script and `Info.plist`) |
| `python/anscombe_regression.py` | Python version (statsmodels, as in Miller's chapter 1 program, plus scipy) with the same Monte Carlo benchmark |
| `r/anscombe_regression.R` | R version (`lm()`, as in Miller's chapter 1 program, plus `.lm.fit()`) with the same Monte Carlo benchmark |
| `data/` | The quartet as CSV: `anscombe.csv` (all sets) and `anscombe_set1.csv` … `anscombe_set4.csv`, usable with `-csv` |
| `results/` | Saved output from the runs reported below: test results, benchmarks, run log, CPU and memory profiles, Python and R output |
| `Makefile` | Shortcuts for testing, benchmarking, profiling, building and comparing |

## The Anscombe Quartet

Anscombe built four small data sets (11 points each) that look completely different when plotted (a noisy line, a curve, a straight line with one outlier, and a vertical stack with one high-leverage point) yet produce nearly identical summary statistics and regression lines. The lesson is to plot data before trusting a model. Here it doubles as a test: any correct regression routine must give about y = 3.00 + 0.500x and R² = 0.67 for all four sets.

## Results: Go matches R and Python

All three languages give the same estimates:

| Set | Intercept | Slope | Intercept SE | Slope SE | R² |
|---|---|---|---|---|---|
| I | 3.000091 | 0.500091 | 1.124747 | 0.117906 | 0.666542 |
| II | 3.000909 | 0.500000 | 1.125302 | 0.117964 | 0.666242 |
| III | 3.002455 | 0.499727 | 1.124481 | 0.117878 | 0.666324 |
| IV | 3.001727 | 0.499909 | 1.123921 | 0.117819 | 0.666707 |

`TestFitMatchesR` compares every statistic (intercept, slope, both standard errors, t statistic, R², residual standard error) with R's `summary(lm(y ~ x))` printed to 12 decimals, and requires agreement within 1e-9. All four sets pass. Python statsmodels output matches R exactly at the six decimals it prints.

### How the Go regression works

`stats.LinearRegression` returns fitted points rather than coefficients, so `Fit` recovers the slope and intercept from two fitted points with different x values. R² comes from `stats.Correlation`, and the standard errors use the textbook OLS formulas with `stats.Mean` and `stats.PopulationVariance`. Bad input returns an error instead of a wrong answer: unequal lengths, fewer than three points, or a constant x (where the slope is undefined).

## Performance

Test machine: 2-core Intel Xeon, Linux; Go 1.24.7; Python 3.13.16 with pandas 3.0.5, scipy 1.18.1, statsmodels 0.15.0; R 4.3.3. Raw output is in `results/`. Timings vary from run to run and machine to machine, so treat the ratios as more meaningful than the exact numbers.

### Anscombe regressions

Each program fits all four sets repeatedly and reports the time per fit. "Formula" means the workflow in Miller's programs (`lm(y ~ x)` with `summary()`, statsmodels `ols("y ~ x").fit()`). "Lean" means the fastest plain-array routine in the same language, so Go is also compared with the best case for Python and R.

| Program | Reps × 4 sets | Time per fit | Memory |
|---|---|---|---|
| **Go** `anscombe.Fit` | 100,000 | **0.4–0.7 µs** | 672 bytes, 6 allocations per fit; 4 MB heap; 13 MB from OS |
| R `.lm.fit()` (lean) | 100,000 | 3.2 µs | |
| R `summary(lm())` (formula) | 10,000 | 627 µs | 72 MB peak R heap (whole run) |
| Python `scipy.stats.linregress` (lean) | 10,000 | 450 µs | |
| Python statsmodels `ols().fit()` (formula) | 1,000 | 4,379 µs | 144 MB peak resident memory (whole run) |

Go's own benchmarks (`go test -bench=. -benchmem ./anscombe`, saved in `results/go_benchmarks.txt`) agree: 530–600 ns per Anscombe fit. On larger synthetic samples, time grows linearly with sample size: 3.8 µs at n = 100, 41 µs at n = 1,000 and 0.49 ms at n = 10,000.

### Monte Carlo benchmark

Each language draws 100,000 random samples of 100 points from the known line y = 3 + 0.5x + N(0, 1.24²), which matches the Anscombe line and noise level. It fits each sample and checks how often the 95% confidence interval for the slope contains the true value. This tests speed on a realistic simulation workload and, at the same time, tests that the standard errors are correct.

| Language | Time per simulate + fit | Total for 100,000 | Mean slope | 95% CI coverage |
|---|---|---|---|---|
| **Go** | **5–8 µs** | **0.5–0.8 s** | 0.50012 | 95.0% |
| R (`.lm.fit()`) | 26 µs | 2.6 s | 0.50010 | 95.0% |
| Python (`scipy.stats.linregress`) | 506 µs | 50.6 s | 0.50000 | 95.0% |

All three languages recover the true slope and give the correct 95% coverage, which confirms that Go's standard errors are right on fresh data, not just on the quartet. Go is about 4× faster than R's lean routine and about 70× faster than Python here.

### Profiles and logs

`anscombe -log run.log -cpuprofile cpu.prof -memprofile mem.prof` writes a timestamped log and Go `pprof` profiles (`make profile` does this and prints the top functions). The saved profile (`results/go_cpu_profile.txt`) shows that about half the CPU time goes to `Fit`, and within that, most goes to the stats package's `LinearRegression`, `Sum` and variance routines. The memory profile shows that `stats.LinearRegression` makes the largest allocations, because it builds a new series of fitted points on every call. That is the first thing to replace if Go ever needed to be faster still.

**What the numbers mean.** Against the workflow data scientists actually use, Go is about 1,000 times faster than R and 6,000 times faster than Python statsmodels, and uses a small fraction of the memory. Most of that gap is overhead from formula parsing, data frames and summary objects, not the arithmetic itself. Against R's lean `.lm.fit()` the gap narrows to about 5–8×. For one regression on 11 points, every option finishes in well under a millisecond, so speed matters only when regressions run millions of times (as in simulation) or inside a latency-sensitive service.

## Recommendation to management

**Use Go for statistics that run in production backend services and for heavy simulation, and let data scientists keep Python or R for exploratory work and modeling for now.**

Reasons Go is a good fit:

- **Correct.** The stats package plus a few lines of code reproduce R and statsmodels to nine or more decimal places, and the Monte Carlo test shows the standard errors hold up on new data. Go's built-in `testing` package makes it easy to lock those results in.
- **Fast and light.** Orders of magnitude faster than typical R and Python code, with tiny memory use. That suits web services that compute statistics per request, and simulation studies that need millions of fits.
- **Easy to deploy.** One self-contained executable per platform, with no interpreter, packages or virtual environments to install.
- **One language.** Models already validated in R or Python can be re-implemented in Go once and served from the same codebase as the rest of the backend.

Concerns data scientists will reasonably raise:

- **Limited library.** `montanaflynn/stats` covers descriptive statistics, correlation and simple regression. It has no multiple regression, standard errors, p-values, GLMs, mixed models or time series. Even this project had to compute standard errors and t critical values by hand. [Gonum](https://www.gonum.org/) adds matrix algebra and more distributions but still falls far short of R or Python's ecosystems.
- **Small community.** As of June 2023 the stats package had 28 contributors and about 15,000 users, compared with the huge user bases, textbooks and Stack Overflow history behind R and Python.
- **No interactive workflow.** No mature notebook culture, no equivalent of ggplot2 or matplotlib, and no data frame as convenient as pandas or R's. That is the exact gap the Anscombe Quartet warns about, since you need to plot data before trusting a regression.
- **Verbosity.** Work that is one line in R (`lm(y ~ x)`) takes a function, error handling and tests in Go.
- **Hiring and training.** Most data scientists know Python or R, not Go.

**Suggested path:** keep Python/R for research and model development; port validated models that need to run in production to Go, using unit tests against the R or Python output (as this repository does) to prove the port is correct; and revisit the decision as Go's statistics ecosystem grows.

## How to run the application

### Windows

Download `bin/anscombe-windows-amd64.exe` and double-click it. A console window shows the results and stays open until you press Enter. To use options, open Command Prompt or PowerShell in the download folder and run, for example:

```powershell
.\anscombe-windows-amd64.exe -csv anscombe_set2.csv -reps 50000
```

If Windows SmartScreen warns about an unrecognized app, click **More info → Run anyway**.

### macOS

Download `bin/Anscombe-macOS.zip`, double-click it to unzip, and double-click **Anscombe.app**. It opens a Terminal window and runs the build that matches your Mac (Apple silicon or Intel).

The app is not signed with an Apple developer certificate, so the first time macOS may say it "cannot be opened". Right-click (or Control-click) **Anscombe.app**, choose **Open**, then click **Open** again. Alternatively, run `xattr -dr com.apple.quarantine Anscombe.app` in Terminal.

To use options on a Mac, run the plain binary in Terminal:

```bash
chmod +x anscombe-darwin-arm64        # Intel Macs: anscombe-darwin-amd64
./anscombe-darwin-arm64 -mc 100000
```

### Linux

```bash
chmod +x anscombe-linux-amd64
./anscombe-linux-amd64
```

### Options

```text
-reps N          timing repetitions for the Anscombe fits (default 10000; 0 skips timing)
-csv FILE        fit your own data: a CSV with x in the first column and y in the second;
                 a header row is allowed
-mc N            Monte Carlo simulations on synthetic data (default 10000; 0 skips)
-n N             observations per Monte Carlo sample (default 100; at least 3)
-seed N          random seed, so a Monte Carlo run can be repeated exactly (default 1)
-log FILE        write a timestamped log of the run
-cpuprofile FILE write a CPU profile (view with: go tool pprof FILE)
-memprofile FILE write a memory profile (view with: go tool pprof FILE)
-h               help
```

The application checks its input and explains problems instead of crashing. For example, it reports a negative or non-numeric `-reps`, `-n` below 3, a stray argument without a dash, an unknown option, a missing CSV file, a CSV line that is not two numbers, a CSV with fewer than three rows, and data where every x is the same.

## Building and testing from source

Requires Go 1.21 or later.

```bash
git clone https://github.com/CHbobray/go-anscombe.git
cd go-anscombe
go test -v -cover ./...                           # unit tests
go test -bench=. -benchmem -run='^$' ./anscombe   # benchmarks
go run ./cmd/anscombe                             # run the app
make build                                        # rebuild every executable into bin/
make app                                          # rebuild bin/Anscombe-macOS.zip
make profile                                      # log + CPU and memory profiles
```

On Windows without `make`, build the executable with:

```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/anscombe-windows-amd64.exe ./cmd/anscombe
```

### Python and R comparison programs

```bash
pip install numpy pandas scipy statsmodels
python3 python/anscombe_regression.py 1000 100000   # args: reps, Monte Carlo runs, [n]
Rscript r/anscombe_regression.R 10000 100000        # uses only base R
```

## Tests

| Test | Checks |
|---|---|
| `TestFitMatchesR` | All seven statistics for all four sets, within 1e-9 of R |
| `TestQuartetLesson` | Every set gives about y = 3.00 + 0.500x and R² = 0.67 |
| `TestFitPerfectLine` | Exact answer for y = 1 + 2x |
| `TestFitRecoversSyntheticModel` | On 10,000 synthetic points, the estimates fall within 4 standard errors of the true line |
| `TestStandardErrorCoverage` | Monte Carlo check: across 2,000 synthetic samples, 95% confidence intervals contain the true slope about 95% of the time |
| `TestFitErrors` | Unequal lengths, too few points, empty input, constant x |
| `TestQuartetReturnsCopies` | The built-in data cannot be changed by callers |
| `ExampleFit` | Runnable documentation example |
| `TestRunQuartet`, `TestRunCSVMatchesBuiltInSet` | App output, and CSV input reproducing Set I |
| `TestRunMonteCarlo` | The app's simulation recovers the true slope with about 95% coverage |
| `TestLogAndProfilesAreWritten` | `-log`, `-cpuprofile` and `-memprofile` produce their files |
| `TestParseFlagsRejectsBadInput`, `TestRunRejectsBadFiles` | Every kind of bad option and bad CSV is rejected with an error |

Coverage: 95.6% of the `anscombe` package and 85.9% of the application. The uncovered part is `main`'s exit handling and the Windows "Press Enter" pause.

## References

Anscombe, F. J. 1973. "Graphs in Statistical Analysis." *The American Statistician* 27 (1): 17–21.

Flynn, Montana. `stats`: A well tested and comprehensive Golang statistics library. https://github.com/montanaflynn/stats

Miller, Thomas W. 2015. *Modeling Techniques in Predictive Analytics with Python and R: A Guide to Data Science.* Upper Saddle River, NJ: Pearson Education.
