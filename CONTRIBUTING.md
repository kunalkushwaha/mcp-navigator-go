# Contributing to MCP Navigator

Thank you for your interest in contributing to MCP Navigator! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Style Guidelines](#style-guidelines)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/mcp-navigator-go.git
   cd mcp-navigator-go
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/kunalkushwaha/mcp-navigator-go.git
   ```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- Docker (for testing Docker-based MCP servers)

### Setup

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Build the project:
   ```bash
   go build -o mcp-navigator main.go
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

## Making Changes

1. Create a new branch for your feature or fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes, following the [style guidelines](#style-guidelines)

3. Add or update tests as necessary

4. Run the full test suite:
   ```bash
   go test -v -race ./...
   ```

5. Run code formatting:
   ```bash
   go fmt ./...
   ```

6. Run static analysis:
   ```bash
   go vet ./...
   ```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./pkg/client
```

### Test Structure

- Unit tests should be in the same package as the code they test
- Integration tests should be in the `tests/` directory
- Use table-driven tests where appropriate
- Mock external dependencies

### Adding Tests

When adding new functionality:

1. Write unit tests for individual functions/methods
2. Add integration tests for end-to-end workflows
3. Update existing tests if you modify behavior
4. Ensure test coverage remains high

## Submitting Changes

1. Commit your changes with a clear commit message:
   ```bash
   git commit -m "feat: add support for HTTP transport

   - Implement HTTP transport for web applications
   - Add configuration options for HTTP endpoints
   - Update documentation with HTTP examples"
   ```

2. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

3. Create a pull request on GitHub with:
   - Clear title and description
   - Reference to any related issues
   - Screenshots/examples if applicable
   - Test results

## Style Guidelines

### Go Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small

### Commit Messages

Use conventional commit format:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test additions/modifications
- `refactor:` for code refactoring
- `style:` for formatting changes
- `chore:` for maintenance tasks

### Documentation

- Update README.md for user-facing changes
- Add/update code comments for complex logic
- Update examples if APIs change
- Keep documentation current with code changes

## Project Structure

```
mcp-navigator-go/
├── cmd/                    # Command-line interfaces
├── pkg/                    # Library packages
│   ├── client/            # MCP client implementation
│   ├── transport/         # Transport implementations
│   ├── discovery/         # Server discovery
│   └── mcp/              # MCP protocol types
├── internal/              # Internal packages
│   └── cli/              # CLI implementation
├── examples/              # Usage examples
├── tests/                 # Integration tests
└── docs/                 # Additional documentation
```

## Questions?

If you have questions about contributing, please:

1. Check existing issues and discussions
2. Open a new issue with the "question" label
3. Reach out to the maintainers

Thank you for contributing to MCP Navigator!
