# Common tasks. On Windows without make, run the go commands directly (see README).

.PHONY: test bench run plots build app compare programs packages miller profile clean

test:
	go vet ./...
	go test -v -cover ./...

bench:
	go test -bench=. -benchmem -run='^$$' ./anscombe

run:
	go run ./cmd/anscombe

# Draw the quartet and the duet as SVG scatter plots with fitted lines.
plots:
	go run ./cmd/anscombe -reps 0 -mc 0 -svg results/anscombe_quartet.svg
	go run ./cmd/anscombe -reps 0 -mc 0 -data duet -svg results/duet.svg

# Build stand-alone executables for Windows, macOS (Apple silicon and Intel) and Linux.
build:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/anscombe-windows-amd64.exe ./cmd/anscombe
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/anscombe-darwin-arm64 ./cmd/anscombe
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/anscombe-darwin-amd64 ./cmd/anscombe
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/anscombe-linux-amd64 ./cmd/anscombe

# Package the macOS builds as a double-clickable Anscombe.app, zipped so
# the executable permissions survive downloading.
app: build
	rm -rf bin/Anscombe.app bin/Anscombe-macOS.zip
	cp -R macos/Anscombe.app bin/
	mkdir -p bin/Anscombe.app/Contents/Resources
	cp bin/anscombe-darwin-arm64 bin/anscombe-darwin-amd64 bin/Anscombe.app/Contents/Resources/
	chmod +x bin/Anscombe.app/Contents/MacOS/Anscombe bin/Anscombe.app/Contents/Resources/*
	cd bin && zip -qry Anscombe-macOS.zip Anscombe.app && rm -rf Anscombe.app

# Time the regression code and the Monte Carlo benchmark in all three languages.
compare:
	go run ./cmd/anscombe -reps 100000 -mc 100000
	python3 python/anscombe_benchmark.py 1000 100000
	Rscript r/anscombe_benchmark.R 10000 100000

# Time Miller's complete Python and R programs against the Go program.
programs: build
	python3 benchmark/compare_programs.py 20

# Compare six Python regression approaches (statsmodels, Patsy, NumPy, SciPy, scikit-learn).
packages:
	python3 python/package_comparison.py 2000

# Run Miller's (2015) original chapter 1 programs, unchanged (plots included).
# Note: the Python original's savefig() call fails on matplotlib 3.8 and later.
miller:
	cd python && python3 miller-mtpa-chapter-1-program.py
	cd r && Rscript miller-mtpa-chapter-1-program.R

# Write a log plus CPU and memory profiles, then show the top CPU users.
profile:
	go run ./cmd/anscombe -reps 100000 -mc 100000 -log run.log -cpuprofile cpu.prof -memprofile mem.prof
	go tool pprof -top -nodecount=10 cpu.prof

clean:
	rm -rf bin/Anscombe.app run.log cpu.prof mem.prof python/fig_anscombe_Python.pdf r/fig_anscombe_R.pdf
