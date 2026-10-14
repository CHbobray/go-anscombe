package anscombe

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

// rReference holds summary(lm(y ~ x)) output from R 4.3.3 for each Anscombe
// set, printed to 12 decimals. Python statsmodels OLS gives the same values
// (see python/anscombe_regression.py).
var rReference = []Result{
	{N: 11, Intercept: 3.000090909091, Slope: 0.500090909091, InterceptSE: 1.124746790809, SlopeSE: 0.117905500596, SlopeT: 4.241455288893, RSquared: 0.666542459509, ResidualSE: 1.236603322726},
	{N: 11, Intercept: 3.000909090909, Slope: 0.500000000000, InterceptSE: 1.125302416245, SlopeSE: 0.117963745968, SlopeT: 4.238590389772, RSquared: 0.666242033727, ResidualSE: 1.237214205342},
	{N: 11, Intercept: 3.002454545455, Slope: 0.499727272727, InterceptSE: 1.124481229640, SlopeSE: 0.117877662221, SlopeT: 4.239372102497, RSquared: 0.666324041067, ResidualSE: 1.236311351390},
	{N: 11, Intercept: 3.001727272727, Slope: 0.499909090909, InterceptSE: 1.123921071854, SlopeSE: 0.117818941730, SlopeT: 4.243028188592, RSquared: 0.666707256898, ResidualSE: 1.235695485681},
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
			if got.N != want.N {
				t.Errorf("N = %d, want %d", got.N, want.N)
			}
			checks := []struct {
				name      string
				got, want float64
			}{
				{"Intercept", got.Intercept, want.Intercept},
				{"Slope", got.Slope, want.Slope},
				{"InterceptSE", got.InterceptSE, want.InterceptSE},
				{"SlopeSE", got.SlopeSE, want.SlopeSE},
				{"SlopeT", got.SlopeT, want.SlopeT},
				{"RSquared", got.RSquared, want.RSquared},
				{"ResidualSE", got.ResidualSE, want.ResidualSE},
			}
			for _, c := range checks {
				if !near(c.got, c.want) {
					t.Errorf("%s = %.12f, want %.12f (R)", c.name, c.got, c.want)
				}
			}
		})
	}
}

// TestQuartetLesson confirms Anscombe's point: the four sets are visibly
// different yet give practically the same line (about y = 3.00 + 0.500x)
// and R-squared (about 0.67).
func TestQuartetLesson(t *testing.T) {
	for _, ds := range Quartet() {
		got, err := Fit(ds.X, ds.Y)
		if err != nil {
			t.Fatalf("%s: %v", ds.Name, err)
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
