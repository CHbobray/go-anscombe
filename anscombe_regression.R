# Anscombe Quartet regressions in R, adapted from Miller (2015), chapter 1.
# Prints coefficients for each data set, then times repeated fits (lm() with
# summary, and the leaner .lm.fit()) so R can be compared with Go and Python.
#
# It also runs the same Monte Carlo benchmark as the Go program: draw many
# synthetic samples from y = 3 + 0.5x + N(0, 1.24^2), fit each, and report
# the mean slope, 95% confidence-interval coverage and time per run.
#
# Usage:  Rscript r/anscombe_regression.R [reps] [monte_carlo_runs] [n]

args <- commandArgs(trailingOnly = TRUE)
reps <- if (length(args) >= 1) as.integer(args[1]) else 10000L
mc_runs <- if (length(args) >= 2) as.integer(args[2]) else 10000L
mc_n <- if (length(args) >= 3) as.integer(args[3]) else 100L

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

# Monte Carlo benchmark: simulate, fit with .lm.fit, compute the slope SE.
if (mc_runs > 0) {
  set.seed(1)
  t_crit <- qt(0.975, mc_n - 2)
  slopes <- numeric(mc_runs)
  covered <- 0L
  elapsed <- system.time({
    for (i in seq_len(mc_runs)) {
      x <- 10 * runif(mc_n)
      y <- 3 + 0.5 * x + rnorm(mc_n, sd = 1.24)
      fit <- .lm.fit(cbind(1, x), y)
      slope <- fit$coefficients[2]
      slope_se <- sqrt(sum(fit$residuals^2) / (mc_n - 2) / sum((x - mean(x))^2))
      slopes[i] <- slope
      if (abs(slope - 0.5) <= t_crit * slope_se) covered <- covered + 1L
    }
  })["elapsed"]
  cat(sprintf("\nMonte Carlo: %d samples of n = %d from y = 3 + 0.5x + N(0, 1.24^2)\n",
              mc_runs, mc_n))
  cat(sprintf("  mean slope %.5f (true 0.5), SD %.5f, 95%% CI coverage %.1f%%\n",
              mean(slopes), sqrt(mean((slopes - mean(slopes))^2)), 100 * covered / mc_runs))
  cat(sprintf("  R: %.3f s total (%.2f microseconds per simulate + fit)\n",
              elapsed, elapsed / mc_runs * 1e6))
}

mem <- gc()
cat(sprintf("Memory: peak R heap during timing %.1f MB (gc 'max used', Ncells + Vcells)\n",
            sum(mem[, ncol(mem)])))
