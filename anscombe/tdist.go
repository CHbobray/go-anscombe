package anscombe

import "math"

// TwoSidedPValue returns P(|T| > |t|) for Student's t distribution with df
// degrees of freedom: the p-value R prints as Pr(>|t|).
//
// It uses the identity P(|T| > |t|) = I_x(df/2, 1/2) with x = df/(df + t^2),
// where I_x is the regularized incomplete beta function.
func TwoSidedPValue(t float64, df int) float64 {
	if df < 1 || math.IsNaN(t) {
		return math.NaN()
	}
	if math.IsInf(t, 0) {
		return 0
	}
	d := float64(df)
	return regularizedIncompleteBeta(d/(d+t*t), d/2, 0.5)
}

// regularizedIncompleteBeta computes I_x(a, b) for 0 <= x <= 1.
func regularizedIncompleteBeta(x, a, b float64) float64 {
	switch {
	case x <= 0:
		return 0
	case x >= 1:
		return 1
	}

	// Front factor x^a (1-x)^b / (a B(a, b)), computed with log-gamma
	// to avoid overflow.
	lgA, _ := math.Lgamma(a)
	lgB, _ := math.Lgamma(b)
	lgAB, _ := math.Lgamma(a + b)
	front := math.Exp(lgAB - lgA - lgB + a*math.Log(x) + b*math.Log(1-x))

	// The continued fraction converges fastest when x < (a+1)/(a+b+2);
	// otherwise use the symmetry I_x(a, b) = 1 - I_{1-x}(b, a).
	if x < (a+1)/(a+b+2) {
		return front * betaContinuedFraction(x, a, b) / a
	}
	return 1 - front*betaContinuedFraction(1-x, b, a)/b
}

// betaContinuedFraction evaluates the continued fraction for the incomplete
// beta function with the modified Lentz method (Press et al., Numerical
// Recipes, section 6.4).
func betaContinuedFraction(x, a, b float64) float64 {
	const (
		maxIterations = 300
		epsilon       = 1e-15
		tiny          = 1e-300
	)
	c := 1.0
	d := 1 - (a+b)*x/(a+1)
	if math.Abs(d) < tiny {
		d = tiny
	}
	d = 1 / d
	result := d

	for m := 1; m <= maxIterations; m++ {
		fm := float64(m)

		// Even step of the recurrence.
		numerator := fm * (b - fm) * x / ((a + 2*fm - 1) * (a + 2*fm))
		d = 1 + numerator*d
		if math.Abs(d) < tiny {
			d = tiny
		}
		c = 1 + numerator/c
		if math.Abs(c) < tiny {
			c = tiny
		}
		d = 1 / d
		result *= d * c

		// Odd step of the recurrence.
		numerator = -(a + fm) * (a + b + fm) * x / ((a + 2*fm) * (a + 2*fm + 1))
		d = 1 + numerator*d
		if math.Abs(d) < tiny {
			d = tiny
		}
		c = 1 + numerator/c
		if math.Abs(c) < tiny {
			c = tiny
		}
		d = 1 / d
		delta := d * c
		result *= delta

		if math.Abs(delta-1) < epsilon {
			break
		}
	}
	return result
}

// CriticalT returns the value c with P(|T| > c) = alpha for Student's t
// with df degrees of freedom, e.g. CriticalT(0.05, 9) = 2.262, the
// multiplier for a 95% confidence interval (R: qt(0.975, 9)).
//
// TwoSidedPValue falls steadily as c grows, so bisection finds c to full
// floating-point precision.
func CriticalT(alpha float64, df int) float64 {
	if df < 1 || alpha <= 0 || alpha >= 1 {
		return math.NaN()
	}
	low, high := 0.0, 1.0
	for TwoSidedPValue(high, df) > alpha { // widen until the answer is bracketed
		high *= 2
	}
	for i := 0; i < 200 && high-low > 1e-14*high; i++ {
		mid := (low + high) / 2
		if TwoSidedPValue(mid, df) > alpha {
			low = mid
		} else {
			high = mid
		}
	}
	return (low + high) / 2
}
