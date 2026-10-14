# Common tasks. On Windows without make, run the go commands directly (see README).

.PHONY: test bench run build app compare profile clean

test:
	go vet ./...
	go test -v -cover ./...

bench:
	go test -bench=. -benchmem -run='^$$' ./anscombe

run:
	go run ./cmd/anscombe

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

# Run all three languages one after another for a timing comparison.
compare:
	go run ./cmd/anscombe -reps 100000 -mc 100000
	python3 python/anscombe_regression.py 1000 100000
	Rscript r/anscombe_regression.R 10000 100000

# Write a log plus CPU and memory profiles, then show the top CPU users.
profile:
	go run ./cmd/anscombe -reps 100000 -mc 100000 -log run.log -cpuprofile cpu.prof -memprofile mem.prof
	go tool pprof -top -nodecount=10 cpu.prof

clean:
	rm -rf bin/Anscombe.app run.log cpu.prof mem.prof
