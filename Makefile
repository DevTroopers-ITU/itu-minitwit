build:
	go build -o minitwit .

test:
	go test -v ./...

lint:
	gofmt -l . | tee /dev/stderr | (! read)
	golangci-lint run
	hadolint Dockerfile

fmt:
	gofmt -w .
	goimports -w .

clean:
	rm -f minitwit