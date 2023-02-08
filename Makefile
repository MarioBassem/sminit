build:
	go build -o sminit ./cmd/main.go

release:
	go build -ldflags="-s -w" -o sminit ./cmd/main.go

test:
	go test ./...

