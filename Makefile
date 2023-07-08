.PHONY: parc parcdocs all test clean spewstat

ifeq ($(OS),Windows_NT)
clean:
	rmdir /S/Q bin
	rmdir /S/Q docs/parc
else
clean:
	rm -rf bin/
	rm -rf docs/parc/
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