.PHONY: all test lint fmt vet clean cyclo cyclo-all cognitive cognitive-high

# Default target - runs lint and tests
all: lint test

# Run all tests
test:
	@echo "Running tests..."
	go test -v -race ./...

# Run linter (golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f coverage.out coverage.html
	go clean

# Install development tools
tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/uudashr/gocognit/cmd/gocognit@latest

# Generate mocks
mocks:
	@echo "Generating mocks..."
	go generate ./...

# Cyclomatic complexity analysis - top 15 offenders (complexity > 10 is considered high)
cyclo:
	@echo "Analyzing cyclomatic complexity (top 15 offenders)..."
	@gocyclo -top 15 -ignore "_test|mocks" .

# Cyclomatic complexity analysis - all functions above threshold
cyclo-all:
	@echo "Functions with cyclomatic complexity > 10:"
	@gocyclo -over 10 -ignore "_test|mocks" . || echo "No functions with complexity > 10 found"

# Cognitive complexity analysis (alternative metric)
cognitive:
	@echo "Analyzing cognitive complexity (top 15 offenders)..."
	@gocognit -top 15 -ignore "_test|mocks" .

# Cognitive complexity - only high complexity (> 15)
cognitive-high:
	@echo "Functions with HIGH cognitive complexity (> 15):"
	@gocognit -over 15 -ignore "_test|mocks" . || echo "No functions with complexity > 15 found"
