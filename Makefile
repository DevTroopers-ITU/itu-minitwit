init:
	sqlite3 /tmp/minitwit.db < schema.sql

build:
	CGO_ENABLED=1 go build -o minitwit .

test:
	CGO_ENABLED=1 go test -v ./...

clean:
	rm -f minitwit
