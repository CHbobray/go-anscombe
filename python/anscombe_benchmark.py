"""Timing and Monte Carlo benchmark for the Anscombe Quartet in Python.

Times the regression code from Miller (2015), Exhibit 1.1 (statsmodels OLS on
an add_constant design matrix, then summary()), plus the leaner
scipy.stats.linregress, so Python can be compared with the Go and R programs.
It then runs the same Monte Carlo benchmark as the Go program: draw many
synthetic samples from y = 3 + 0.5x + N(0, 1.24^2), fit each, and report the
mean slope, 95% confidence-interval coverage and time per run.

Usage:  python3 python/anscombe_benchmark.py [reps] [monte_carlo_runs] [n]
"""

import sys
import time

import numpy as np
import pandas as pd
import statsmodels.api as sm
from scipy import stats

# The Anscombe data frame exactly as defined in Miller (2015), Exhibit 1.1.
anscombe = pd.DataFrame({'x1' : [10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5],
    'x2' : [10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5],
    'x3' : [10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5],
    'x4' : [8, 8, 8, 8, 8, 8, 8, 19, 8, 8, 8],
    'y1' : [8.04, 6.95,  7.58, 8.81, 8.33, 9.96, 7.24, 4.26,10.84, 4.82, 5.68],
    'y2' : [9.14, 8.14,  8.74, 8.77, 9.26, 8.1, 6.13, 3.1,  9.13, 7.26, 4.74],
    'y3' : [7.46, 6.77, 12.74, 7.11, 7.81, 8.84, 6.08, 5.39, 8.15, 6.42, 5.73],
    'y4' : [6.58, 5.76, 7.71, 8.84, 8.47, 7.04, 5.25, 12.5, 5.56, 7.91, 6.89]})

SET_NAMES = ["Set I", "Set II", "Set III", "Set IV"]
MC_INTERCEPT, MC_SLOPE, MC_SIGMA = 3.0, 0.5, 1.24


def miller_fit(i: int):
    """Fit set i (1-4) exactly as Miller's Exhibit 1.1 does."""
    design_matrix = sm.add_constant(anscombe[f'x{i}'])
    model = sm.OLS(anscombe[f'y{i}'], design_matrix)
    return model.fit()


def print_results() -> None:
    print(f"{'Set':<8} {'Intercept':>12} {'Slope':>12} {'Int SE':>12} "
          f"{'Slope SE':>12} {'R^2':>10}")
    for i, name in enumerate(SET_NAMES, start=1):
        fit = miller_fit(i)
        print(f"{name:<8} {fit.params['const']:12.6f} {fit.params[f'x{i}']:12.6f} "
              f"{fit.bse['const']:12.6f} {fit.bse[f'x{i}']:12.6f} {fit.rsquared:10.6f}")


def time_miller(reps: int) -> None:
    """Time Miller's steps: design matrix, OLS fit and summary table."""
    start = time.perf_counter()
    for _ in range(reps):
        for i in range(1, 5):
            miller_fit(i).summary()
    elapsed = time.perf_counter() - start
    print(f"\nPython statsmodels (Miller's code, with summary()): {reps} reps x 4 sets "
          f"in {elapsed:.3f} s ({elapsed / (reps * 4) * 1e6:.2f} microseconds per fit)")


def time_lean(reps: int) -> None:
    """Time scipy.stats.linregress on plain arrays (no data frames or summary)."""
    arrays = [(anscombe[f'x{i}'].to_numpy(float), anscombe[f'y{i}'].to_numpy(float))
              for i in range(1, 5)]
    start = time.perf_counter()
    for _ in range(reps):
        for x, y in arrays:
            stats.linregress(x, y)
    elapsed = time.perf_counter() - start
    print(f"Python scipy linregress: {reps} reps x 4 sets in {elapsed:.3f} s "
          f"({elapsed / (reps * 4) * 1e6:.2f} microseconds per fit)")


def monte_carlo(runs: int, n: int, seed: int = 1) -> None:
    """Simulate `runs` samples of size n from a known line and fit each one."""
    rng = np.random.default_rng(seed)
    t_crit = stats.t.ppf(0.975, n - 2)
    slopes = np.empty(runs)
    covered = 0
    start = time.perf_counter()
    for i in range(runs):
        x = 10 * rng.random(n)
        y = MC_INTERCEPT + MC_SLOPE * x + MC_SIGMA * rng.standard_normal(n)
        fit = stats.linregress(x, y)
        slopes[i] = fit.slope
        if abs(fit.slope - MC_SLOPE) <= t_crit * fit.stderr:
            covered += 1
    elapsed = time.perf_counter() - start
    print(f"\nMonte Carlo: {runs} samples of n = {n} from y = {MC_INTERCEPT} + "
          f"{MC_SLOPE}x + N(0, {MC_SIGMA}^2)")
    print(f"  mean slope {slopes.mean():.5f} (true {MC_SLOPE}), SD {slopes.std():.5f}, "
          f"95% CI coverage {100 * covered / runs:.1f}%")
    print(f"  Python: {elapsed:.3f} s total ({elapsed / runs * 1e6:.2f} microseconds "
          f"per simulate + fit)")


def print_peak_memory() -> None:
    try:  # resource is Unix-only; skip the memory line on Windows.
        import resource
    except ImportError:
        return
    peak = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
    scale = 1e6 if sys.platform == "darwin" else 1e3  # bytes on macOS, KB on Linux
    print(f"Memory: peak resident set size {peak / scale:.1f} MB")


def read_count(position: int, default: int, minimum: int) -> int:
    """Read a whole-number argument, explaining any bad value."""
    if len(sys.argv) <= position:
        return default
    try:
        value = int(sys.argv[position])
    except ValueError:
        value = minimum - 1
    if value < minimum:
        sys.exit(f"error: argument {position} must be a whole number >= {minimum}; "
                 f"got {sys.argv[position]!r}\n\n{__doc__}")
    return value


def main() -> None:
    reps = read_count(1, 1000, 0)
    mc_runs = read_count(2, 10000, 0)
    mc_n = read_count(3, 100, 3)

    print_results()
    if reps > 0:
        time_miller(reps)
        time_lean(reps * 10)
    if mc_runs > 0:
        monte_carlo(mc_runs, mc_n)
    print_peak_memory()


if __name__ == "__main__":
    main()
