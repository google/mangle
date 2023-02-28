// Copyright 2023 Google LLC
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

// Package json2struct contains code to convert JSON objects to Mangle structs.
package json2struct

import (
	"encoding/json"
	"fmt"

	"github.com/google/mangle/ast"
)

// JSONtoStruct converts a JSON blob to a Mangle struct.
func JSONtoStruct(jsonBlob []byte) (ast.Constant, error) {
	var m map[string]any
	if err := json.Unmarshal(jsonBlob, &m); err != nil {
		return ast.Constant{}, err
	}
	structEntries := map[*ast.Constant]*ast.Constant{}
	for k, v := range m {
		fieldName, err := ast.Name("/" + k)
		if err != nil {
			return ast.Constant{}, err
		}
		value, err := ConvertValue(v)
		if err != nil {
			return ast.Constant{}, err
		}
		structEntries[&fieldName] = &value
	}
	return *ast.Struct(structEntries), nil
}

// ConvertValue converts an encoding/json value into a Mangle value.
func ConvertValue(value any) (ast.Constant, error) {
	switch value := value.(type) {
	case string:
		return ast.String(value), nil
	case float64:
		return ast.Float64(value), nil
	case bool:
		if value {
			return ast.TrueConstant, nil
		}
		return ast.FalseConstant, nil
	case []any:
		l := &ast.ListNil
		for i := len(value) - 1; i >= 0; i-- {
			value, err := ConvertValue(value[i])
			if err != nil {
				return ast.Constant{}, err
			}
			next := ast.ListCons(&value, l)
			l = &next
		}
		return *l, nil
	case map[string]any:
		structEntries := map[*ast.Constant]*ast.Constant{}
		for k, v := range value {
			fieldName, err := ast.Name("/" + k)
			if err != nil {
				return ast.Constant{}, err
			}
			value, err := ConvertValue(v)
			if err != nil {
				return ast.Constant{}, err
			}
			structEntries[&fieldName] = &value
		}
		return *ast.Struct(structEntries), nil
	default:
		return ast.Constant{}, fmt.Errorf("unexpected value %v type %T: ", value, value)
	}
}
