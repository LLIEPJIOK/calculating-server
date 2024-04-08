run:
	go run -race cmd/main.go

test:
	go test -v -cover ./...

pprof-test:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html