test:
	go test ./...

dep:
	go mod download

vet:
	go vet ./...

fmt:
	go fmt ./...

run: test fmt
	go run ./eventloop/