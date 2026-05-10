# kit

[![Go Version](https://img.shields.io/github/go-mod/go-version/onnttf/kit)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/onnttf/kit.svg)](https://pkg.go.dev/github.com/onnttf/kit)

`kit` is a small collection of Go utility packages for application code. It
focuses on practical helpers for concurrency, database access, Excel files,
HTTP downloads, DingTalk robots, trees, slices, pointers, and calendar ranges.


## Getting Started

```bash
go get github.com/onnttf/kit
```

Import only the package you need:

```go
import "github.com/onnttf/kit/tree"
```

## Packages

| Package      | Purpose                                                                                       |
| ------------ | --------------------------------------------------------------------------------------------- |
| `concurrent` | Run bounded concurrent work with retry, backoff, timeout, panic policy, and result summaries. |
| `container`  | Generic slice helpers such as difference, intersection, union, grouping, and partitioning.    |
| `dal`        | Generic GORM repository operations and reusable query scopes.                                 |
| `dingtalk`   | Build and send DingTalk robot messages.                                                       |
| `download`   | Download HTTP resources as files or byte slices with size limits and atomic file writes.      |
| `excel`      | Read Excel workbooks and parse rows into structs with `excel` tags.                           |
| `ptr`        | Create and dereference pointers safely.                                                       |
| `time`       | Compute day, week, month, and year boundaries.                                                |
| `tree`       | Build, validate, query, transform, filter, and flatten typed trees.                           |

## Contributing

### Prerequisites

- Go 1.23 or newer
- `golangci-lint` for lint checks

### Setup

```bash
git clone https://github.com/onnttf/kit.git
cd kit
go mod download
```

### Checks

Run these before opening a pull request:

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
golangci-lint run
```

Use a writable cache directory when your environment restricts the default Go
cache:

```bash
GOCACHE=/private/tmp/go-cache go test ./...
```

### Pull Requests

- Keep changes focused.
- Prefer standard library APIs over low-value wrappers.
- Add or update tests for observable behavior, important edge cases, and real error paths.
- Avoid tests that only exercise getters, constants, or implementation details.
- Document exported APIs when adding or changing public symbols.

## Contributors

Contributions are welcome through issues and pull requests.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
