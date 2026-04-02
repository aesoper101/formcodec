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

var (
	// ErrNilPointer is returned when a nil pointer is passed to Decode.
	ErrNilPointer = errors.New("formcodec: destination must be a non-nil pointer")
	// ErrNotPointer is returned when a non-pointer is passed to Decode.
	ErrNotPointer = errors.New("formcodec: destination must be a pointer")
	// ErrNotStruct is returned when the destination is not a pointer to a struct.
	ErrNotStruct = errors.New("formcodec: destination must be a pointer to a struct")
)

var unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

// Decoder decodes map[string][]string into Go structs.
type Decoder struct {
	opts decoderOptions
}

// NewDecoder creates a new Decoder with the given options.
func NewDecoder(opts ...DecoderOption) *Decoder {
	o := defaultDecoderOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return &Decoder{opts: o}
}

// Decode decodes map[string][]string into a struct.
// v must be a non-nil pointer to a struct.
func (d *Decoder) Decode(data map[string][]string, v any) error {
	if v == nil {
		return ErrNilPointer
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return ErrNotPointer
	}
	if rv.IsNil() {
		return ErrNilPointer
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	return d.decodeStruct(data, rv)
}

// Unmarshal decodes map[string][]string into a struct using default options.
// v must be a non-nil pointer to a struct.
func Unmarshal(data map[string][]string, v any) error {
	return NewDecoder().Decode(data, v)
}

func (d *Decoder) decodeStruct(data map[string][]string, rv reflect.Value) error {
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		info := parseTag(d.opts.tagName, field)
		if info.skip {
			continue
		}

		values, ok := data[info.name]
		if !ok || len(values) == 0 {
			// If default value is specified, use it
			if info.hasDefault {
				values = info.defaultValue
			} else {
				continue
			}
		}

		fieldValue := rv.Field(i)
		if err := d.decodeValue(fieldValue, values, info.name); err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoder) decodeValue(fieldValue reflect.Value, values []string, fieldName string) error {
	// Check if the field implements Unmarshaler interface
	if d.implementsUnmarshaler(fieldValue) {
		return d.callUnmarshalValues(fieldValue, values)
	}

	// Handle time.Time (after Unmarshaler check)
	if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
		if len(values) == 0 || values[0] == "" {
			return nil
		}
		t, err := time.Parse(time.RFC3339, values[0])
		if err != nil {
			return fmt.Errorf("formcodec: cannot parse time for field %s: %w", fieldName, err)
		}
		fieldValue.Set(reflect.ValueOf(t))
		return nil
	}

	// Handle []byte specially (before generic slice handling)
	if fieldValue.Type() == reflect.TypeOf([]byte{}) {
		if len(values) == 0 || values[0] == "" {
			return nil
		}
		fieldValue.SetBytes([]byte(values[0]))
		return nil
	}

	// Handle based on kind
	return d.decodeByKind(fieldValue, values, fieldName)
}

func (d *Decoder) implementsUnmarshaler(v reflect.Value) bool {
	// Don't call methods on nil pointers, let decodeByKind handle them
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return false
	}

	// Check if value type implements Unmarshaler
	if v.Type().Implements(unmarshalerType) {
		return true
	}

	// Check if pointer to value type implements Unmarshaler
	if v.CanAddr() && v.Addr().Type().Implements(unmarshalerType) {
		return true
	}

	return false
}

func (d *Decoder) callUnmarshalValues(v reflect.Value, values []string) error {
	var unmarshaler Unmarshaler

	if v.Type().Implements(unmarshalerType) {
		unmarshaler = v.Interface().(Unmarshaler)
	} else if v.CanAddr() && v.Addr().Type().Implements(unmarshalerType) {
		unmarshaler = v.Addr().Interface().(Unmarshaler)
	}

	if unmarshaler != nil {
		return unmarshaler.UnmarshalValues(values)
	}

	return nil
}

func (d *Decoder) decodeByKind(fieldValue reflect.Value, values []string, fieldName string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(values[0])

	case reflect.Bool:
		b, err := strconv.ParseBool(values[0])
		if err != nil {
			return fmt.Errorf("formcodec: cannot parse %q as bool for field %s: %w", values[0], fieldName, err)
		}
		fieldValue.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(values[0], 10, fieldValue.Type().Bits())
		if err != nil {
			return fmt.Errorf("formcodec: cannot parse %q as %s for field %s: %w", values[0], fieldValue.Kind(), fieldName, err)
		}
		fieldValue.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(values[0], 10, fieldValue.Type().Bits())
		if err != nil {
			return fmt.Errorf("formcodec: cannot parse %q as %s for field %s: %w", values[0], fieldValue.Kind(), fieldName, err)
		}
		fieldValue.SetUint(n)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(values[0], fieldValue.Type().Bits())
		if err != nil {
			return fmt.Errorf("formcodec: cannot parse %q as %s for field %s: %w", values[0], fieldValue.Kind(), fieldName, err)
		}
		fieldValue.SetFloat(f)

	case reflect.Pointer:
		// Create a new instance of the pointed-to type
		elemType := fieldValue.Type().Elem()
		newElem := reflect.New(elemType)

		// Check if the pointer type implements Unmarshaler
		if newElem.Type().Implements(unmarshalerType) {
			if err := newElem.Interface().(Unmarshaler).UnmarshalValues(values); err != nil {
				return err
			}
			fieldValue.Set(newElem)
			return nil
		}

		// Handle *time.Time specially
		if elemType == reflect.TypeOf(time.Time{}) {
			if len(values) == 0 || values[0] == "" {
				return nil
			}
			t, err := time.Parse(time.RFC3339, values[0])
			if err != nil {
				return fmt.Errorf("formcodec: cannot parse time for field %s: %w", fieldName, err)
			}
			newElem.Elem().Set(reflect.ValueOf(t))
			fieldValue.Set(newElem)
			return nil
		}

		// Handle *[]byte specially
		if elemType == reflect.TypeOf([]byte{}) {
			if len(values) == 0 || values[0] == "" {
				return nil // keep nil
			}
			newElem.Elem().SetBytes([]byte(values[0]))
			fieldValue.Set(newElem)
			return nil
		}

		// Decode into the new element
		if err := d.decodeByKind(newElem.Elem(), values, fieldName); err != nil {
			return err
		}
		fieldValue.Set(newElem)

	case reflect.Slice:
		elemType := fieldValue.Type().Elem()
		slice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))

		for i, val := range values {
			elemValue := slice.Index(i)

			// Check if element type implements Unmarshaler (need pointer for interface check)
			elemPtr := reflect.New(elemType)
			if elemPtr.Type().Implements(unmarshalerType) {
				if err := elemPtr.Interface().(Unmarshaler).UnmarshalValues([]string{val}); err != nil {
					return err
				}
				elemValue.Set(elemPtr.Elem())
				continue
			}

			// Also check if pointer to element type implements Unmarshaler
			if reflect.PointerTo(elemType).Implements(unmarshalerType) {
				if err := elemPtr.Interface().(Unmarshaler).UnmarshalValues([]string{val}); err != nil {
					return err
				}
				elemValue.Set(elemPtr.Elem())
				continue
			}

			// Use standard decoding
			if err := d.decodeByKind(elemValue, []string{val}, fieldName); err != nil {
				return err
			}
		}
		fieldValue.Set(slice)

	default:
		return fmt.Errorf("formcodec: unsupported type %s for field %s", fieldValue.Kind(), fieldName)
	}

	return nil
}
