// Copyright 2026 aesoper101
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package formcodec

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// ErrInvalidValue is returned when the input is nil.
var ErrInvalidValue = errors.New("formcodec: input value is nil")

var marshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()

// Encoder encodes Go structs into map[string][]string.
type Encoder struct {
	opts encoderOptions
}

// NewEncoder creates a new Encoder with the given options.
func NewEncoder(opts ...EncoderOption) *Encoder {
	o := defaultEncoderOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return &Encoder{opts: o}
}

// Encode encodes a struct into map[string][]string.
func (e *Encoder) Encode(v any) (map[string][]string, error) {
	if v == nil {
		return nil, ErrInvalidValue
	}

	rv := reflect.ValueOf(v)
	rt := rv.Type()

	// Handle pointer
	if rt.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, ErrInvalidValue
		}
		rv = rv.Elem()
		rt = rv.Type()
	}

	if rt.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	result := make(map[string][]string)

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse tag
		info := parseTag(e.opts.tagName, field)
		if info.skip {
			continue
		}

		// 快速路径：无默认值时，零值跳过（保持原行为）
		if info.omitempty && !info.hasDefault && isZeroValue(fieldValue) {
			continue
		}

		// 编码字段值
		values, err := e.encodeValue(fieldValue)
		if err != nil {
			return nil, fmt.Errorf("formcodec: error encoding field %s: %w", field.Name, err)
		}

		// 有默认值 + omitempty：值等于默认值时跳过
		if info.omitempty && info.hasDefault && slicesEqual(values, info.defaultValue) {
			continue
		}

		if values != nil {
			result[info.name] = values
		}
	}

	return result, nil
}

// encodeValue encodes a single reflect.Value into []string.
func (e *Encoder) encodeValue(v reflect.Value) ([]string, error) {
	// Check if the value implements Marshaler interface first (before dereferencing pointer)
	if values, ok, err := e.tryMarshal(v); ok {
		return values, err
	}

	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
		// Check again after dereferencing
		if values, ok, err := e.tryMarshal(v); ok {
			return values, err
		}
	}

	// Handle time.Time (after Marshaler check, before slice)
	if v.Type() == reflect.TypeOf(time.Time{}) {
		t := v.Interface().(time.Time)
		if t.IsZero() {
			return nil, nil // zero value time is treated as empty
		}
		return []string{t.Format(time.RFC3339)}, nil
	}

	// Handle []byte specially (before generic slice handling)
	if v.Type() == reflect.TypeOf([]byte{}) {
		if v.IsNil() || v.Len() == 0 {
			return nil, nil
		}
		return []string{string(v.Bytes())}, nil
	}

	// Handle slice
	if v.Kind() == reflect.Slice {
		return e.encodeSlice(v)
	}

	// Handle basic types
	return e.encodeBasicType(v)
}

// tryMarshal checks if the value implements Marshaler and calls MarshalValues if so.
func (e *Encoder) tryMarshal(v reflect.Value) ([]string, bool, error) {
	// Check value type implements Marshaler
	if v.Type().Implements(marshalerType) {
		if v.CanInterface() {
			m := v.Interface().(Marshaler)
			values, err := m.MarshalValues()
			return values, true, err
		}
	}

	// Check pointer type implements Marshaler
	ptrType := reflect.PointerTo(v.Type())
	if ptrType.Implements(marshalerType) {
		// If value is addressable, use its address directly
		if v.CanAddr() {
			pv := v.Addr()
			if pv.CanInterface() {
				m := pv.Interface().(Marshaler)
				values, err := m.MarshalValues()
				return values, true, err
			}
		} else {
			// Create a copy that is addressable
			tmp := reflect.New(v.Type())
			tmp.Elem().Set(v)
			if tmp.CanInterface() {
				m := tmp.Interface().(Marshaler)
				values, err := m.MarshalValues()
				return values, true, err
			}
		}
	}

	return nil, false, nil
}

// encodeSlice encodes a slice into []string.
func (e *Encoder) encodeSlice(v reflect.Value) ([]string, error) {
	if v.IsNil() {
		return nil, nil
	}

	result := make([]string, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)

		// Check if element implements Marshaler
		if values, ok, err := e.tryMarshal(elem); ok {
			if err != nil {
				return nil, err
			}
			result = append(result, values...)
			continue
		}

		// Handle pointer element
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}

		// Encode basic type element
		values, err := e.encodeBasicType(elem)
		if err != nil {
			return nil, err
		}
		result = append(result, values...)
	}

	return result, nil
}

// encodeBasicType encodes a basic type into []string.
func (e *Encoder) encodeBasicType(v reflect.Value) ([]string, error) {
	switch v.Kind() {
	case reflect.String:
		return []string{v.String()}, nil

	case reflect.Bool:
		return []string{strconv.FormatBool(v.Bool())}, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []string{strconv.FormatInt(v.Int(), 10)}, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []string{strconv.FormatUint(v.Uint(), 10)}, nil

	case reflect.Float32:
		return []string{strconv.FormatFloat(v.Float(), 'f', -1, 32)}, nil

	case reflect.Float64:
		return []string{strconv.FormatFloat(v.Float(), 'f', -1, 64)}, nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", v.Type())
	}
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

// slicesEqual checks if two string slices are equal.
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Marshal encodes a struct into map[string][]string using default options.
func Marshal(v any) (map[string][]string, error) {
	return NewEncoder().Encode(v)
}
