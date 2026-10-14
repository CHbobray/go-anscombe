package anscombe

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

// TestFitRecoversSyntheticModel fits a large synthetic sample whose true
// line is known and checks that the estimates land close to it.
func TestFitRecoversSyntheticModel(t *testing.T) {
	const trueIntercept, trueSlope, sigma = 2.0, 0.75, 1.0
	rng := rand.New(rand.NewSource(42)) // fixed seed so the test is repeatable

	ds := Synthetic(rng, 10000, trueIntercept, trueSlope, sigma)
	got, err := Fit(ds.X, ds.Y)
	if err != nil {
		t.Fatal(err)
	}

	// With 10,000 points the slope SE is about 0.0035, so 4 SEs is a safe band.
	if math.Abs(got.Slope-trueSlope) > 4*got.SlopeSE {
		t.Errorf("slope = %.4f, want %.2f ± %.4f", got.Slope, trueSlope, 4*got.SlopeSE)
	}
	if math.Abs(got.Intercept-trueIntercept) > 4*got.InterceptSE {
		t.Errorf("intercept = %.4f, want %.2f ± %.4f", got.Intercept, trueIntercept, 4*got.InterceptSE)
	}
	if math.Abs(got.ResidualSE-sigma) > 0.05 {
		t.Errorf("residual SE = %.4f, want about %.2f", got.ResidualSE, sigma)
	}
}

// TestStandardErrorCoverage is a small Monte Carlo check of the standard
// errors: across many synthetic samples, the 95% confidence interval for
// the slope should contain the true slope about 95% of the time.
func TestStandardErrorCoverage(t *testing.T) {
	const trueSlope, runs, n = 0.5, 2000, 30
	const tCrit = 2.048 // t(0.975, df = 28)
	rng := rand.New(rand.NewSource(7))

	covered := 0
	for i := 0; i < runs; i++ {
		ds := Synthetic(rng, n, 3, trueSlope, 1.2)
		r, err := Fit(ds.X, ds.Y)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(r.Slope-trueSlope) <= tCrit*r.SlopeSE {
			covered++
		}
	}
	rate := float64(covered) / runs
	// Binomial SE of the rate is about 0.005; allow roughly 4 SEs.
	if rate < 0.93 || rate > 0.97 {
		t.Errorf("95%% CI coverage = %.3f, want about 0.95", rate)
	}
}

// BenchmarkFitSynthetic shows how Fit scales with sample size.
func BenchmarkFitSynthetic(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		ds := Synthetic(rand.New(rand.NewSource(1)), n, 3, 0.5, 1)
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := Fit(ds.X, ds.Y); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
