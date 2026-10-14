package anscombe

// Duet returns two data sets of my own, answering the suggestion at the end
// of Miller's (2015) Anscombe programs: "See if you can develop a quartet of
// your own, or perhaps just a duet, two very different data sets with the
// same fitted model."
//
// Set A is a straight line with random scatter. Set B is a smooth U-shaped
// curve. Both give y = 2.52 + 0.555x with R-squared 0.81 (to rounding).
//
// How they were built: Set A came from y = 2 + 0.6x + N(0, 1). Set B keeps
// Set A's fitted values and replaces its residuals with a quadratic pattern
// that is uncorrelated with x and has the same sum of squares. That leaves
// the coefficients, standard errors and R-squared unchanged. Both sets were
// then rounded to two decimals, as Anscombe's data are.
func Duet() []DataSet {
	x := []float64{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	return []DataSet{
		{
			Name: "Duet A",
			X:    copyOf(x),
			Y:    []float64{3.23, 5.16, 5.62, 4.49, 5.30, 5.67, 7.37, 7.34, 8.75, 6.75, 10.77, 9.70},
		},
		{
			Name: "Duet B",
			X:    copyOf(x),
			Y:    []float64{5.22, 4.91, 4.77, 4.80, 5.01, 5.39, 5.95, 6.67, 7.58, 8.65, 9.90, 11.32},
		},
	}
}
