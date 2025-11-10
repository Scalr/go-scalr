package value

import (
	"encoding/json"
	"testing"
)

// TestValueStates demonstrates the three states: unset, null, and set
func TestValueStates(t *testing.T) {
	type TestStruct struct {
		Field1 *Value[string]   `json:"field1,omitempty"`
		Field2 *Value[string]   `json:"field2,omitempty"`
		Field3 *Value[string]   `json:"field3,omitempty"`
		Field4 *Value[[]string] `json:"field4,omitempty"`
		Field5 *Value[[]string] `json:"field5,omitempty"`
		Field6 *Value[[]string] `json:"field6,omitempty"`
	}

	tests := []struct {
		name     string
		input    TestStruct
		expected string
	}{
		{
			name: "unset via Unset() - should be omitted",
			input: TestStruct{
				Field1: Unset[string](),
			},
			expected: `{}`,
		},
		{
			name: "unset via nil - should be omitted",
			input: TestStruct{
				Field1: nil,
			},
			expected: `{}`,
		},
		{
			name:  "unset by not setting - should be omitted",
			input: TestStruct{
				// Field1 not set at all
			},
			expected: `{}`,
		},
		{
			name: "null via Null() - should be included as null",
			input: TestStruct{
				Field1: Null[string](),
			},
			expected: `{"field1":null}`,
		},
		{
			name: "set with value - should be included",
			input: TestStruct{
				Field1: Set("hello"),
			},
			expected: `{"field1":"hello"}`,
		},
		{
			name: "set with empty string - should be included",
			input: TestStruct{
				Field1: Set(""),
			},
			expected: `{"field1":""}`,
		},
		{
			name: "slice null - should be included as null",
			input: TestStruct{
				Field4: Null[[]string](),
			},
			expected: `{"field4":null}`,
		},
		{
			name: "slice empty - should be included as empty array",
			input: TestStruct{
				Field4: Set([]string{}),
			},
			expected: `{"field4":[]}`,
		},
		{
			name: "slice with values - should be included",
			input: TestStruct{
				Field4: Set([]string{"a", "b"}),
			},
			expected: `{"field4":["a","b"]}`,
		},
		{
			name: "slice unset - should be omitted",
			input: TestStruct{
				Field4: Unset[[]string](),
			},
			expected: `{}`,
		},
		{
			name: "mixed states",
			input: TestStruct{
				Field1: Set("value1"),
				Field2: Null[string](),
				Field3: nil, // unset
				Field4: Set([]string{"item"}),
				Field5: Set([]string{}),
				Field6: Unset[[]string](),
			},
			expected: `{"field1":"value1","field2":null,"field4":["item"],"field5":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			got := string(data)
			if got != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, got)
			}
		})
	}
}

// TestValueMethods tests the helper methods
func TestValueMethods(t *testing.T) {
	t.Run("IsSet", func(t *testing.T) {
		if Unset[string]().IsSet() {
			t.Error("Unset value should not be set")
		}
		if !Null[string]().IsSet() {
			t.Error("Null value should be set")
		}
		if !Set("hello").IsSet() {
			t.Error("Set value should be set")
		}
		var nilValue *Value[string]
		if nilValue.IsSet() {
			t.Error("nil value should not be set")
		}
	})

	t.Run("IsNull", func(t *testing.T) {
		if Unset[string]().IsNull() {
			t.Error("Unset value should not be null")
		}
		if !Null[string]().IsNull() {
			t.Error("Null value should be null")
		}
		if Set("hello").IsNull() {
			t.Error("Set value should not be null")
		}
		var nilValue *Value[string]
		if nilValue.IsNull() {
			t.Error("nil value should not be null")
		}
	})

	t.Run("Value", func(t *testing.T) {
		if _, ok := Unset[string]().Value(); ok {
			t.Error("Unset value should return false")
		}
		if _, ok := Null[string]().Value(); ok {
			t.Error("Null value should return false")
		}
		if v, ok := Set("hello").Value(); !ok || v != "hello" {
			t.Errorf("Set value should return true and 'hello', got %v, %v", v, ok)
		}
	})
}

// TestUnsetReturnsNil verifies that Unset() returns nil
func TestUnsetReturnsNil(t *testing.T) {
	if v := Unset[string](); v != nil {
		t.Errorf("Unset() should return nil, got %v", v)
	}
	if v := Unset[[]string](); v != nil {
		t.Errorf("Unset[[]string]() should return nil, got %v", v)
	}
}

// TestClear tests the Clear method
func TestClear(t *testing.T) {
	v := Set("hello")
	if !v.IsSet() {
		t.Error("Value should be set before Clear")
	}

	v.Clear()
	if v.IsSet() {
		t.Error("Value should not be set after Clear")
	}

	// Clear on nil should not panic
	var nilValue *Value[string]
	nilValue.Clear() // Should not panic
}
