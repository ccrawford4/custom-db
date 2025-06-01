# Custom DB

A simple B+tree-based key-value database implementation in Go.

## Features

- B+tree data structure for efficient key-value storage
- Get, Set, and Delete operations
- Range queries with iterator support
- Page-based storage with 4KB pages
- Support for variable-length keys and values

## Usage

The database provides a simple KV interface:

```go
type KV interface {
    Get(key []byte) (val []byte, ok bool)
    Set(key []byte, val []byte)
    Del(key []byte)
    FindGreaterThan(key []byte) Iterator
}
```

## Building

```bash
go build ./cmd
```

## Testing

```bash
go test
```

## Implementation Details

- Uses B+tree nodes with internal and leaf node types
- Fixed page size of 4KB
- Maximum key size: 1000 bytes
- Maximum value size: 3000 bytes
- Immutable node design for data integrity

