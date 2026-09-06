clean:
	rm -rf bin dist .goreleaser-dist
	go clean ./...

test:
	go test ./... -v -cover

lint:
	go tool golangci-lint run --fix ./...

format:
	gofmt -w .

typecheck:
	go build ./...

prepare: lint format typecheck test
