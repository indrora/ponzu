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

ifeq ($(OS),Windows_NT)
docs: parc 
	./bin/parc.exe gendocs --path ./site/data/
else
clean: parc
	./bin/parc gendocs --path ./site/data/
endif


test:
	go test ./ponzu/...
all: parc docs