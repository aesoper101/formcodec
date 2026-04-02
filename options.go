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
	"reflect"
	"strings"
)

const (
	defaultTagName        = "form"
	defaultValueSeparator = "|"
)

// encoderOptions holds encoder-specific configuration.
type encoderOptions struct {
	tagName string
}

// decoderOptions holds decoder-specific configuration.
type decoderOptions struct {
	tagName string
}

// EncoderOption configures the Encoder.
type EncoderOption func(*encoderOptions)

// DecoderOption configures the Decoder.
type DecoderOption func(*decoderOptions)

// WithEncoderTagName sets the struct tag name for the Encoder.
func WithEncoderTagName(name string) EncoderOption {
	return func(o *encoderOptions) {
		o.tagName = name
	}
}

// WithDecoderTagName sets the struct tag name for the Decoder.
func WithDecoderTagName(name string) DecoderOption {
	return func(o *decoderOptions) {
		o.tagName = name
	}
}

func defaultEncoderOptions() encoderOptions {
	return encoderOptions{
		tagName: defaultTagName,
	}
}

func defaultDecoderOptions() decoderOptions {
	return decoderOptions{
		tagName: defaultTagName,
	}
}

// fieldInfo holds parsed struct tag information.
type fieldInfo struct {
	name         string
	omitempty    bool
	skip         bool
	hasDefault   bool     // whether a default value is specified
	defaultValue []string // parsed default values
}

// parseTag parses a struct field's tag with the given tag name.
func parseTag(tagName string, field reflect.StructField) fieldInfo {
	tag := field.Tag.Get(tagName)
	if tag == "-" {
		return fieldInfo{skip: true}
	}

	info := fieldInfo{}

	if tag == "" {
		info.name = field.Name
		return info
	}

	parts := strings.Split(tag, ",")
	if parts[0] == "" {
		info.name = field.Name
	} else {
		info.name = parts[0]
	}

	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			info.omitempty = true
		} else if strings.HasPrefix(opt, "default=") {
			defaultStr := strings.TrimPrefix(opt, "default=")
			info.hasDefault = true
			info.defaultValue = strings.Split(defaultStr, defaultValueSeparator)
		}
	}

	return info
}
