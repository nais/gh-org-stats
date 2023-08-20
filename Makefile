.PHONY: build test run

build:
		go build -o build/gh-org-stats main.go

test:
		go test ./...

run:
		go run main.go