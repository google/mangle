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

// Package proto2struct contains code to convert protocol buffers to Mangle structs.
package proto2struct

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
	"github.com/google/mangle/ast"
)

// ProtoToStruct uses reflection to convert a proto to a Mangle struct.
// The resulting struct only contains that are considered populated by the
// go protobuf runtime, i.e. proto fields set to default values will be missing.
func ProtoToStruct(msg protoreflect.Message) (ast.Constant, error) {
	r := &ast.StructNil
	var err error
	msg.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		fieldName, errLocal := ast.Name("/" + fd.TextName())
		if errLocal != nil {
			err = fmt.Errorf("Could not convert field name %v: %v", fieldName, errLocal)
			return false
		}
		if fd.IsMap() {
			entries := map[*ast.Constant]*ast.Constant{}
			// We cannot use ast.MapCons here since the runtime representation assumes a
			// particular order of the entries. So we collect all entries first.
			val.Map().Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
				key, errLocal := ProtoValueToConstant(fd.MapKey(), k.Value())
				if errLocal != nil {
					err = fmt.Errorf("Could not convert map key %v: %v", k, errLocal)
					return false
				}
				value, errLocal := ProtoValueToConstant(fd.MapValue(), v)
				if errLocal != nil {
					err = fmt.Errorf("Could not convert message field %v: %v", fd.TextName(), errLocal)
					return false
				}
				entries[&key] = &value
				return true
			})
			if err != nil {
				err = fmt.Errorf("Could not convert map %v: %v", val, err)
				return false
			}
			fields := ast.StructCons(&fieldName, ast.Map(entries), r)
			r = &fields
			return true
		}

		if fd.IsList() {
			l := &ast.ListNil
			for i := val.List().Len() - 1; 0 <= i; i-- {
				value, errLocal := ProtoValueToConstant(fd, val.List().Get(i))
				if errLocal != nil {
					err = errLocal
					return false
				}
				elem := ast.ListCons(&value, l)
				l = &elem
			}
			fields := ast.StructCons(&fieldName, l, r)
			r = &fields
			return true
		}
		value, errLocal := ProtoValueToConstant(fd, val)
		if errLocal != nil {
			err = errLocal
			return false
		}
		fields := ast.StructCons(&fieldName, &value, r)
		r = &fields
		return true
	})
	if err != nil {
		return ast.Constant{}, err
	}
	return *r, nil
}

// ProtoValueToConstant uses reflection to convert a proto field value to a Mangle value.
func ProtoValueToConstant(fd protoreflect.FieldDescriptor, val protoreflect.Value) (ast.Constant, error) {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		if val.Bool() {
			return ast.TrueConstant, nil
		}
		return ast.FalseConstant, nil
	case protoreflect.Sint32Kind:
		fallthrough
	case protoreflect.Sint64Kind:
		fallthrough
	case protoreflect.Int32Kind:
		fallthrough
	case protoreflect.Int64Kind:
		return ast.Number(val.Int()), nil
	case protoreflect.FloatKind:
		fallthrough
	case protoreflect.DoubleKind:
		return ast.Float64(val.Float()), nil
	case protoreflect.BytesKind:
		return ast.String(string(val.Bytes())), nil
	case protoreflect.StringKind:
		return ast.String(val.String()), nil
	case protoreflect.MessageKind:
		return ProtoToStruct(val.Message())
	case protoreflect.EnumKind:
		return ProtoEnumToConstant(fd.Enum(), val)
	}
	return ast.Constant{}, fmt.Errorf("proto field %v of unsupported kind: %v", fd, fd.Kind())
}

// ProtoEnumToConstant uses reflection to convert a proto field value to a Mangle value.
func ProtoEnumToConstant(ed protoreflect.EnumDescriptor, val protoreflect.Value) (ast.Constant, error) {
	enumValueDescr := ed.Values().ByNumber(val.Enum())
	if enumValueDescr == nil {
		return ast.Constant{}, fmt.Errorf("could not find enum value %d in %v", val.Enum(), ed)
	}
	return ast.Name("/" + string(ed.Name()) + "/" + string(enumValueDescr.Name()))
}
