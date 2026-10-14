# Anscombe Quartet regressions in R, adapted from Miller (2015), chapter 1.
# Prints coefficients for each data set, then times repeated fits (lm() with
# summary, and the leaner .lm.fit()) so R can be compared with Go and Python.
#
# Usage:  Rscript r/anscombe_regression.R [reps]

args <- commandArgs(trailingOnly = TRUE)
reps <- if (length(args) >= 1) as.integer(args[1]) else 10000L

# The anscombe data frame ships with base R (datasets package).
sets <- list(
  "Set I"   = data.frame(x = anscombe$x1, y = anscombe$y1),
  "Set II"  = data.frame(x = anscombe$x2, y = anscombe$y2),
  "Set III" = data.frame(x = anscombe$x3, y = anscombe$y3),
  "Set IV"  = data.frame(x = anscombe$x4, y = anscombe$y4)
)

cat(sprintf("%-8s %12s %12s %12s %12s %10s\n",
            "Set", "Intercept", "Slope", "Int SE", "Slope SE", "R^2"))
for (name in names(sets)) {
  fit <- summary(lm(y ~ x, data = sets[[name]]))
  co <- coef(fit)
  cat(sprintf("%-8s %12.6f %12.6f %12.6f %12.6f %10.6f\n",
              name, co[1, 1], co[2, 1], co[1, 2], co[2, 2], fit$r.squared))
}

invisible(gc(reset = TRUE))

# Timing: fit all four sets `reps` times, as the Go and Python programs do.
elapsed <- system.time({
  for (i in seq_len(reps)) {
    for (d in sets) {
      fit <- summary(lm(y ~ x, data = d))
    }
  }
})["elapsed"]

cat(sprintf("\nR lm(): %d reps x 4 sets in %.3f s (%.2f microseconds per fit)\n",
            reps, elapsed, elapsed / (reps * 4) * 1e6))

# Lean alternative: .lm.fit on a design matrix (no formula or summary),
# closer to what the Go program does.
mats <- lapply(sets, function(d) list(X = cbind(1, d$x), y = d$y))
lean_reps <- reps * 10L
elapsed <- system.time({
  for (i in seq_len(lean_reps)) {
    for (m in mats) {
      fit <- .lm.fit(m$X, m$y)
    }
  }
})["elapsed"]

cat(sprintf("R .lm.fit(): %d reps x 4 sets in %.3f s (%.2f microseconds per fit)\n",
            lean_reps, elapsed, elapsed / (lean_reps * 4) * 1e6))

mem <- gc()
cat(sprintf("Memory: peak R heap during timing %.1f MB (gc 'max used', Ncells + Vcells)\n",
            sum(mem[, ncol(mem)])))
