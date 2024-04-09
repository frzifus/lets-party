// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package form

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

type TestStruct struct {
	UUIDField   uuid.UUID   `form:"uuid_field"`
	StringField string      `form:"string_field"`
	BoolField   bool        `form:"bool_field"`
	IntField    int         `form:"int_field"`
	FloatField  float64     `form:"float_field"`
	SliceField  []string    `form:"slice_field"`
	StructField FieldStruct `form:"struct_field"`
}

type FieldStruct struct {
	StringField string `form:"field_struct_strfield"`
	BoolField   bool   `form:"field_struct_boolfield"`
}

func TestUnmarshal(t *testing.T) {
	testCases := []struct {
		name        string
		input       url.Values
		expected    TestStruct
		expectedErr bool
	}{
		{
			name: "Valid input data",
			input: url.Values{
				"uuid_field":                          {"ca07d617-c87c-4ac3-affc-27a5e941b28f"},
				"string_field":                        {"test_string"},
				"bool_field":                          {"true"},
				"int_field":                           {"42"},
				"float_field":                         {"3.14"},
				"slice_field":                         {"1", "2", "3"},
				"struct_field.field_struct_strfield":  {"stringfield"},
				"struct_field.field_struct_boolfield": {"true"},
			},
			expected: TestStruct{
				UUIDField:   uuid.MustParse("ca07d617-c87c-4ac3-affc-27a5e941b28f"),
				StringField: "test_string",
				BoolField:   true,
				IntField:    42,
				FloatField:  3.14,
				SliceField:  []string{"1", "2", "3"},
				StructField: FieldStruct{
					StringField: "stringfield",
					BoolField:   true,
				},
			},
			expectedErr: false,
		},
		{
			name:        "Empty input",
			input:       url.Values{},
			expected:    TestStruct{},
			expectedErr: false,
		},
		{
			name: "Missing fields",
			input: url.Values{
				"string_field": {"test_string"},
			},
			expected: TestStruct{
				StringField: "test_string",
			},
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var target TestStruct
			err := Unmarshal(tc.input, &target)
			if (err != nil) != tc.expectedErr {
				t.Errorf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(target, tc.expected) {
				t.Errorf("Unmarshal did not produce expected result. got: %+v, expected: %+v", target, tc.expected)
			}
		})
	}
}
