"""Anscombe Quartet regressions in Python, adapted from Miller (2015), chapter 1.

Prints coefficients for each data set using statsmodels OLS (as in Miller's
program), then times repeated fits (statsmodels and the leaner
scipy.stats.linregress) so the result can be compared with the Go
and R programs. Plotting from the original program is omitted.

Usage:  python3 python/anscombe_regression.py [reps]
"""

import sys
import time

import pandas as pd
from scipy import stats
import statsmodels.formula.api as smf

SHARED_X = [10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5]

SETS = {
    "Set I": pd.DataFrame({
        "x": SHARED_X,
        "y": [8.04, 6.95, 7.58, 8.81, 8.33, 9.96, 7.24, 4.26, 10.84, 4.82, 5.68],
    }),
    "Set II": pd.DataFrame({
        "x": SHARED_X,
        "y": [9.14, 8.14, 8.74, 8.77, 9.26, 8.10, 6.13, 3.10, 9.13, 7.26, 4.74],
    }),
    "Set III": pd.DataFrame({
        "x": SHARED_X,
        "y": [7.46, 6.77, 12.74, 7.11, 7.81, 8.84, 6.08, 5.39, 8.15, 6.42, 5.73],
    }),
    "Set IV": pd.DataFrame({
        "x": [8, 8, 8, 8, 8, 8, 8, 19, 8, 8, 8],
        "y": [6.58, 5.76, 7.71, 8.84, 8.47, 7.04, 5.25, 12.50, 5.56, 7.91, 6.89],
    }),
}


def main() -> None:
    reps = int(sys.argv[1]) if len(sys.argv) > 1 else 1000

    print(f"{'Set':<8} {'Intercept':>12} {'Slope':>12} {'Int SE':>12} "
          f"{'Slope SE':>12} {'R^2':>10}")
    for name, df in SETS.items():
        fit = smf.ols("y ~ x", data=df).fit()
        print(f"{name:<8} {fit.params['Intercept']:12.6f} {fit.params['x']:12.6f} "
              f"{fit.bse['Intercept']:12.6f} {fit.bse['x']:12.6f} {fit.rsquared:10.6f}")

    start = time.perf_counter()
    for _ in range(reps):
        for df in SETS.values():
            smf.ols("y ~ x", data=df).fit()
    elapsed = time.perf_counter() - start

    print(f"\nPython statsmodels: {reps} reps x 4 sets in {elapsed:.3f} s "
          f"({elapsed / (reps * 4) * 1e6:.2f} microseconds per fit)")

    # Lean alternative: scipy.stats.linregress on plain arrays (no formula
    # parsing or data frames), closer to what the Go program does.
    arrays = [(df["x"].to_numpy(float), df["y"].to_numpy(float)) for df in SETS.values()]
    lean_reps = reps * 10
    start = time.perf_counter()
    for _ in range(lean_reps):
        for x, y in arrays:
            stats.linregress(x, y)
    elapsed = time.perf_counter() - start
    print(f"Python scipy linregress: {lean_reps} reps x 4 sets in {elapsed:.3f} s "
          f"({elapsed / (lean_reps * 4) * 1e6:.2f} microseconds per fit)")

    try:  # resource is Unix-only; skip the memory line on Windows.
        import resource
        peak = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
        scale = 1e6 if sys.platform == "darwin" else 1e3  # bytes vs KB
        print(f"Memory: peak resident set size {peak / scale:.1f} MB")
    except ImportError:
        pass


if __name__ == "__main__":
    main()
