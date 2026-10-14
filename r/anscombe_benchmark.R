# Timing and Monte Carlo benchmark for the Anscombe Quartet in R.
#
# Times the regression code from Miller (2015), Exhibit 1.2
# (summary(lm(y1 ~ x1, data = anscombe))), plus the leaner .lm.fit(), so R
# can be compared with the Go and Python programs. It then runs the same
# Monte Carlo benchmark as the Go program: draw many synthetic samples from
# y = 3 + 0.5x + N(0, 1.24^2), fit each, and report the mean slope, 95%
# confidence-interval coverage and time per run.
#
# Usage:  Rscript r/anscombe_benchmark.R [reps] [monte_carlo_runs] [n]

# Read a whole-number argument, explaining any bad value.
read_count <- function(args, position, default, minimum) {
  if (length(args) < position) return(default)
  value <- suppressWarnings(as.integer(args[position]))
  if (is.na(value) || value < minimum) {
    stop(sprintf("argument %d must be a whole number >= %d; got '%s'",
                 position, minimum, args[position]), call. = FALSE)
  }
  value
}

args <- commandArgs(trailingOnly = TRUE)
reps <- read_count(args, 1, 10000L, 0L)
mc_runs <- read_count(args, 2, 10000L, 0L)
mc_n <- read_count(args, 3, 100L, 3L)

# The anscombe data frame exactly as defined in Miller (2015), Exhibit 1.2.
anscombe <- data.frame(
    x1 = c(10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5),
    x2 = c(10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5),
    x3 = c(10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5),
    x4 = c(8, 8, 8, 8, 8, 8, 8, 19, 8, 8, 8),
    y1 = c(8.04, 6.95,  7.58, 8.81, 8.33, 9.96, 7.24, 4.26,10.84, 4.82, 5.68),
    y2 = c(9.14, 8.14,  8.74, 8.77, 9.26, 8.1, 6.13, 3.1,  9.13, 7.26, 4.74),
    y3 = c(7.46, 6.77, 12.74, 7.11, 7.81, 8.84, 6.08, 5.39, 8.15, 6.42, 5.73),
    y4 = c(6.58, 5.76,  7.71, 8.84, 8.47, 7.04, 5.25, 12.5, 5.56, 7.91, 6.89))

set_names <- c("Set I", "Set II", "Set III", "Set IV")

# Miller's model formulas, one per set: y1 ~ x1, ..., y4 ~ x4.
formulas <- lapply(1:4, function(i) as.formula(sprintf("y%d ~ x%d", i, i)))

cat(sprintf("%-8s %12s %12s %12s %12s %10s\n",
            "Set", "Intercept", "Slope", "Int SE", "Slope SE", "R^2"))
for (i in 1:4) {
  fit <- summary(lm(formulas[[i]], data = anscombe))
  co <- coef(fit)
  cat(sprintf("%-8s %12.6f %12.6f %12.6f %12.6f %10.6f\n",
              set_names[i], co[1, 1], co[2, 1], co[1, 2], co[2, 2], fit$r.squared))
}

invisible(gc(reset = TRUE))

# Timing: Miller's summary(lm()) on all four sets, `reps` times.
if (reps > 0) {
  elapsed <- system.time({
    for (r in seq_len(reps)) {
      for (f in formulas) {
        fit <- summary(lm(f, data = anscombe))
      }
    }
  })["elapsed"]
  cat(sprintf("\nR lm() (Miller's code, with summary()): %d reps x 4 sets in %.3f s (%.2f microseconds per fit)\n",
              reps, elapsed, elapsed / (reps * 4) * 1e6))

  # Lean alternative: .lm.fit on a design matrix (no formula or summary).
  mats <- lapply(1:4, function(i) list(X = cbind(1, anscombe[[paste0("x", i)]]),
                                       y = anscombe[[paste0("y", i)]]))
  lean_reps <- reps * 10L
  elapsed <- system.time({
    for (r in seq_len(lean_reps)) {
      for (m in mats) {
        fit <- .lm.fit(m$X, m$y)
      }
    }
  })["elapsed"]
  cat(sprintf("R .lm.fit(): %d reps x 4 sets in %.3f s (%.2f microseconds per fit)\n",
              lean_reps, elapsed, elapsed / (lean_reps * 4) * 1e6))
}

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
cat(sprintf("Memory: peak R heap %.1f MB (gc 'max used', Ncells + Vcells)\n",
            sum(mem[, ncol(mem)])))
