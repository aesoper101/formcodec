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

func resetDecoderMap() {
	decoderMap = sync.Map{}
}

func TestFormDecode_Basic(t *testing.T) {
	resetDecoderMap()
	type S struct {
		Name string `form:"name"`
	}
	data := map[string][]string{"name": {"alice"}}
	var result S
	if err := Decode(data, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "alice" {
		t.Errorf("Name = %q, want %q", result.Name, "alice")
	}
}

func TestFormDecode_WithOptions(t *testing.T) {
	resetDecoderMap()
	type S struct {
		Name string `query:"name"`
	}
	data := map[string][]string{"name": {"bob"}}
	var result S
	if err := Decode(data, &result, WithDecoderTagName("query")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "bob" {
		t.Errorf("Name = %q, want %q", result.Name, "bob")
	}
}

func TestFormDecode_ErrorNil(t *testing.T) {
	resetDecoderMap()
	if err := Decode(nil, nil); err != ErrNilPointer {
		t.Errorf("error = %v, want ErrNilPointer", err)
	}
}

func TestFormDecode_ErrorNonPointer(t *testing.T) {
	resetDecoderMap()
	type S struct{ Name string }
	var s S
	if err := Decode(nil, s); err != ErrNotPointer {
		t.Errorf("error = %v, want ErrNotPointer", err)
	}
}

func TestFormDecode_ErrorNilPointer(t *testing.T) {
	resetDecoderMap()
	var s *struct{ Name string }
	if err := Decode(nil, s); err != ErrNilPointer {
		t.Errorf("error = %v, want ErrNilPointer", err)
	}
}

func TestFormDecode_ErrorNotStruct(t *testing.T) {
	resetDecoderMap()
	var s string
	if err := Decode(nil, &s); err != ErrNotStruct {
		t.Errorf("error = %v, want ErrNotStruct", err)
	}
}