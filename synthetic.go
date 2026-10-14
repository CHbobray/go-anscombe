package anscombe

import "math/rand"

// Synthetic generates n observations from the known model
// y = intercept + slope*x + e, where x is uniform on [0, 10) and
// e is normal with mean 0 and standard deviation sigma.
//
// Because the true coefficients are known, synthetic data lets tests and
// Monte Carlo simulations check that Fit recovers them.
func Synthetic(rng *rand.Rand, n int, intercept, slope, sigma float64) DataSet {
	ds := DataSet{
		Name: "Synthetic",
		X:    make([]float64, n),
		Y:    make([]float64, n),
	}
	for i := 0; i < n; i++ {
		x := 10 * rng.Float64()
		ds.X[i] = x
		ds.Y[i] = intercept + slope*x + sigma*rng.NormFloat64()
	}
	return ds
}
