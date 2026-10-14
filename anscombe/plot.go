package anscombe

import (
	"fmt"
	"io"
	"math"
	"strings"
)

// Plot layout in SVG pixels. Each data set gets one panel; panels are laid
// out two per row, like the 2 x 2 grid in Miller's (2015) figure 1.5.
const (
	panelWidth   = 320
	panelHeight  = 260
	plotLeft     = 50 // room for y-axis labels
	plotRight    = 15
	plotTop      = 35 // room for the panel title
	plotBottom   = 45 // room for x-axis labels
	panelsPerRow = 2
)

// axisRange is the span shown on one axis, shared by every panel so the
// sets can be compared by eye (as Miller does with xlim and ylim).
type axisRange struct{ min, max float64 }

// WriteSVG draws a scatter plot of each data set with its fitted
// regression line and writes it to w as a standalone SVG image.
func WriteSVG(w io.Writer, sets []DataSet) error {
	if len(sets) == 0 {
		return ErrTooFewPoints
	}
	xr, yr := sharedRanges(sets)

	rows := (len(sets) + panelsPerRow - 1) / panelsPerRow
	cols := panelsPerRow
	if len(sets) < panelsPerRow {
		cols = len(sets)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" font-family="Helvetica, Arial, sans-serif" font-size="11">`+"\n",
		cols*panelWidth, rows*panelHeight)
	b.WriteString(`<rect width="100%" height="100%" fill="white"/>` + "\n")

	for i, ds := range sets {
		fit, err := Fit(ds.X, ds.Y)
		if err != nil {
			return fmt.Errorf("%s: %w", ds.Name, err)
		}
		offsetX := (i % panelsPerRow) * panelWidth
		offsetY := (i / panelsPerRow) * panelHeight
		drawPanel(&b, i, ds, fit, xr, yr, offsetX, offsetY)
	}

	b.WriteString("</svg>\n")
	_, err := io.WriteString(w, b.String())
	return err
}

// sharedRanges finds axis limits that hold every point in every set,
// widened to whole numbers with a margin of one unit.
func sharedRanges(sets []DataSet) (axisRange, axisRange) {
	xr := axisRange{math.Inf(1), math.Inf(-1)}
	yr := xr
	for _, ds := range sets {
		for i := range ds.X {
			xr.min = math.Min(xr.min, ds.X[i])
			xr.max = math.Max(xr.max, ds.X[i])
			yr.min = math.Min(yr.min, ds.Y[i])
			yr.max = math.Max(yr.max, ds.Y[i])
		}
	}
	widen := func(r axisRange) axisRange {
		return axisRange{math.Floor(r.min) - 1, math.Ceil(r.max) + 1}
	}
	return widen(xr), widen(yr)
}

// drawPanel draws one data set: frame, tick labels, points and fitted line.
func drawPanel(b *strings.Builder, index int, ds DataSet, fit Result, xr, yr axisRange, offsetX, offsetY int) {
	innerW := float64(panelWidth - plotLeft - plotRight)
	innerH := float64(panelHeight - plotTop - plotBottom)
	left := float64(offsetX + plotLeft)
	top := float64(offsetY + plotTop)

	// Convert data coordinates to pixel coordinates.
	px := func(x float64) float64 { return left + (x-xr.min)/(xr.max-xr.min)*innerW }
	py := func(y float64) float64 { return top + innerH - (y-yr.min)/(yr.max-yr.min)*innerH }

	fmt.Fprintf(b, `<text x="%.1f" y="%d" text-anchor="middle" font-size="13" font-weight="bold">%s</text>`+"\n",
		left+innerW/2, offsetY+20, ds.Name)
	fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="#888"/>`+"\n",
		left, top, innerW, innerH)

	for _, t := range ticks(xr) {
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="middle" fill="#444">%g</text>`+"\n", px(t), top+innerH+15, t)
	}
	for _, t := range ticks(yr) {
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="end" fill="#444">%g</text>`+"\n", left-6, py(t)+4, t)
	}
	fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="middle" fill="#444">x</text>`+"\n", left+innerW/2, top+innerH+33)
	fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="middle" fill="#444">y</text>`+"\n", left-35, top+innerH/2)

	// Fitted line across the full x range, clipped to the plot area, then
	// the points on top.
	fmt.Fprintf(b, `<clipPath id="panel%d"><rect x="%.1f" y="%.1f" width="%.1f" height="%.1f"/></clipPath>`+"\n",
		index, left, top, innerW, innerH)
	fmt.Fprintf(b, `<line clip-path="url(#panel%d)" x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#c0392b" stroke-width="1.5"/>`+"\n",
		index, px(xr.min), py(fit.Intercept+fit.Slope*xr.min), px(xr.max), py(fit.Intercept+fit.Slope*xr.max))
	for i := range ds.X {
		fmt.Fprintf(b, `<circle cx="%.1f" cy="%.1f" r="4" fill="#1f3a93"/>`+"\n", px(ds.X[i]), py(ds.Y[i]))
	}

	fmt.Fprintf(b, `<text x="%.1f" y="%.1f" fill="#c0392b">y = %.2f + %.3fx, R² = %.2f</text>`+"\n",
		left+6, top+14, fit.Intercept, fit.Slope, fit.RSquared)
}

// ticks returns about five evenly spaced whole-number tick positions.
func ticks(r axisRange) []float64 {
	step := math.Max(1, math.Ceil((r.max-r.min)/5))
	var out []float64
	for t := math.Ceil(r.min/step) * step; t <= r.max; t += step {
		out = append(out, t)
	}
	return out
}
