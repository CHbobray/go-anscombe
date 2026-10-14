# Go for Statistics: Anscombe Quartet Regression

This repository tests whether Go, using the [`github.com/montanaflynn/stats`](https://github.com/montanaflynn/stats) package, can stand in for Python and R for basic statistical work. It runs simple linear regression on the Anscombe Quartet (Anscombe 1973; Miller 2015), checks the Go results against R and Python with Go's `testing` package, and compares execution time and memory across the three languages.

**Short answer:** Go reproduces R's `lm()` and Python's statsmodels results to at least nine decimal places, and runs the regression roughly 1,000 times faster than the R and Python programs from Miller (2015). It is a sound choice for statistics that run inside backend services. It is not yet a full replacement for R or Python for exploratory data science. The recommendation section below explains why.

## Repository layout

| Path | Contents |
|---|---|
| `anscombe/data.go` | The four Anscombe data sets |
| `anscombe/regression.go` | `Fit(x, y)`: slope, intercept, standard errors, t statistic, R², residual standard error |
| `anscombe/regression_test.go` | Unit tests against R reference values, edge-case tests, benchmarks, runnable example |
| `cmd/anscombe/` | The command-line application and its tests |
| `bin/` | Ready-to-run executables for Windows, macOS and Linux |
| `python/anscombe_regression.py` | Python version (statsmodels, as in Miller's chapter 1 program, plus scipy) |
| `r/anscombe_regression.R` | R version (`lm()`, as in Miller's chapter 1 program, plus `.lm.fit()`) |
| `data/` | The quartet as CSV: `anscombe.csv` (all sets) and `anscombe_set1.csv` … `anscombe_set4.csv` |
| `Makefile` | Shortcuts for testing, benchmarking, building and comparing |

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

Each program fits all four sets repeatedly and reports the time per fit. "Formula" means the usual data-scientist workflow from Miller's programs (`lm(y ~ x)` with `summary()`, statsmodels `ols("y ~ x").fit()`). "Lean" means the fastest plain-array routine in the same language, so Go is also compared with the best case for Python and R.

| Program | Reps × 4 sets | Time per fit | Memory |
|---|---|---|---|
| **Go** `anscombe.Fit` | 100,000 | **0.43–0.58 µs** | 672 bytes, 6 allocations per fit; 2.7 MB heap; 11 MB from OS |
| R `.lm.fit()` (lean) | 100,000 | 2.8 µs | — |
| R `summary(lm())` (formula) | 10,000 | 595 µs | 49.6 MB peak R heap |
| Python `scipy.stats.linregress` (lean) | 10,000 | 470 µs | — |
| Python statsmodels `ols().fit()` (formula) | 1,000 | 4,615 µs | 142 MB peak resident memory |

Go's own benchmark (`go test -bench=. -benchmem ./anscombe`) agrees: 479–583 ns per fit and 2.3 µs for all four sets.

Test machine: 2-core Intel Xeon at 2.1 GHz, Linux; Go 1.24.7; Python 3.13.16 with pandas 3.0.5, scipy 1.18.1, statsmodels 0.15.0; R 4.3.3. Timings vary from run to run and machine to machine, so treat the ratios as more meaningful than the exact numbers.

**What the numbers mean.** Against the workflow data scientists actually use, Go is about 1,000 times faster than R and 10,000 times faster than Python statsmodels, and uses a small fraction of the memory. Most of that gap is overhead from formula parsing, data frames and summary objects, not the arithmetic itself. Against R's lean `.lm.fit()` the gap narrows to about 5–6×. For one regression on 11 points, every option finishes in well under a millisecond, so speed matters only when regressions run millions of times or inside a latency-sensitive service.

## Recommendation to management

**Use Go for statistics that run in production backend services, and let data scientists keep Python or R for exploratory work and modeling for now.**

Reasons Go is a good fit:

- **Correct.** The stats package plus a few lines of code reproduce R and statsmodels to nine or more decimal places, and Go's built-in `testing` package makes it easy to lock those results in.
- **Fast and light.** Orders of magnitude faster than typical R and Python code, with tiny memory use. That suits web services that compute statistics per request.
- **Easy to deploy.** One self-contained executable per platform, with no interpreter, packages or virtual environments to install.
- **One language.** Models already validated in R or Python can be re-implemented in Go once and served from the same codebase as the rest of the backend.

Concerns data scientists will reasonably raise:

- **Limited library.** `montanaflynn/stats` covers descriptive statistics, correlation and simple regression. It has no multiple regression, standard errors, p-values, GLMs, mixed models or time series. Even this project had to compute standard errors by hand. [Gonum](https://www.gonum.org/) adds matrix algebra and more distributions but still falls far short of R or Python's ecosystems.
- **Small community.** As of June 2023 the stats package had 28 contributors and about 15,000 users, compared with the huge user bases, textbooks and Stack Overflow history behind R and Python.
- **No interactive workflow.** No mature notebook culture, no equivalent of ggplot2 or matplotlib, and no data frame as convenient as pandas or R's. That is the exact gap the Anscombe Quartet warns about, since you need to plot data before trusting a regression.
- **Verbosity.** Work that is one line in R (`lm(y ~ x)`) takes a function, error handling and tests in Go.
- **Hiring and training.** Most data scientists know Python or R, not Go.

**Suggested path:** keep Python/R for research and model development; port validated models that need to run in production to Go, using unit tests against the R or Python output (as this repository does) to prove the port is correct; and revisit the decision as Go's statistics ecosystem grows.

## How to run

### Ready-made executables (no Go needed)

Download the file for your computer from `bin/`:

- Windows: `anscombe-windows-amd64.exe`. Double-click it, or run it from Command Prompt or PowerShell.
- macOS, Apple silicon (M1 and later): `anscombe-darwin-arm64`
- macOS, Intel: `anscombe-darwin-amd64`
- Linux: `anscombe-linux-amd64`

On macOS and Linux, open Terminal in the download folder and run:

```bash
chmod +x anscombe-darwin-arm64
./anscombe-darwin-arm64
```

If macOS says the app cannot be opened because the developer cannot be verified, run `xattr -d com.apple.quarantine anscombe-darwin-arm64` once, or right-click the file in Finder and choose Open.

Options:

```text
-reps N       number of timing repetitions (default 10000; 0 skips timing)
-csv FILE     fit your own data: a CSV with x in the first column and y in the second;
              a header row is allowed
-h            help
```

Example: `anscombe -csv data/anscombe_set2.csv -reps 50000`

### From source (Go 1.21 or later)

```bash
git clone <repository URL>
cd go-anscombe
go test -v -cover ./...                         # unit tests
go test -bench=. -benchmem -run='^$' ./anscombe   # benchmarks
go run ./cmd/anscombe                            # run the app
make build                                       # rebuild every executable into bin/
```

On Windows without `make`, build one executable with:

```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o bin/anscombe-windows-amd64.exe ./cmd/anscombe
```

### Python and R comparison programs

```bash
pip install pandas scipy statsmodels
python3 python/anscombe_regression.py 1000      # argument is the number of reps
Rscript r/anscombe_regression.R 10000           # uses only base R
```

## Tests

| Test | Checks |
|---|---|
| `TestFitMatchesR` | All seven statistics for all four sets, within 1e-9 of R |
| `TestQuartetLesson` | Every set gives about y = 3.00 + 0.500x and R² = 0.67 |
| `TestFitPerfectLine` | Exact answer for y = 1 + 2x |
| `TestFitErrors` | Unequal lengths, too few points, empty input, constant x |
| `TestQuartetReturnsCopies` | The built-in data cannot be changed by callers |
| `ExampleFit` | Runnable documentation example |
| `cmd/anscombe` tests | App output, CSV input matching Set I, and rejection of bad arguments and files |

Coverage: 94.9% of the `anscombe` package, 71.9% of the application (the uncovered part is `main`'s flag handling and exit).

## References

Anscombe, F. J. 1973. "Graphs in Statistical Analysis." *The American Statistician* 27 (1): 17–21.

Miller, Thomas W. 2015. *Modeling Techniques in Predictive Analytics with Python and R: A Guide to Data Science.* Upper Saddle River, NJ: Pearson Education.

Flynn, Montana. `stats`: A well tested and comprehensive Golang statistics library. https://github.com/montanaflynn/stats
