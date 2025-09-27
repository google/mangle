// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interpreter

import (
	"bytes"
	"strings"
	"testing"
)

func TestQueryInteractive_UnicodeStringsPrintedRaw(t *testing.T) {
	var buf bytes.Buffer
	i := New(&buf, "", nil)

	if err := i.Define(`jstr("私"). jstr("あなた").`); err != nil {
		t.Fatalf("Define failed: %v", err)
	}
	buf.Reset()

	if err := i.QueryInteractive(`jstr(X)`); err != nil {
		t.Fatalf("QueryInteractive failed: %v", err)
	}
	out := buf.String()

	if strings.Contains(out, `\u{`) {
		t.Fatalf("output contains escaped unicode: %q", out)
	}
	if !strings.Contains(out, `jstr("私")`) {
		t.Fatalf("expected output to contain jstr(\"私\"): %q", out)
	}
	if !strings.Contains(out, `jstr("あなた")`) {
		t.Fatalf("expected output to contain jstr(\"あなた\"): %q", out)
	}
	if !strings.Contains(out, "Found 2 entries for jstr(X).") {
		t.Fatalf("expected summary line, got: %q", out)
	}
}
