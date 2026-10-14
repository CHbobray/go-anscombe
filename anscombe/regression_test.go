package anscombe

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

// rReference holds summary(lm(y ~ x)) output from R 4.3.3 for each Anscombe
// set, printed to 12 decimals (p-values to 12 significant digits).
// Python statsmodels OLS gives the same values (see the python folder).
var rReference = []Result{
	{N: 11, DF: 9, MeanX: 9.000000000000, MeanY: 7.500909090909,
		Intercept: 3.000090909091, InterceptSE: 1.124746790809, InterceptT: 2.667347827624, InterceptP: 0.0257340513992,
		Slope: 0.500090909091, SlopeSE: 0.117905500596, SlopeT: 4.241455288893, SlopeP: 0.00216962887308,
		ResidualSE: 1.236603322726, RSquared: 0.666542459509, AdjRSquared: 0.629491621676, FStatistic: 17.989942967677, FPValue: 0.00216962887308},
	{N: 11, DF: 9, MeanX: 9.000000000000, MeanY: 7.500909090909,
		Intercept: 3.000909090909, InterceptSE: 1.125302416245, InterceptT: 2.666757884447, InterceptP: 0.0257589410308,
		Slope: 0.500000000000, SlopeSE: 0.117963745968, SlopeT: 4.238590389772, SlopeP: 0.00217881623691,
		ResidualSE: 1.237214205342, RSquared: 0.666242033727, AdjRSquared: 0.629157815253, FStatistic: 17.965648492271, FPValue: 0.00217881623691},
	{N: 11, DF: 9, MeanX: 9.000000000000, MeanY: 7.500000000000,
		Intercept: 3.002454545455, InterceptSE: 1.124481229640, InterceptT: 2.670079736605, InterceptP: 0.0256191088395,
		Slope: 0.499727272727, SlopeSE: 0.117877662221, SlopeT: 4.239372102497, SlopeP: 0.00217630527923,
		ResidualSE: 1.236311351390, RSquared: 0.666324041067, AdjRSquared: 0.629248934518, FStatistic: 17.972275823429, FPValue: 0.00217630527923},
	{N: 11, DF: 9, MeanX: 9.000000000000, MeanY: 7.500909090909,
		Intercept: 3.001727272727, InterceptSE: 1.123921071854, InterceptT: 2.670763408480, InterceptP: 0.0255904252008,
		Slope: 0.499909090909, SlopeSE: 0.117818941730, SlopeT: 4.243028188592, SlopeP: 0.0021646023472,
		ResidualSE: 1.235695485681, RSquared: 0.666707256898, AdjRSquared: 0.629674729887, FStatistic: 18.003288209183, FPValue: 0.0021646023472},
}

// tolerance allows for the 12-decimal rounding of the reference values and
// floating-point differences between Go's summation and R's QR decomposition.
const tolerance = 1e-9

func near(a, b float64) bool { return math.Abs(a-b) <= tolerance }

// TestFitMatchesR checks every reported statistic against R for all four sets.
func TestFitMatchesR(t *testing.T) {
	for i, ds := range Quartet() {
		want := rReference[i]
		t.Run(ds.Name, func(t *testing.T) {
			got, err := Fit(ds.X, ds.Y)
			if err != nil {
				t.Fatalf("Fit returned error: %v", err)
			}
			if got.N != want.N || got.DF != want.DF {
				t.Errorf("N, DF = %d, %d; want %d, %d", got.N, got.DF, want.N, want.DF)
			}
			checks := []struct {
				name      string
				got, want float64
			}{
				{"MeanX", got.MeanX, want.MeanX},
				{"MeanY", got.MeanY, want.MeanY},
				{"Intercept", got.Intercept, want.Intercept},
				{"InterceptSE", got.InterceptSE, want.InterceptSE},
				{"InterceptT", got.InterceptT, want.InterceptT},
				{"InterceptP", got.InterceptP, want.InterceptP},
				{"Slope", got.Slope, want.Slope},
				{"SlopeSE", got.SlopeSE, want.SlopeSE},
				{"SlopeT", got.SlopeT, want.SlopeT},
				{"SlopeP", got.SlopeP, want.SlopeP},
				{"ResidualSE", got.ResidualSE, want.ResidualSE},
				{"RSquared", got.RSquared, want.RSquared},
				{"AdjRSquared", got.AdjRSquared, want.AdjRSquared},
				{"FStatistic", got.FStatistic, want.FStatistic},
				{"FPValue", got.FPValue, want.FPValue},
			}
			for _, c := range checks {
				if !near(c.got, c.want) {
					t.Errorf("%s = %.12f, want %.12f (R)", c.name, c.got, c.want)
				}
			}
		})
	}
}

// TestQuartetLesson confirms the summary Miller (2015, 9) gives for the
// quartet: in every set the mean of x is 9, the mean of y is 7.5, the fitted
// line is y = 3 + 0.5x, and R-squared is 0.67, even though the plotted data
// look nothing alike.
func TestQuartetLesson(t *testing.T) {
	for _, ds := range Quartet() {
		got, err := Fit(ds.X, ds.Y)
		if err != nil {
			t.Fatalf("%s: %v", ds.Name, err)
		}
		if got.MeanX != 9 || math.Abs(got.MeanY-7.5) > 0.001 {
			t.Errorf("%s: means x = %.4f, y = %.4f; want 9 and 7.5", ds.Name, got.MeanX, got.MeanY)
		}
		if math.Abs(got.Intercept-3.00) > 0.005 ||
			math.Abs(got.Slope-0.500) > 0.001 ||
			math.Abs(got.RSquared-0.67) > 0.005 {
			t.Errorf("%s: got y = %.4f + %.4fx, R^2 = %.4f; want about y = 3.00 + 0.500x, R^2 = 0.67",
				ds.Name, got.Intercept, got.Slope, got.RSquared)
		}
	}
}

// TestFitPerfectLine checks a case with a known exact answer.
func TestFitPerfectLine(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{3, 5, 7, 9, 11} // y = 1 + 2x
	got, err := Fit(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !near(got.Intercept, 1) || !near(got.Slope, 2) || !near(got.RSquared, 1) || !near(got.ResidualSE, 0) {
		t.Errorf("got %+v, want intercept 1, slope 2, R^2 1, residual SE 0", got)
	}
}

// TestFitErrors checks that bad input is rejected with the right error.
func TestFitErrors(t *testing.T) {
	cases := []struct {
		name string
		x, y []float64
		want error
	}{
		{"length mismatch", []float64{1, 2, 3}, []float64{1, 2}, ErrLengthMismatch},
		{"too few points", []float64{1, 2}, []float64{1, 2}, ErrTooFewPoints},
		{"empty", nil, nil, ErrTooFewPoints},
		{"constant x", []float64{4, 4, 4}, []float64{1, 2, 3}, ErrConstantX},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Fit(c.x, c.y); !errors.Is(err, c.want) {
				t.Errorf("got error %v, want %v", err, c.want)
			}
		})
	}
}

// TestQuartetReturnsCopies ensures callers cannot change the shared data.
func TestQuartetReturnsCopies(t *testing.T) {
	q := Quartet()
	q[0].X[0] = -999
	if Quartet()[0].X[0] == -999 || Quartet()[1].X[0] == -999 {
		t.Error("Quartet shares its slices between calls")
	}
}

// BenchmarkFit times one regression on each Anscombe set.
// Run with: go test -bench=. -benchmem ./...
func BenchmarkFit(b *testing.B) {
	for _, ds := range Quartet() {
		b.Run(ds.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := Fit(ds.X, ds.Y); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkFitQuartet times fitting all four sets, the unit used by the
// Python and R timing loops.
func BenchmarkFitQuartet(b *testing.B) {
	q := Quartet()
	for i := 0; i < b.N; i++ {
		for _, ds := range q {
			if _, err := Fit(ds.X, ds.Y); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func ExampleFit() {
	ds := Quartet()[0]
	r, _ := Fit(ds.X, ds.Y)
	fmt.Printf("y = %.2f + %.3fx, R^2 = %.2f\n", r.Intercept, r.Slope, r.RSquared)
	// Output: y = 3.00 + 0.500x, R^2 = 0.67
}
