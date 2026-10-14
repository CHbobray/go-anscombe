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

// Result holds the ordinary least squares estimates for y = b0 + b1*x,
// matching the quantities reported by R's summary(lm(y ~ x)) and
// Python's statsmodels OLS.
type Result struct {
	N           int     // number of observations
	Intercept   float64 // b0
	Slope       float64 // b1
	InterceptSE float64 // standard error of b0
	SlopeSE     float64 // standard error of b1
	SlopeT      float64 // t statistic for b1
	RSquared    float64 // coefficient of determination
	ResidualSE  float64 // residual standard error, sqrt(SSE / (n - 2))
}

// Fit runs a simple linear regression of y on x.
//
// The fitted line comes from stats.LinearRegression. That function returns
// fitted points rather than coefficients, so the slope and intercept are
// recovered from two fitted points with different x values. R-squared uses
// stats.Correlation, and standard errors use the usual OLS formulas.
func Fit(x, y []float64) (Result, error) {
	n := len(x)
	if n != len(y) {
		return Result{}, ErrLengthMismatch
	}
	if n < 3 {
		return Result{}, ErrTooFewPoints
	}

	series := make(stats.Series, n)
	for i := range x {
		series[i] = stats.Coordinate{X: x[i], Y: y[i]}
	}

	fitted, err := stats.LinearRegression(series)
	if err != nil {
		return Result{}, fmt.Errorf("anscombe: linear regression: %w", err)
	}

	// Find a point whose x differs from the first one.
	j := -1
	for k := 1; k < n; k++ {
		if x[k] != x[0] {
			j = k
			break
		}
	}
	if j < 0 {
		return Result{}, ErrConstantX
	}

	slope := (fitted[j].Y - fitted[0].Y) / (fitted[j].X - fitted[0].X)
	intercept := fitted[0].Y - slope*fitted[0].X

	r, err := stats.Correlation(x, y)
	if err != nil {
		return Result{}, fmt.Errorf("anscombe: correlation: %w", err)
	}

	meanX, _ := stats.Mean(x)
	popVarX, _ := stats.PopulationVariance(x)
	sxx := popVarX * float64(n) // sum of squared deviations of x

	var sse float64
	for i := range y {
		resid := y[i] - fitted[i].Y
		sse += resid * resid
	}
	df := float64(n - 2)
	sigma := math.Sqrt(sse / df)
	slopeSE := sigma / math.Sqrt(sxx)
	interceptSE := sigma * math.Sqrt(1/float64(n)+meanX*meanX/sxx)

	return Result{
		N:           n,
		Intercept:   intercept,
		Slope:       slope,
		InterceptSE: interceptSE,
		SlopeSE:     slopeSE,
		SlopeT:      slope / slopeSE,
		RSquared:    r * r,
		ResidualSE:  sigma,
	}, nil
}
