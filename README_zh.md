[English](README.md)

# formcodec

一个 Go 语言库，用于 Go 结构体与 `map[string][]string` 之间的双向转换。

常用于 HTTP 表单数据、查询参数等场景的编码与解码。

## 特性

- 支持所有 Go 基础类型（string、bool、int/int8/.../int64、uint/uint8/.../uint64、float32/float64）
- 支持所有基础类型的指针类型
- 支持所有基础类型的切片类型
- 内置 `time.Time` / `*time.Time` 支持（RFC3339 格式）
- 内置 `[]byte` / `*[]byte` 支持（作为字符串处理）
- 通过 `Marshaler` / `Unmarshaler` 接口实现自定义类型转换（优先于内置转换）
- 支持默认值（单值和多值，多值使用 `|` 分隔）
- 可配置 struct tag 名称（默认：`form`）
- 同时提供包级函数和 Encoder/Decoder 类型两套 API

## 安装

```bash
go get github.com/aesoper101/formcodec
```

## 快速开始

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
    // 编码
    user := User{Name: "Alice", Age: 30, Tags: []string{"admin", "user"}}
    values, err := formcodec.Marshal(user)
    if err != nil {
        panic(err)
    }
    fmt.Println(values)
    // 输出: map[string][]string{"name": {"Alice"}, "age": {"30"}, "tags": {"admin", "user"}}

    // 解码
    var decoded User
    err = formcodec.Unmarshal(values, &decoded)
    if err != nil {
        panic(err)
    }
    fmt.Println(decoded)
    // 输出: User{Name: "Alice", Age: 30, Tags: ["admin", "user"]}
}
```

## API 参考

### 包级函数

- `Marshal(v any) (map[string][]string, error)` — 将结构体编码为 map
- `Unmarshal(data map[string][]string, v any) error` — 将 map 解码为结构体

### Encoder / Decoder 类型

```go
// 创建编码器
func NewEncoder(opts ...EncoderOption) *Encoder

// 编码结构体
func (e *Encoder) Encode(v any) (map[string][]string, error)

// 创建解码器
func NewDecoder(opts ...DecoderOption) *Decoder

// 解码到结构体
func (d *Decoder) Decode(data map[string][]string, v any) error
```

### 选项

- `WithEncoderTagName(name string) EncoderOption` — 设置编码器的 tag 名称
- `WithDecoderTagName(name string) DecoderOption` — 设置解码器的 tag 名称

### 接口

```go
// 自定义编码接口
type Marshaler interface {
    MarshalValues() ([]string, error)
}

// 自定义解码接口
type Unmarshaler interface {
    UnmarshalValues(values []string) error
}
```

### 错误

| 错误 | 说明 |
|------|------|
| `ErrInvalidValue` | 输入值为 nil |
| `ErrNilPointer` | 解码目标为 nil 指针 |
| `ErrNotPointer` | 解码目标不是指针 |
| `ErrNotStruct` | 目标不是结构体 |

## 结构体标签语法

```
`form:"field_name,option1,option2,..."`
```

| 选项 | 语法 | 说明 |
|------|------|------|
| 字段名 | `form:"custom_name"` | 自定义在 map 中的键名 |
| 忽略字段 | `form:"-"` | 编码和解码时跳过该字段 |
| 省略空值 | `form:"name,omitempty"` | 编码时跳过零值字段 |
| 默认值 | `form:"name,default=value"` | 键缺失时使用默认值 |
| 多值默认 | `form:"name,default=a\|b\|c"` | 多个默认值用 `\|` 分隔 |

> 注意：`omitempty` 和 `default` 选项可以任意顺序出现。

## 支持的类型

### 基础类型

- `string`
- `bool`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`

### 指针类型

所有基础类型的指针：`*string`, `*bool`, `*int`, `*float64` 等

### 切片类型

所有基础类型的切片：`[]string`, `[]int`, `[]float64` 等

### 内置特殊类型

- `time.Time` / `*time.Time` — 使用 RFC3339 格式编解码
- `[]byte` / `*[]byte` — 作为字符串整体处理

## 自定义类型

通过实现 `Marshaler` 和 `Unmarshaler` 接口，可以自定义任意类型的编解码行为。自定义接口的优先级高于内置类型转换。

```go
type StringList struct {
    Items []string
}

// 实现 Marshaler 接口
func (s StringList) MarshalValues() ([]string, error) {
    return s.Items, nil
}

// 实现 Unmarshaler 接口
func (s *StringList) UnmarshalValues(values []string) error {
    s.Items = values
    return nil
}

// 使用示例
type Form struct {
    Tags StringList `form:"tags"`
}

func main() {
    f := Form{Tags: StringList{Items: []string{"a", "b", "c"}}}
    values, _ := formcodec.Marshal(f)
    // values["tags"] = []string{"a", "b", "c"}
}
```

## 默认值

默认值在编码和解码中的行为：

- **解码时**：当 map 中缺少对应的 key 时，使用默认值填充字段
- **编码时**：配合 `omitempty` 使用时，如果字段值等于默认值则省略输出
- 多个默认值使用 `|` 作为分隔符

```go
type Config struct {
    // 解码时如果缺少 mode，默认为 "debug"
    Mode    string   `form:"mode,default=debug"`

    // 配合 omitempty，编码时如果值为 "debug" 则省略
    Mode2   string   `form:"mode2,omitempty,default=debug"`

    // 多值默认
    Regions []string `form:"regions,default=us|eu|asia"`
}
```

## 选项配置

可以通过选项自定义 tag 名称：

```go
// 使用自定义 tag 名称 "query"
type SearchParams struct {
    Keyword string `query:"q"`
    Page    int    `query:"page"`
}

// 编码
encoder := formcodec.NewEncoder(formcodec.WithEncoderTagName("query"))
values, err := encoder.Encode(SearchParams{Keyword: "golang", Page: 1})
// values = map[string][]string{"q": {"golang"}, "page": {"1"}}

// 解码
decoder := formcodec.NewDecoder(formcodec.WithDecoderTagName("query"))
var params SearchParams
err = decoder.Decode(values, &params)
```

## 许可证

Apache-2.0
