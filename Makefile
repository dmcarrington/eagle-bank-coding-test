build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

docker-build:
	docker build -t eagle-bank .

.PHONY: build test vet lint docker-build
