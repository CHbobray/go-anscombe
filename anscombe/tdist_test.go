package anscombe

import (
	"math"
	"testing"
)

// TestTwoSidedPValueMatchesR compares p-values with R's 2 * pt(-|t|, df).
// The check is relative so that very small p-values are tested as strictly
// as large ones.
func TestTwoSidedPValueMatchesR(t *testing.T) {
	cases := []struct {
		t    float64
		df   int
		want float64
	}{
		{2.5, 3, 0.0877066470080655},
		{-2.5, 3, 0.0877066470080655}, // sign of t does not matter
		{0.3, 40, 0.765730716771019},
		{12, 100, 4.39508771560436e-21},
		{1, 1, 0.5},
		{4.2414552888931, 9, 0.00216962887307794},
		{0, 5, 1},
	}
	for _, c := range cases {
		got := TwoSidedPValue(c.t, c.df)
		if math.Abs(got-c.want) > 1e-10*c.want {
			t.Errorf("TwoSidedPValue(%g, %d) = %.15g, want %.15g", c.t, c.df, got, c.want)
		}
	}
}

func TestTwoSidedPValueEdgeCases(t *testing.T) {
	if p := TwoSidedPValue(math.Inf(1), 5); p != 0 {
		t.Errorf("infinite t: p = %g, want 0", p)
	}
	if p := TwoSidedPValue(2, 0); !math.IsNaN(p) {
		t.Errorf("zero df: p = %g, want NaN", p)
	}
	if p := TwoSidedPValue(math.NaN(), 5); !math.IsNaN(p) {
		t.Errorf("NaN t: p = %g, want NaN", p)
	}
}

// TestCriticalTMatchesR compares with R's qt(0.975, df).
func TestCriticalTMatchesR(t *testing.T) {
	cases := []struct {
		df   int
		want float64
	}{
		{1, 12.7062047361747},
		{9, 2.2621571627982},
		{28, 2.04840714179524},
		{98, 1.98446745450848},
		{1000, 1.96233908082641},
	}
	for _, c := range cases {
		if got := CriticalT(0.05, c.df); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("CriticalT(0.05, %d) = %.12f, want %.12f", c.df, got, c.want)
		}
	}
	for _, bad := range []float64{0, 1, -0.5, 2} {
		if got := CriticalT(bad, 5); !math.IsNaN(got) {
			t.Errorf("CriticalT(%g, 5) = %g, want NaN", bad, got)
		}
	}
}
