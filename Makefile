build:
	go build -o minitwit .

test:
	go test -v ./...

clean:
	rm -f minitwit