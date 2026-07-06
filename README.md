# kit

[![Go Version](https://img.shields.io/github/go-mod/go-version/onnttf/kit)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/onnttf/kit.svg)](https://pkg.go.dev/github.com/onnttf/kit)

`kit` is a small collection of Go utility packages for application code. It
focuses on practical helpers for task execution, Excel files, HTTP downloads,
DingTalk robots, trees, slices, pointers, and calendar ranges.


## Getting Started

```bash
go get github.com/onnttf/kit
```

Import only the package you need:

```go
import "github.com/onnttf/kit/tree"
```

## Packages

| Package    | Purpose                                                                                  |
| ---------- | ---------------------------------------------------------------------------------------- |
| `dingtalk` | Build and send DingTalk robot messages with explicit request context.                    |
| `download` | Download HTTP resources as files or byte slices with clients, size limits, and headers.  |
| `excel`    | Read Excel workbooks, stream sheet rows, and decode rows into structs with `excel` tags. |
| `fsm`      | Run a small concurrency-safe finite state machine.                                       |
| `ptr`      | Create and dereference pointers safely.                                                  |
| `slicex`   | Slice helpers that add behavior beyond the standard library.                             |
| `task`     | Run in-memory tasks concurrently and collect ordered item results.                       |
| `timex`    | Compute half-open calendar ranges and start/end boundaries.                              |
| `tree`     | Build, edit, validate, query, and snapshot typed trees.                                  |

## Contributing

### Prerequisites

- Go 1.25 or newer
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
