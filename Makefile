# Common tasks. On Windows without make, run the go commands directly.

.PHONY: test bench run build compare clean

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

# Run all three languages one after another for a timing comparison.
compare:
	go run ./cmd/anscombe -reps 100000
	python3 python/anscombe_regression.py 1000
	Rscript r/anscombe_regression.R 10000

clean:
	rm -rf bin
