# Contributing

Thanks for your interest in contributing to Seigen.

## Getting Started

Clone the repository and install dependencies:

```bash
git clone https://github.com/JhayceeCodes/seigen.git
cd seigen
go mod download
```

Run the test suite:
```bash
go test ./...
```

Run the race detector:
```bash
go test -race ./...
```

Run the static checks:
```bash
go vet ./...
```

## Making Changes
- Keep changes focused and minimal.
- Follow standard Go conventions.
- Add or update tests for behavioral changes.
- Run gofmt on changed Go files.
- Update documentation when changing public APIs or behavior.
- For performance-related changes, update or add benchmarks where appropriate.

## Pull Requests
Please include:
- A clear description of the change.
- Tests for new or changed behavior.
- Any relevant documentation updates.

Keep pull requests focused and avoid unrelated changes.

## Database Changes

Do not modify migrations that have already been applied.

Create a new migration for schema changes instead.

## License

By contributing to Seigen, you agree that your contributions will be licensed under the [MIT License](LICENSE).