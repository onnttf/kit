# kit

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/onnttf/kit.svg)](https://pkg.go.dev/github.com/onnttf/kit)

A lightweight, modular Go toolkit for building production-ready applications.

## Features

- **Concurrent Execution** - Task executor with retry, backoff, and error handling
- **Database Abstraction** - Generic CRUD repository with GORM integration
- **Container Utilities** - Slice operations (intersection, union, difference, deduplicate)
- **Excel Processing** - Read and stream Excel files with ease
- **DingTalk Integration** - Send messages to DingTalk robots
- **Tree Structures** - Build and manipulate hierarchical data
- **Pointer Helpers** - Safe pointer operations and utilities
- **Time Utilities** - Day boundary calculations

## Installation

```bash
go get github.com/onnttf/kit
```

## Packages

| Package | Description |
| --------- | ------------- |
| [concurrent](concurrent/) | Concurrent task execution with retry and error handling |
| [container](container/) | Slice operations and utilities |
| [dal](dal/) | Generic database repository with GORM |
| [dingtalk](dingtalk/) | DingTalk robot message sender |
| [excel](excel/) | Excel file reading and streaming |
| [ptr](ptr/) | Pointer helper functions |
| [time](time/) | Time calculation utilities |
| [tree](tree/) | Tree structure builder and operations |

## Documentation

- [API Reference](https://pkg.go.dev/github.com/onnttf/kit) - Complete package documentation

## Contributing

Contributions are welcome! Please follow these guidelines:

**Before contributing:**

1. Ensure all tests pass: `go test ./...`
2. Format your code: `goimports -w .`

**Contribution process:**

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -am 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests with race detection
go test ./... -race
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by Go kit and other open-source toolkits
- Thanks to all contributors who help improve this project!
