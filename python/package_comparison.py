"""Compare Python packages for simple linear regression on the Anscombe Quartet.

The assignment notes that different Python packages may give different
results: statsmodels with Patsy formulas, statsmodels with arrays, NumPy or
SciPy used directly, and scikit-learn. This program fits all four Anscombe
sets with each package and reports

  * the largest difference from R's lm() coefficients (R values to 12
    decimals, the same reference the Go tests use),
  * which summary statistics the package provides, and
  * the time per fit.

Usage:  python3 python/package_comparison.py [reps]     (default 2000)
"""

import sys
import time

import numpy as np
import pandas as pd
import statsmodels.api as sm
import statsmodels.formula.api as smf
from scipy import stats
from sklearn.linear_model import LinearRegression

# Anscombe (1973) data, entered exactly as in Miller (2015), Exhibit 1.1.
anscombe = pd.DataFrame({'x1': [10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5],
    'x2': [10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5],
    'x3': [10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5],
    'x4': [8, 8, 8, 8, 8, 8, 8, 19, 8, 8, 8],
    'y1': [8.04, 6.95, 7.58, 8.81, 8.33, 9.96, 7.24, 4.26, 10.84, 4.82, 5.68],
    'y2': [9.14, 8.14, 8.74, 8.77, 9.26, 8.1, 6.13, 3.1, 9.13, 7.26, 4.74],
    'y3': [7.46, 6.77, 12.74, 7.11, 7.81, 8.84, 6.08, 5.39, 8.15, 6.42, 5.73],
    'y4': [6.58, 5.76, 7.71, 8.84, 8.47, 7.04, 5.25, 12.5, 5.56, 7.91, 6.89]})

# (intercept, slope) from R's summary(lm(y ~ x)), printed to 12 decimals.
R_COEFFICIENTS = [(3.000090909091, 0.500090909091), (3.000909090909, 0.500000000000),
                  (3.002454545455, 0.499727272727), (3.001727272727, 0.499909090909)]


# Each function fits one set and returns (intercept, slope).

def statsmodels_arrays(i):
    """Miller's approach: statsmodels OLS on an add_constant design matrix."""
    fit = sm.OLS(anscombe[f'y{i}'], sm.add_constant(anscombe[f'x{i}'])).fit()
    return fit.params.iloc[0], fit.params.iloc[1]


def statsmodels_patsy(i):
    """statsmodels formula interface; Patsy parses 'y ~ x' into a design matrix."""
    fit = smf.ols(f'y{i} ~ x{i}', data=anscombe).fit()
    return fit.params.iloc[0], fit.params.iloc[1]


def numpy_polyfit(i):
    """NumPy directly: fit a degree-1 polynomial (returns slope first)."""
    slope, intercept = np.polyfit(anscombe[f'x{i}'], anscombe[f'y{i}'], 1)
    return intercept, slope


def numpy_lstsq(i):
    """NumPy directly: solve the least squares problem [1, x] b = y."""
    x = anscombe[f'x{i}'].to_numpy(float)
    design = np.column_stack([np.ones_like(x), x])
    coefficients, *_ = np.linalg.lstsq(design, anscombe[f'y{i}'].to_numpy(float), rcond=None)
    return coefficients[0], coefficients[1]


def scipy_linregress(i):
    """SciPy directly: closed-form simple regression."""
    fit = stats.linregress(anscombe[f'x{i}'], anscombe[f'y{i}'])
    return fit.intercept, fit.slope


def sklearn_linear(i):
    """scikit-learn: the machine-learning estimator interface."""
    model = LinearRegression().fit(anscombe[[f'x{i}']], anscombe[f'y{i}'])
    return model.intercept_, model.coef_[0]


# name, function, statistics the package reports beyond the coefficients
METHODS = [
    ("statsmodels OLS, arrays (Miller)", statsmodels_arrays,
     "SEs, t, p-values, R², F, CIs, full summary()"),
    ("statsmodels formula (Patsy)", statsmodels_patsy,
     "SEs, t, p-values, R², F, CIs, full summary()"),
    ("NumPy polyfit", numpy_polyfit, "coefficients only"),
    ("NumPy linalg.lstsq", numpy_lstsq, "coefficients and residual sum of squares"),
    ("SciPy stats.linregress", scipy_linregress, "slope SE, intercept SE, r, p-value for slope"),
    ("scikit-learn LinearRegression", sklearn_linear, "coefficients and R² (via score()); no SEs or p-values"),
]


def max_difference_from_r(fit_one_set) -> float:
    """Largest absolute coefficient difference from R across the four sets."""
    worst = 0.0
    for i, (r_intercept, r_slope) in enumerate(R_COEFFICIENTS, start=1):
        intercept, slope = fit_one_set(i)
        worst = max(worst, abs(intercept - r_intercept), abs(slope - r_slope))
    return worst


def microseconds_per_fit(fit_one_set, reps: int) -> float:
    start = time.perf_counter()
    for _ in range(reps):
        for i in range(1, 5):
            fit_one_set(i)
    return (time.perf_counter() - start) / (reps * 4) * 1e6


def main() -> None:
    try:
        reps = int(sys.argv[1]) if len(sys.argv) > 1 else 2000
    except ValueError:
        reps = 0
    if reps < 1:
        sys.exit(f"error: reps must be a whole number of at least 1\n\n{__doc__}")

    print("Coefficients for each set (intercept, slope):")
    for name, fit_one_set, _ in METHODS:
        pairs = "  ".join(f"({a:.6f}, {b:.6f})" for a, b in (fit_one_set(i) for i in range(1, 5)))
        print(f"  {name:<32} {pairs}")

    print(f"\n| {'Package':<32} | {'Max diff from R':>15} | {'µs per fit':>10} | Other statistics provided |")
    print(f"|{'-' * 34}|{'-' * 16}:|{'-' * 11}:|{'-' * 27}|")
    for name, fit_one_set, extras in METHODS:
        diff = max_difference_from_r(fit_one_set)
        speed = microseconds_per_fit(fit_one_set, reps)
        print(f"| {name:<32} | {diff:>15.1e} | {speed:>10.1f} | {extras} |", flush=True)
    print(f"\n{reps} reps x 4 sets per package. R values are rounded to 12 decimals, "
          "so differences below 1e-12 mean exact agreement.")


if __name__ == "__main__":
    main()
