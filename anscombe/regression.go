package anscombe

import (
	"errors"
	"fmt"
	"math"

	"github.com/montanaflynn/stats"
)

// Errors returned by Fit when the input cannot support a regression.
var (
	ErrLengthMismatch = errors.New("anscombe: x and y must have the same length")
	ErrTooFewPoints   = errors.New("anscombe: need at least 3 observations")
	ErrConstantX      = errors.New("anscombe: x has zero variance, slope is undefined")
)

// Result holds the ordinary least squares estimates for y = b0 + b1*x:
// the same quantities reported by R's summary(lm(y ~ x)) and by
// statsmodels' OLS(...).fit().summary() in Miller's (2015) programs.
type Result struct {
	N  int // number of observations
	DF int // residual degrees of freedom, n - 2

	MeanX float64 // mean of the explanatory variable
	MeanY float64 // mean of the response

	Intercept   float64 // b0
	InterceptSE float64 // standard error of b0
	InterceptT  float64 // t statistic for b0
	InterceptP  float64 // two-sided p-value for b0

	Slope   float64 // b1
	SlopeSE float64 // standard error of b1
	SlopeT  float64 // t statistic for b1
	SlopeP  float64 // two-sided p-value for b1

	ResidualSE  float64 // residual standard error, sqrt(SSE / (n - 2))
	RSquared    float64 // coefficient of determination
	AdjRSquared float64 // R-squared adjusted for the number of predictors
	FStatistic  float64 // overall F test; with one predictor F = SlopeT^2
	FPValue     float64 // p-value of the F test (equal to SlopeP here)
}

// Fit runs a simple linear regression of y on x.
//
// It works in three steps: check the input, fit the line with the stats
// package, then compute the inference statistics R's summary(lm()) reports.
func Fit(x, y []float64) (Result, error) {
	if err := checkInput(x, y); err != nil {
		return Result{}, err
	}
	intercept, slope, fitted, err := fitLine(x, y)
	if err != nil {
		return Result{}, err
	}
	r, err := stats.Correlation(x, y)
	if err != nil {
		return Result{}, fmt.Errorf("anscombe: correlation: %w", err)
	}
	return summarize(x, y, fitted, intercept, slope, r*r), nil
}

// checkInput rejects data that cannot support a regression line.
func checkInput(x, y []float64) error {
	switch {
	case len(x) != len(y):
		return ErrLengthMismatch
	case len(x) < 3:
		return ErrTooFewPoints
	case allEqual(x):
		return ErrConstantX
	}
	return nil
}

// allEqual reports whether every value in v equals the first one.
func allEqual(v []float64) bool {
	for _, value := range v[1:] {
		if value != v[0] {
			return false
		}
	}
	return true
}

// fitLine gets the least squares line, and the fitted point for each
// observation, from stats.LinearRegression.
//
// That function returns fitted points rather than coefficients, so the slope
// is recovered from two fitted points with different x values, and the
// intercept from where the line crosses x = 0. checkInput guarantees such a
// pair of points exists.
func fitLine(x, y []float64) (intercept, slope float64, fitted stats.Series, err error) {
	series := make(stats.Series, len(x))
	for i := range x {
		series[i] = stats.Coordinate{X: x[i], Y: y[i]}
	}
	line, err := stats.LinearRegression(series)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("anscombe: linear regression: %w", err)
	}

	first, other := line[0], line[indexOfDifferentX(x)]
	slope = (other.Y - first.Y) / (other.X - first.X)
	intercept = first.Y - slope*first.X
	return intercept, slope, line, nil
}

// indexOfDifferentX returns the index of the first x that differs from x[0].
func indexOfDifferentX(x []float64) int {
	for i, value := range x {
		if value != x[0] {
			return i
		}
	}
	return 0 // unreachable after checkInput
}

// summarize computes standard errors, t and F statistics, and p-values with
// the textbook OLS formulas for one predictor.
func summarize(x, y []float64, fitted stats.Series, intercept, slope, rSquared float64) Result {
	n := len(x)
	df := n - 2
	meanX, _ := stats.Mean(x)
	meanY, _ := stats.Mean(y)
	popVarX, _ := stats.PopulationVariance(x)
	sxx := popVarX * float64(n) // sum of squared deviations of x

	var sse float64 // sum of squared residuals
	for i := range y {
		residual := y[i] - fitted[i].Y
		sse += residual * residual
	}
	sigma := math.Sqrt(sse / float64(df))
	slopeSE := sigma / math.Sqrt(sxx)
	interceptSE := sigma * math.Sqrt(1/float64(n)+meanX*meanX/sxx)

	slopeT := slope / slopeSE
	interceptT := intercept / interceptSE
	slopeP := TwoSidedPValue(slopeT, df)

	return Result{
		N:           n,
		DF:          df,
		MeanX:       meanX,
		MeanY:       meanY,
		Intercept:   intercept,
		InterceptSE: interceptSE,
		InterceptT:  interceptT,
		InterceptP:  TwoSidedPValue(interceptT, df),
		Slope:       slope,
		SlopeSE:     slopeSE,
		SlopeT:      slopeT,
		SlopeP:      slopeP,
		ResidualSE:  sigma,
		RSquared:    rSquared,
		AdjRSquared: 1 - (1-rSquared)*float64(n-1)/float64(df),
		FStatistic:  slopeT * slopeT, // with one predictor, F = t^2
		FPValue:     slopeP,
	}
}
