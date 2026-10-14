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
	meanY, _ := stats.Mean(y)
	popVarX, _ := stats.PopulationVariance(x)
	sxx := popVarX * float64(n) // sum of squared deviations of x

	var sse float64
	for i := range y {
		resid := y[i] - fitted[i].Y
		sse += resid * resid
	}
	df := n - 2
	sigma := math.Sqrt(sse / float64(df))
	slopeSE := sigma / math.Sqrt(sxx)
	interceptSE := sigma * math.Sqrt(1/float64(n)+meanX*meanX/sxx)

	slopeT := slope / slopeSE
	interceptT := intercept / interceptSE
	slopeP := TwoSidedPValue(slopeT, df)
	rSquared := r * r

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
		FStatistic:  slopeT * slopeT,
		FPValue:     slopeP,
	}, nil
}
