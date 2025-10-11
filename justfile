set windows-shell := ["pwsh", "-c"]

# List targets
list:
    just -l 

# Binary paths
parc_bin := if os() == "windows" { "bin/parc.exe" } else { "bin/parc" }

# Build parc binary
parc:
    go build -o bin/ ./parc/

# Generate documentation (requires parc to be built first)
docs: parc
    {{parc_bin}} gendocs --path ./site/data/

# Run tests
test:
    go test ./ponzu/...

# Clean build artifacts
[unix]
clean:
    rm -rf bin/

[windows]
clean:
    Remove-Item -Recurse bin

# Build everything
all: parc docs
