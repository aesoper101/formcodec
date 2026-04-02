# formcodec

[中文文档](README_zh.md)

A Go library for bidirectional conversion between Go structs and `map[string][]string`. Ideal for handling HTTP form data, query parameters, and similar use cases.

## Features

- All Go basic types supported (string, bool, int/int8/.../int64, uint/uint8/.../uint64, float32/float64)
- Pointer types for all basic types
- Slice types for all basic types
- Built-in `time.Time` / `*time.Time` support (RFC3339 format)
- Built-in `[]byte` / `*[]byte` support (treated as string)
- Custom type conversion via `Marshaler` / `Unmarshaler` interfaces (prioritized over built-in conversion)
- Default values support (single and multi-value with `|` separator)
- Configurable struct tag name (default: `form`)
- Both package-level functions and Encoder/Decoder types

## Installation

```bash
go get github.com/aesoper101/formcodec
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/aesoper101/formcodec"
)

type User struct {
    Name  string   `form:"name"`
    Age   int      `form:"age"`
    Tags  []string `form:"tags"`
}

func main() {
    // Encode struct to map
    user := User{Name: "Alice", Age: 30, Tags: []string{"admin", "user"}}
    values, err := formcodec.Marshal(user)
    if err != nil {
        panic(err)
    }
    fmt.Println(values)
    // Output: map[string][]string{"name": {"Alice"}, "age": {"30"}, "tags": {"admin", "user"}}

    // Decode map to struct
    var decoded User
    err = formcodec.Unmarshal(values, &decoded)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", decoded)
    // Output: {Name:Alice Age:30 Tags:[admin user]}
}
```

## API Reference

### Package-Level Functions

```go
// Marshal encodes a struct to map[string][]string
func Marshal(v any) (map[string][]string, error)

// Unmarshal decodes map[string][]string to a struct
func Unmarshal(data map[string][]string, v any) error
```

### Encoder / Decoder Types

```go
// Create encoder with options
func NewEncoder(opts ...EncoderOption) *Encoder
func (e *Encoder) Encode(v any) (map[string][]string, error)

// Create decoder with options
func NewDecoder(opts ...DecoderOption) *Decoder
func (d *Decoder) Decode(data map[string][]string, v any) error
```

### Options

```go
// Customize the struct tag name for encoding (default: "form")
func WithEncoderTagName(name string) EncoderOption

// Customize the struct tag name for decoding (default: "form")
func WithDecoderTagName(name string) DecoderOption
```

### Interfaces

Implement these interfaces to customize how your types are converted:

```go
// Marshaler is implemented by types that can marshal themselves to []string
type Marshaler interface {
    MarshalValues() ([]string, error)
}

// Unmarshaler is implemented by types that can unmarshal []string to themselves
type Unmarshaler interface {
    UnmarshalValues(values []string) error
}
```

### Errors

| Error | Description |
|-------|-------------|
| `ErrInvalidValue` | Input value is nil |
| `ErrNilPointer` | Decode target is a nil pointer |
| `ErrNotPointer` | Decode target is not a pointer |
| `ErrNotStruct` | Target is not a struct |

## Struct Tag Syntax

```
`form:"field_name,option1,option2,..."`
```

| Option | Syntax | Description |
|--------|--------|-------------|
| Field name | `form:"custom_name"` | Custom key name in map |
| Skip field | `form:"-"` | Field is ignored during encode/decode |
| Omit empty | `form:"name,omitempty"` | Skip zero values during encoding |
| Default value | `form:"name,default=value"` | Use default when key is missing |
| Multi-value default | `form:"name,default=a\|b\|c"` | Multiple defaults separated by `\|` |

> **Note:** The `omitempty` and `default` options can appear in any order.

## Supported Types

| Category | Types |
|----------|-------|
| Basic types | `string`, `bool`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64` |
| Pointer types | `*string`, `*bool`, `*int`, `*int8`, ..., `*float64` |
| Slice types | `[]string`, `[]bool`, `[]int`, `[]int8`, ..., `[]float64` |
| Time | `time.Time`, `*time.Time` (RFC3339 format) |
| Bytes | `[]byte`, `*[]byte` (treated as string) |

## Custom Types

Implement `Marshaler` and `Unmarshaler` interfaces for custom type conversion:

```go
type StringList struct {
    Items []string
}

// MarshalValues implements formcodec.Marshaler
func (s StringList) MarshalValues() ([]string, error) {
    return s.Items, nil
}

// UnmarshalValues implements formcodec.Unmarshaler
func (s *StringList) UnmarshalValues(values []string) error {
    s.Items = values
    return nil
}

// Usage
type Form struct {
    List StringList `form:"list"`
}
```

> Custom `Marshaler`/`Unmarshaler` implementations take priority over built-in type conversion.

## Default Values

Default values are used in both encoding and decoding:

- **Decoding**: When a key is missing in the input map, the default value is used
- **Encoding**: With `omitempty`, values equal to the default are omitted from output

```go
type Config struct {
    Host  string   `form:"host,default=localhost"`
    Ports []int    `form:"ports,default=8080|8081"`  // Multi-value default
    Debug bool     `form:"debug,omitempty,default=false"`
}
```

> Multi-value defaults use `|` as the separator.

## Options

Customize the struct tag name:

```go
// Using custom tag name "query" instead of default "form"
type Request struct {
    Page  int `query:"page"`
    Limit int `query:"limit"`
}

// Encoder with custom tag
encoder := formcodec.NewEncoder(formcodec.WithEncoderTagName("query"))
values, err := encoder.Encode(Request{Page: 1, Limit: 10})

// Decoder with custom tag
decoder := formcodec.NewDecoder(formcodec.WithDecoderTagName("query"))
var req Request
err = decoder.Decode(values, &req)
```

## License

Apache-2.0
