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
	"sync"
	"testing"
)

func resetEncoderMap() {
	encoderMap = sync.Map{}
}

func TestFormEncode_Basic(t *testing.T) {
	resetEncoderMap()
	type S struct {
		Name string `form:"name"`
	}
	input := S{Name: "alice"}
	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals := result["name"]; len(vals) != 1 || vals[0] != "alice" {
		t.Errorf("name: got %v, want [alice]", vals)
	}
}

func TestFormEncode_WithOptions(t *testing.T) {
	resetEncoderMap()
	type S struct {
		Name string `query:"name"`
	}
	input := S{Name: "bob"}
	result, err := Encode(input, WithEncoderTagName("query"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals := result["name"]; len(vals) != 1 || vals[0] != "bob" {
		t.Errorf("name: got %v, want [bob]", vals)
	}
}

func TestFormEncode_ErrorNil(t *testing.T) {
	resetEncoderMap()
	_, err := Encode(nil)
	if err != ErrInvalidValue {
		t.Errorf("error = %v, want ErrInvalidValue", err)
	}
}

func TestFormEncode_ErrorNilPointer(t *testing.T) {
	resetEncoderMap()
	var ptr *struct{ Name string }
	_, err := Encode(ptr)
	if err != ErrInvalidValue {
		t.Errorf("error = %v, want ErrInvalidValue", err)
	}
}

func TestFormEncode_ErrorNotStruct(t *testing.T) {
	resetEncoderMap()
	_, err := Encode(42)
	if err != ErrNotStruct {
		t.Errorf("error = %v, want ErrNotStruct", err)
	}
}