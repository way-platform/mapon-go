# Agent Instructions

## Package Manager

Use **Go Modules**: `go mod tidy`, `go test ./...`

## Structure

- One API operation per file: `client_<collection>_<operation>.go`
- Client returns protobuf messages — see `proto/` for schemas
- Protobuf conversion functions live in `proto.go`

## CLI Architecture

The CLI is split into two layers to keep credential storage pluggable:

```
cli/
├── cli.go       # Store interface, Credentials, FileStore, Options
└── command.go   # NewCommand() — full command tree
cmd/mapon/
└── main.go      # Thin wrapper: wires FileStore to XDG paths
```

- `cli.Store` — interface with `Read(any)`, `Write(any)`, `Clear()` methods
- `cli.NewCommand(...Option)` — builds the Cobra command tree; receives stores via functional options (`WithCredentialStore`)
- `cmd/mapon/main.go` — only wires `FileStore` instances and calls `cli.NewCommand()`

This separation lets consumers embed the CLI in a larger tool or swap the storage backend (e.g. use an in-memory store in tests, or a keychain-backed store) without forking.

### Authentication

The Mapon API uses a single API key for all requests. The key is stored in the credential store.

### Conventions

- Subcommands are organized by entity using `cobra.Group`
- Fully paginate results where applicable
- Flat command structure except `unit-groups` (which has `list` and `units` subcommands)
