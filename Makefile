.PHONY: parc parcdocs all test clean spewstat

ifeq ($(OS),Windows_NT)
clean:
	rmdir /S/Q bin
else
clean:
	rm -rf bin/
endif


parc:
	go build -o bin/ ./parc/
spewstat:
	go build -o bin spewstat.go
parcdocs:
	go run ./parc/gendocs.go
test:
	go test ./ponzu/...
all: parc parcdocs