.PHONY: parc all test clean

ifeq ($(OS),Windows_NT)
clean:
	rmdir /S/Q bin
else
clean:
	rm -rf bin/
endif


parc:
	go build -o bin/ ./parc/
test:
	go test ./ponzu/...
all: parc