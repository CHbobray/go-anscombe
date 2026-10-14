# Miller (2015), Exhibit 1.2: Programming the Anscombe Quartet (R).
# Reproduced as provided for the course, except that the plotting section
# is removed, as the assignment allows. This is the program timed as a whole
# by benchmark/compare_programs.py. The Go version of the plots is produced
# with: anscombe -svg plot.svg

# The Anscombe Quartet (R)

# demonstration data from
# Anscombe, F. J. 1973, February. Graphs in statistical analysis. 
#  The American Statistician 27: 17–21.

# define the anscombe data frame
anscombe <- data.frame(
    x1 = c(10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5),
    x2 = c(10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5),
    x3 = c(10, 8, 13, 9, 11, 14, 6, 4, 12, 7, 5),
    x4 = c(8, 8, 8, 8, 8, 8, 8, 19, 8, 8, 8),
    y1 = c(8.04, 6.95,  7.58, 8.81, 8.33, 9.96, 7.24, 4.26,10.84, 4.82, 5.68),
    y2 = c(9.14, 8.14,  8.74, 8.77, 9.26, 8.1, 6.13, 3.1,  9.13, 7.26, 4.74),
    y3 = c(7.46, 6.77, 12.74, 7.11, 7.81, 8.84, 6.08, 5.39, 8.15, 6.42, 5.73),
    y4 = c(6.58, 5.76,  7.71, 8.84, 8.47, 7.04, 5.25, 12.5, 5.56, 7.91, 6.89))

# show results from four regression analyses
with(anscombe, print(summary(lm(y1 ~ x1, data = anscombe))))
with(anscombe, print(summary(lm(y2 ~ x2, data = anscombe))))
with(anscombe, print(summary(lm(y3 ~ x3, data = anscombe))))
with(anscombe, print(summary(lm(y4 ~ x4, data = anscombe))))

# [plotting section removed; see the note at the top of this file]

# Suggestions for the student:
# See if you can develop a quartet of your own, 
# or perhaps just a duet, two very different data sets 
# with the same fitted model.