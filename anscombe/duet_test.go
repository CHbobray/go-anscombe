package anscombe

import (
	"math"
	"testing"

	"github.com/montanaflynn/stats"
)

// TestDuetSameModel checks that the two duet sets give the same fitted
// model to within the effect of rounding the data to two decimals.
func TestDuetSameModel(t *testing.T) {
	d := Duet()
	a, err := Fit(d[0].X, d[0].Y)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Fit(d[1].X, d[1].Y)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(a.Intercept-b.Intercept) > 0.01 ||
		math.Abs(a.Slope-b.Slope) > 0.001 ||
		math.Abs(a.RSquared-b.RSquared) > 0.001 ||
		math.Abs(a.SlopeSE-b.SlopeSE) > 0.001 {
		t.Errorf("duet fits differ:\nA: %+v\nB: %+v", a, b)
	}
}

// TestDuetDifferentShapes checks that the duet really is "two very different
// data sets": Set B's residuals follow a U-shaped curve, Set A's do not.
func TestDuetDifferentShapes(t *testing.T) {
	for i, wantCurved := range []bool{false, true} {
		ds := Duet()[i]
		r, err := Fit(ds.X, ds.Y)
		if err != nil {
			t.Fatal(err)
		}
		residuals := make([]float64, len(ds.X))
		curve := make([]float64, len(ds.X))
		for k := range ds.X {
			residuals[k] = ds.Y[k] - (r.Intercept + r.Slope*ds.X[k])
			curve[k] = (ds.X[k] - r.MeanX) * (ds.X[k] - r.MeanX)
		}
		corr, err := stats.Correlation(residuals, curve)
		if err != nil {
			t.Fatal(err)
		}
		curved := corr > 0.95
		if curved != wantCurved {
			t.Errorf("%s: correlation of residuals with a U-shaped curve = %.3f", ds.Name, corr)
		}
	}
}
