.PHONY: build test clean deps

# Build the application
build:
	go build -o bin/steganography cmd/main.go

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Download dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
install:
	go install

# Run the application (example)
run-embed:
	./bin/steganography embed -c test/cover-1.mp3 -m test/secret.txt -k mykey -l 2 -r -o test/stego.mp3 -e true

run-extract:
	./bin/steganography extract -s test/stego.mp3 -k mykey -o test/extracted.pdf -d true
