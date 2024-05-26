# SVault-Engine

SVault-Engine is the core engine for [SVault](https://github.com/Owbird/SVault). It's a vault-based Virtual File System offering secure file encryption, cross-device sharing, and file server hosting.

## Installation

To install SVault-Engine, use the `go install` command or download the latest [release](https://github.com/Owbird/SVault-Engine/releases):

```bash
go install github.com/Owbird/SVault-Engine@latest
```

## Usage

### Command Line Interface (CLI)

Run the main program:

```bash
SVault-Engine [command]
```

### Go Package

To use SVault-Engine as a package in your Go application, import it and utilize its features:

```go
import "github.com/Owbird/SVault-Engine/pkg/vault"


func main() {
    vault := vault.NewVault()
}
```

For detailed documentation, visit the [Go package documentation](https://pkg.go.dev/github.com/Owbird/SVault-Engine).

## Contributing

We welcome contributions!

## License

This project is licensed under the [MIT License](LICENSE).
