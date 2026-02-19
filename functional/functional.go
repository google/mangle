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

// Package functional provides evaluation of function expressions.
package functional

import (
	"errors"
	"fmt"
	"iter"
	"math"
	"strconv"
	"strings"
	"time"

	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/symbols"
)

var (
	// ErrDivisionByZero indicates a division by zero runtime error.
	ErrDivisionByZero = errors.New("div by zero")

	// errFound is used for exiting a loop
	errFound = errors.New("found")
)

// EvalExpr evaluates any apply-expression in b and applies subst.
func EvalExpr(b ast.BaseTerm, subst ast.Subst) (ast.BaseTerm, error) {
	expr, ok := b.(ast.ApplyFn)
	if !ok {
		return b.ApplySubstBase(subst), nil
	}
	return EvalApplyFn(expr, subst)
}

// EvalExprsBase evaluates any apply-expressions in args and applies subst.
func EvalExprsBase(args []ast.BaseTerm, subst ast.Subst) ([]ast.BaseTerm, error) {
	res := make([]ast.BaseTerm, len(args))
	for i, expr := range args {
		r, err := EvalExpr(expr, subst)
		if err != nil {
			return nil, err
		}
		res[i] = r
	}
	return res, nil
}

// EvalExprs evaluates any apply-expressions in args and applies subst.
func EvalExprs(args []ast.BaseTerm, subst ast.Subst) ([]ast.Constant, error) {
	res := make([]ast.Constant, len(args))
	for i, expr := range args {
		r, err := EvalExpr(expr, subst)
		if err != nil {
			return nil, err
		}
		c, ok := r.(ast.Constant)
		if !ok {
			return nil, fmt.Errorf("evaluation produced something that is not a value: %v %T", r, r)
		}
		res[i] = c
	}
	return res, nil
}

func getTimeLayout(precision string) (string, error) {
	switch precision {
	case "/year":
		return "2006", nil
	case "/month":
		return "2006-01", nil
	case "/day":
		return "2006-01-02", nil
	case "/hour":
		return "2006-01-02T15Z07:00", nil
	case "/minute":
		return "2006-01-02T15:04Z07:00", nil
	case "/second":
		return "2006-01-02T15:04:05Z07:00", nil
	case "/millisecond":
		return "2006-01-02T15:04:05.000Z07:00", nil
	case "/microsecond":
		return "2006-01-02T15:04:05.000000Z07:00", nil
	case "/nanosecond":
		return time.RFC3339Nano, nil
	default:
		return "", fmt.Errorf("unknown precision %q", precision)
	}
}

// EvalApplyFn evaluates a built-in function application.
func EvalApplyFn(applyFn ast.ApplyFn, subst ast.Subst) (ast.Constant, error) {
	evaluatedArgs, err := EvalExprs(applyFn.Args, subst)
	if err != nil {
		return ast.Constant{}, err
	}
	switch applyFn.Function.Symbol {
	case symbols.Append.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("expected 2 list arguments, got %d argument(s)", l)
		}
		listElems, err := evaluatedArgs[0].ListSeq()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("expected list got %v (%v) in %v (subst %v)", evaluatedArgs[0], evaluatedArgs[0].Type, applyFn, subst)
		}
		elem := evaluatedArgs[1]
		var res []ast.Constant
		for c := range listElems {
			res = append(res, c)
		}
		res = append(res, elem)
		return ast.List(res), nil

	case symbols.ListContains.Symbol: // fn:list:contains(List, Member)
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("expected 2 list arguments, got %d argument(s)", l)
		}
		listElems, err := evaluatedArgs[0].ListSeq()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("expected list got %v (%v) in %v (subst %v)", evaluatedArgs[0], evaluatedArgs[0].Type, applyFn, subst)
		}
		elem := evaluatedArgs[1]
		for c := range listElems {
			if c.Equals(elem) {
				return ast.TrueConstant, nil
			}
		}
		return ast.FalseConstant, nil

	case symbols.Cons.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("expected 2 list arguments, got %d argument(s)", l)
		}
		fst := evaluatedArgs[0]
		snd := evaluatedArgs[1]
		if snd.Type != ast.ListShape {
			return ast.Constant{}, fmt.Errorf("second argument has to be a list, got %v", snd.Type)
		}
		return ast.ListCons(&fst, &snd), nil

	case symbols.Pair.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("expected 2 list arguments, got %d argument(s)", l)
		}
		fst := evaluatedArgs[0]
		snd := evaluatedArgs[1]
		return ast.Pair(&fst, &snd), nil

	case symbols.Len.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 list argument, got %d argument(s)", l)
		}
		listElems, err := evaluatedArgs[0].ListSeq()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("expected list got %v (%v) in %v (subst %v)", evaluatedArgs[0], evaluatedArgs[0].Type, applyFn, subst)
		}
		var length int64
		for range listElems {
			length++
		}
		return ast.Number(length), nil

	case symbols.List.Symbol:
		list := &ast.ListNil
		for i := len(evaluatedArgs) - 1; i >= 0; i-- {
			elem := evaluatedArgs[i]
			next := ast.ListCons(&elem, list)
			list = &next
		}
		return *list, nil

	case symbols.Map.Symbol:
		if l := len(evaluatedArgs); l%2 != 0 {
			return ast.Constant{}, fmt.Errorf("expected even list argument, got %d argument(s)", l)
		}
		kvMap := make(map[*ast.Constant]*ast.Constant)
		for i := 0; i < len(evaluatedArgs); i++ {
			label := &evaluatedArgs[i]
			i++
			value := &evaluatedArgs[i]
			kvMap[label] = value
		}
		return *ast.Map(kvMap), nil

	case symbols.Struct.Symbol:
		if l := len(evaluatedArgs); l%2 != 0 {
			return ast.Constant{}, fmt.Errorf("expected even list argument, got %d argument(s)", l)
		}
		kvMap := make(map[*ast.Constant]*ast.Constant)
		for i := 0; i < len(evaluatedArgs); i++ {
			label := &evaluatedArgs[i]
			i++
			value := &evaluatedArgs[i]
			kvMap[label] = value
		}
		return *ast.Struct(kvMap), nil

	case symbols.Tuple.Symbol:
		if len(evaluatedArgs) == 1 {
			return evaluatedArgs[0], nil
		}
		if l := len(evaluatedArgs); l == 0 {
			return ast.Constant{}, fmt.Errorf("expected at least 1 argument, got %d argument(s)", l)
		}
		width := len(evaluatedArgs)
		pair, err := EvalApplyFn(ast.ApplyFn{symbols.Pair, applyFn.Args[width-2:]}, subst)
		if err != nil {
			return ast.Constant{}, fmt.Errorf("could not eval args %v (subst %v)", applyFn, subst)
		}
		tuple := &pair
		for i := width - 3; i >= 0; i-- {
			arg := evaluatedArgs[i]
			next := ast.Pair(&arg, tuple)
			tuple = &next
		}
		return *tuple, nil

	case symbols.Max.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 list argument, got %d argument(s)", l)
		}
		listElems, err := evaluatedArgs[0].ListSeq()
		if err != nil {
			return ast.Constant{}, err
		}
		return evalMax(listElems)

	case symbols.FloatMax.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 list argument, got %d argument(s)", l)
		}
		listElems, err := evaluatedArgs[0].ListSeq()
		if err != nil {
			return ast.Constant{}, err
		}
		return evalFloatMax(listElems)

	case symbols.Min.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 list argument, got %d argument(s)", l)
		}
		listElems, err := evaluatedArgs[0].ListSeq()
		if err != nil {
			return ast.Constant{}, err
		}
		return evalMin(listElems)

	case symbols.FloatMin.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 list argument, got %d argument(s)", l)
		}
		listElems, err := evaluatedArgs[0].ListSeq()
		if err != nil {
			return ast.Constant{}, err
		}
		return evalFloatMin(listElems)

	case symbols.NumberToString.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 argument, got %d argument(s)", l)
		}
		val := evaluatedArgs[0]
		if val.Type != ast.NumberType {
			return ast.Constant{}, fmt.Errorf("cannot convert to string: fn:number:to_string() only converts ast.NumberType type")
		}

		return ast.String(ast.FormatNumber(val.NumValue)), nil

	case symbols.Float64ToString.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 argument, got %d argument(s)", l)
		}
		val := evaluatedArgs[0]
		if val.Type != ast.Float64Type {
			return ast.Constant{}, fmt.Errorf("cannot convert to string: fn:float64:to_string() only converts ast.Float64Type type")
		}

		f, err := val.Float64Value()
		if err != nil {
			return ast.Constant{}, err
		}

		return ast.String(strconv.FormatFloat(f, 'f', -1, 64)), nil

	case symbols.NameRoot.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 argument, got %d argument(s)", l)
		}
		val := evaluatedArgs[0]
		if val.Type != ast.NameType {
			return ast.Constant{}, fmt.Errorf("cannot take root: fn:name:root() expects ast.NameType type")
		}
		i := strings.Index(val.Symbol[1:], "/")
		if i == -1 {
			return val, nil
		}
		n, err := ast.Name(val.Symbol[:i+1])
		return n, err

	case symbols.NameTip.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 argument, got %d argument(s)", l)
		}
		val := evaluatedArgs[0]
		if val.Type != ast.NameType {
			return ast.Constant{}, fmt.Errorf("cannot take tip: fn:name:tip() expects ast.NameType type")
		}
		i := strings.LastIndex(val.Symbol, "/")
		if i == 0 {
			return val, nil
		}
		n, _ := ast.Name(val.Symbol[i:])
		return n, nil

	case symbols.NameList.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 argument, got %d argument(s)", l)
		}
		val := evaluatedArgs[0]
		if val.Type != ast.NameType {
			return ast.Constant{}, fmt.Errorf("cannot convert to list: fn:name:list() expects ast.NameType type")
		}

		i := len(val.Symbol)
		j := i
		list := &ast.ListNil
		for {
			i--
			if i == -1 {
				break
			}
			if val.Symbol[i] == '/' {
				elem, _ := ast.Name(val.Symbol[i:j])
				tmp := ast.ListCons(&elem, list)
				list = &tmp
				j = i
			}
		}
		return *list, nil

	case symbols.NameToString.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 argument, got %d argument(s)", l)
		}
		val := evaluatedArgs[0]
		if val.Type != ast.NameType {
			return ast.Constant{}, fmt.Errorf("cannot convert to string: fn:name:to_string() only converts ast.NameType type")
		}

		return ast.String(val.Symbol), nil

	case symbols.StringConcatenate.Symbol:
		var values []string
		for i, val := range evaluatedArgs {
			res, err := evalToString(val)
			if err != nil {
				return ast.Constant{}, fmt.Errorf("cannot string concatenate at position %v: %v", i, err)
			}
			values = append(values, res.Symbol)
		}
		return ast.String(strings.Join(values, "")), nil

	case symbols.StringReplace.Symbol:
		if l := len(evaluatedArgs); l != 4 {
			return ast.Constant{}, fmt.Errorf("expected 4 arguments, got %d argument(s)", l)
		}
		provided := evaluatedArgs[0]
		old := evaluatedArgs[1]
		new := evaluatedArgs[2]
		count := evaluatedArgs[3]
		if provided.Type != ast.StringType {
			return ast.Constant{}, fmt.Errorf("cannot string replace: fn:string:replace() expects ast.StringType type for 1st argument")
		}
		if old.Type != ast.StringType {
			return ast.Constant{}, fmt.Errorf("cannot string replace: fn:string:replace() expects ast.StringType type for 2nd argument")
		}
		if new.Type != ast.StringType {
			return ast.Constant{}, fmt.Errorf("cannot string replace: fn:string:replace() expects ast.StringType type for 3rd argument")
		}
		if count.Type != ast.NumberType {
			return ast.Constant{}, fmt.Errorf("cannot string replace: fn:string:replace() expects ast.NumberType type for 4th argument")
		}
		countValue, _ := count.NumberValue()
		return ast.String(strings.Replace(provided.Symbol, old.Symbol, new.Symbol, int(countValue))), nil

	case symbols.Sum.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 list argument, got %d argument(s)", l)
		}
		listElems, err := evaluatedArgs[0].ListSeq()
		if err != nil {
			return ast.Constant{}, err
		}
		return evalSum(listElems)

	case symbols.FloatSum.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 list argument, got %d argument(s)", l)
		}
		listElems, err := evaluatedArgs[0].ListSeq()
		if err != nil {
			return ast.Constant{}, err
		}
		return evalFloatSum(listElems)

	case symbols.ListGet.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("expected 2 arguments, got %d argument(s)", l)
		}
		arg := evaluatedArgs[0]
		indexConstant := evaluatedArgs[1]
		index, err := indexConstant.NumberValue()
		if err != nil {
			return ast.Constant{}, err
		}
		i := 0
		var res *ast.Constant
		_, loopErr := arg.ListValues(func(c ast.Constant) error {
			if i == int(index) {
				res = &c
				return errFound
			}
			i++
			return nil
		}, func() error {
			return nil
		})
		if errors.Is(loopErr, errFound) {
			return *res, nil
		}
		return ast.Constant{}, fmt.Errorf("index out of bounds: %d", index)

	case symbols.MapGet.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("expected 2 arguments, got %d argument(s)", l)
		}
		arg := evaluatedArgs[0] // map value
		lookupKey := evaluatedArgs[1]
		var res *ast.Constant
		_, loopErr := arg.MapValues(func(key ast.Constant, val ast.Constant) error {
			if key.Equals(lookupKey) {
				res = &val
				return errFound
			}
			return nil
		}, func() error {
			return nil
		})
		if errors.Is(loopErr, errFound) {
			return *res, nil
		}
		return ast.Constant{}, fmt.Errorf("key does not exist: %v", lookupKey)

	case symbols.StructGet.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("expected 2 arguments, got %d argument(s)", l)
		}
		arg := evaluatedArgs[0] // struct value
		lookupField := evaluatedArgs[1]
		var res *ast.Constant
		_, loopErr := arg.StructValues(func(field ast.Constant, val ast.Constant) error {
			if field.Equals(lookupField) {
				res = &val
				return errFound
			}
			return nil
		}, func() error {
			return nil
		})
		if errors.Is(loopErr, errFound) {
			return *res, nil
		}
		return ast.Constant{}, fmt.Errorf("key does not exist: %v", lookupField)

		// Time functions
	case symbols.TimeNow.Symbol:
		return ast.Time(time.Now().UnixNano()), nil

	case symbols.TimeAdd.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("fn:time:add expected 2 arguments, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:add first argument must be time: %w", err)
		}
		d, err := evaluatedArgs[1].DurationValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:add second argument must be duration: %w", err)
		}
		return ast.Time(t + d), nil

	case symbols.TimeSub.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("fn:time:sub expected 2 arguments, got %d", l)
		}
		t1, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:sub first argument must be time: %w", err)
		}
		t2, err := evaluatedArgs[1].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:sub second argument must be time: %w", err)
		}
		return ast.Duration(t1 - t2), nil

	case symbols.TimeFormat.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("fn:time:format expected 2 arguments, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:format first argument must be time: %w", err)
		}
		precision, err := evaluatedArgs[1].NameValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:format second argument must be a name constant: %w", err)
		}
		tm := time.Unix(0, t).UTC()

		pattern, err := getTimeLayout(precision)
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:format %w", err)
		}

		return ast.String(tm.Format(pattern)), nil

	case symbols.TimeFormatCivil.Symbol:
		if l := len(evaluatedArgs); l != 3 {
			return ast.Constant{}, fmt.Errorf("fn:time:format_civil expected 3 arguments, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:format_civil first argument must be time: %w", err)
		}
		tz, err := evaluatedArgs[1].StringValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:format_civil second argument must be string: %w", err)
		}
		precision, err := evaluatedArgs[2].NameValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:format_civil third argument must be name constant: %w", err)
		}

		loc, err := time.LoadLocation(tz)
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:format_civil unknown timezone %q: %w", tz, err)
		}

		tm := time.Unix(0, t).In(loc)

		pattern, err := getTimeLayout(precision)
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:format_civil %w", err)
		}

		return ast.String(tm.Format(pattern)), nil

	case symbols.TimeParseRFC3339.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:time:parse_rfc3339 expected 1 argument, got %d", l)
		}
		s, err := evaluatedArgs[0].StringValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:parse_rfc3339 argument must be string: %w", err)
		}
		tm, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:parse_rfc3339 failed: %w", err)
		}
		return ast.Time(tm.UTC().UnixNano()), nil

	case symbols.TimeParseCivil.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("fn:time:parse_civil expected 2 arguments, got %d", l)
		}
		s, err := evaluatedArgs[0].StringValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:parse_civil first argument must be string: %w", err)
		}
		tz, err := evaluatedArgs[1].StringValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:parse_civil second argument must be string: %w", err)
		}
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:parse_civil unknown timezone %q: %w", tz, err)
		}
		// Try parsing with nanoseconds first
		tm, err := time.ParseInLocation("2006-01-02T15:04:05.000000000", s, loc)
		if err != nil {
			// Fallback to seconds
			tm, err = time.ParseInLocation("2006-01-02T15:04:05", s, loc)
			if err != nil {
				return ast.Constant{}, fmt.Errorf("fn:time:parse_civil failed: %w", err)
			}
		}
		return ast.Time(tm.UTC().UnixNano()), nil

	case symbols.TimeYear.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:time:year expected 1 argument, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:year argument must be time: %w", err)
		}
		return ast.Number(int64(time.Unix(0, t).UTC().Year())), nil

	case symbols.TimeMonth.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:time:month expected 1 argument, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:month argument must be time: %w", err)
		}
		return ast.Number(int64(time.Unix(0, t).UTC().Month())), nil

	case symbols.TimeDay.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:time:day expected 1 argument, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:day argument must be time: %w", err)
		}
		return ast.Number(int64(time.Unix(0, t).UTC().Day())), nil

	case symbols.TimeHour.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:time:hour expected 1 argument, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:hour argument must be time: %w", err)
		}
		return ast.Number(int64(time.Unix(0, t).UTC().Hour())), nil

	case symbols.TimeMinute.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:time:minute expected 1 argument, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:minute argument must be time: %w", err)
		}
		return ast.Number(int64(time.Unix(0, t).UTC().Minute())), nil

	case symbols.TimeSecond.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:time:second expected 1 argument, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:second argument must be time: %w", err)
		}
		return ast.Number(int64(time.Unix(0, t).UTC().Second())), nil

	case symbols.TimeFromUnixNanos.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:time:from_unix_nanos expected 1 argument, got %d", l)
		}
		n, err := evaluatedArgs[0].NumberValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:from_unix_nanos argument must be number: %w", err)
		}
		return ast.Time(n), nil

	case symbols.TimeToUnixNanos.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:time:to_unix_nanos expected 1 argument, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:to_unix_nanos argument must be time: %w", err)
		}
		return ast.Number(t), nil

	case symbols.TimeTrunc.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("fn:time:trunc expected 2 arguments, got %d", l)
		}
		t, err := evaluatedArgs[0].TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:trunc first argument must be time: %w", err)
		}
		unitName, err := evaluatedArgs[1].NameValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:time:trunc second argument must be a name constant: %w", err)
		}

		var d time.Duration
		switch unitName {
		case "/hour":
			d = time.Hour
		case "/minute":
			d = time.Minute
		case "/second":
			d = time.Second
		case "/millisecond":
			d = time.Millisecond
		case "/microsecond":
			d = time.Microsecond
		case "/nanosecond":
			d = time.Nanosecond
		default:
			// "day", "month", "year" are not fixed durations, so Truncate doesn't support them directly in the same way.
			// However, for "day" we can use 24 * time.Hour if we assume UTC day boundaries.
			// Let's support /day for convenience, assuming standard 24h days.
			if unitName == "/day" {
				d = 24 * time.Hour
			} else {
				return ast.Constant{}, fmt.Errorf("fn:time:trunc unknown unit %q", unitName)
			}
		}

		tm := time.Unix(0, t).UTC()
		return ast.Time(tm.Truncate(d).UnixNano()), nil

	// Duration functions
	case symbols.DurationAdd.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("fn:duration:add expected 2 arguments, got %d", l)
		}
		d1, err := evaluatedArgs[0].DurationValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:add first argument must be duration: %w", err)
		}
		d2, err := evaluatedArgs[1].DurationValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:add second argument must be duration: %w", err)
		}
		return ast.Duration(d1 + d2), nil

	case symbols.DurationMult.Symbol:
		if l := len(evaluatedArgs); l != 2 {
			return ast.Constant{}, fmt.Errorf("fn:duration:mult expected 2 arguments, got %d", l)
		}
		d, err := evaluatedArgs[0].DurationValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:mult first argument must be duration: %w", err)
		}
		n, err := evaluatedArgs[1].NumberValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:mult second argument must be number: %w", err)
		}
		return ast.Duration(d * n), nil

	case symbols.DurationHours.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:duration:hours expected 1 argument, got %d", l)
		}
		d, err := evaluatedArgs[0].DurationValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:hours argument must be duration: %w", err)
		}
		return ast.Float64(time.Duration(d).Hours()), nil

	case symbols.DurationMinutes.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:duration:minutes expected 1 argument, got %d", l)
		}
		d, err := evaluatedArgs[0].DurationValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:minutes argument must be duration: %w", err)
		}
		return ast.Float64(time.Duration(d).Minutes()), nil

	case symbols.DurationSeconds.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:duration:seconds expected 1 argument, got %d", l)
		}
		d, err := evaluatedArgs[0].DurationValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:seconds argument must be duration: %w", err)
		}
		return ast.Float64(time.Duration(d).Seconds()), nil

	case symbols.DurationNanos.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:duration:nanos expected 1 argument, got %d", l)
		}
		d, err := evaluatedArgs[0].DurationValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:nanos argument must be duration: %w", err)
		}
		return ast.Number(d), nil

	case symbols.DurationFromNanos.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:duration:from_nanos expected 1 argument, got %d", l)
		}
		n, err := evaluatedArgs[0].NumberValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:from_nanos argument must be number: %w", err)
		}
		return ast.Duration(n), nil

	case symbols.DurationFromHours.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:duration:from_hours expected 1 argument, got %d", l)
		}
		h, err := valueAsFloat(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:from_hours argument must be number: %w", err)
		}
		return ast.Duration(int64(h * float64(time.Hour))), nil

	case symbols.DurationFromMinutes.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:duration:from_minutes expected 1 argument, got %d", l)
		}
		m, err := valueAsFloat(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:from_minutes argument must be number: %w", err)
		}
		return ast.Duration(int64(m * float64(time.Minute))), nil

	case symbols.DurationFromSeconds.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:duration:from_seconds expected 1 argument, got %d", l)
		}
		s, err := valueAsFloat(evaluatedArgs[0])
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:from_seconds argument must be number: %w", err)
		}
		return ast.Duration(int64(s * float64(time.Second))), nil

	case symbols.DurationParse.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("fn:duration:parse expected 1 argument, got %d", l)
		}
		str, err := evaluatedArgs[0].StringValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:parse argument must be string: %w", err)
		}
		// Go's time.ParseDuration supports: h, m, s, ms, us (or µs), ns
		// Examples: "1h30m", "500ms", "-2h45m30s", "1.5h"
		d, err := time.ParseDuration(str)
		if err != nil {
			return ast.Constant{}, fmt.Errorf("fn:duration:parse failed: %w", err)
		}
		return ast.Duration(int64(d)), nil

	// Interval functions - work with intervals represented as pairs of time values
	case symbols.IntervalStart.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 interval argument, got %d argument(s)", l)
		}
		interval := evaluatedArgs[0]
		if interval.Type != ast.PairShape {
			return ast.Constant{}, fmt.Errorf("expected pair (interval), got %v", interval.Type)
		}
		start, _, err := interval.PairValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("invalid interval: %w", err)
		}
		return start, nil

	case symbols.IntervalEnd.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 interval argument, got %d argument(s)", l)
		}
		interval := evaluatedArgs[0]
		if interval.Type != ast.PairShape {
			return ast.Constant{}, fmt.Errorf("expected pair (interval), got %v", interval.Type)
		}
		_, end, err := interval.PairValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("invalid interval: %w", err)
		}
		return end, nil

	case symbols.IntervalDuration.Symbol:
		if l := len(evaluatedArgs); l != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 interval argument, got %d argument(s)", l)
		}
		interval := evaluatedArgs[0]
		if interval.Type != ast.PairShape {
			return ast.Constant{}, fmt.Errorf("expected pair (interval), got %v", interval.Type)
		}
		start, end, err := interval.PairValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("invalid interval: %w", err)
		}
		startNano, err := start.TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("invalid interval start: %w", err)
		}
		endNano, err := end.TimeValue()
		if err != nil {
			return ast.Constant{}, fmt.Errorf("invalid interval end: %w", err)
		}
		// Handle unbounded intervals (represented by special values)
		if startNano == math.MinInt64 || endNano == math.MaxInt64 {
			return ast.Constant{}, fmt.Errorf("cannot compute duration of unbounded interval")
		}
		return ast.Duration(endNano - startNano), nil

	default:
		return EvalNumericApplyFn(applyFn, subst)
	}
}

// EvalNumericApplyFn evaluates a numeric built-in function application.
func EvalNumericApplyFn(applyFn ast.ApplyFn, subst ast.Subst) (ast.Constant, error) {
	args, err := EvalExprs(applyFn.Args, subst)
	if err != nil {
		return ast.Constant{}, err
	}

	switch applyFn.Function.Symbol {
	case symbols.Div.Symbol:
		res, err := evalDiv(args)
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.Number(res), nil
	case symbols.Mult.Symbol:
		res, err := evalMult(args)
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.Number(res), nil
	case symbols.Plus.Symbol:
		res, err := evalPlus(args)
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.Number(res), nil
	case symbols.Minus.Symbol:
		res, err := evalMinus(args)
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.Number(res), nil
	case symbols.FloatDiv.Symbol:
		res, err := evalFloatDiv(args)
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.Float64(res), nil
	case symbols.FloatMult.Symbol:
		resF, err := evalFloatMult(args)
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.Float64(resF), nil
	case symbols.Sqrt.Symbol:
		if len(args) != 1 {
			return ast.Constant{}, fmt.Errorf("expected 1 argument for sqrt, got %d", len(args))
		}
		fval, err := valueAsFloat(args[0])
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.Float64(math.Sqrt(fval)), nil
	case symbols.FloatPlus.Symbol:
		resF, err := evalFloatPlus(args)
		if err != nil {
			return ast.Constant{}, err
		}
		return ast.Float64(resF), nil
	default:
		return ast.Constant{}, fmt.Errorf("unknown function %s in %s", applyFn, applyFn.Function)
	}
}

func evalDiv(args []ast.Constant) (int64, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("empty argument list")
	}
	if len(args) == 1 {
		v, err := args[0].NumberValue()
		if err != nil {
			return 0, err
		}
		switch v {
		case 0:
			return 0, ErrDivisionByZero
		case 1:
			return 1, nil
		default:
			return 0, nil // integer division 1 / arg[0]
		}
	}
	res, err := args[0].NumberValue()
	if err != nil {
		return 0, err
	}
	for _, divisorConst := range args[1:] {
		divisor, err := divisorConst.NumberValue()
		if err != nil {
			return 0, err
		}
		if divisor == 0 {
			return 0, ErrDivisionByZero
		}
		res = res / divisor
		if res == 0 {
			return 0, nil
		}
	}
	return res, nil
}

func valueAsFloat(a ast.Constant) (float64, error) {
	switch a.Type {
	case ast.Float64Type:
		f, err := a.Float64Value()
		if err != nil {
			return 0, err
		}
		return f, nil
	case ast.NumberType:
		v, err := a.NumberValue()
		if err != nil {
			return 0, err
		}
		return float64(v), nil
	default:
		return 0, fmt.Errorf("unsupported non numeric type %v", a.Type)
	}
}

func evalFloatDiv(args []ast.Constant) (float64, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("empty argument list")
	}
	if len(args) == 1 {
		f, err := valueAsFloat(args[0])
		if err != nil {
			return 0, err
		}
		if f == 0 {
			return 0, ErrDivisionByZero
		}
		return 1 / f, nil
	}
	res, err := valueAsFloat(args[0])
	if err != nil {
		return 0, err
	}
	for _, divisorConst := range args[1:] {
		divisor, err := valueAsFloat(divisorConst)
		if err != nil {
			return 0, err
		}
		if divisor == 0 {
			return 0, ErrDivisionByZero
		}
		res = res / divisor
	}
	return res, nil
}

func evalFloatMult(args []ast.Constant) (float64, error) {
	var product float64 = 1
	for _, c := range args {
		f, err := valueAsFloat(c)
		if err != nil {
			return 0, err
		}
		product *= f
	}
	return product, nil
}

func evalFloatPlus(args []ast.Constant) (float64, error) {
	var sum float64
	for _, c := range args {
		f, err := valueAsFloat(c)
		if err != nil {
			return 0, err
		}
		sum += f
	}
	return sum, nil
}

func evalMult(args []ast.Constant) (int64, error) {
	var product int64 = 1
	for _, factor := range args {
		v, err := factor.NumberValue()
		if err != nil {
			return 0, err
		}
		product = product * v
	}
	return product, nil
}

func evalPlus(args []ast.Constant) (int64, error) {
	var sum int64 = 0
	for _, num := range args {
		v, err := num.NumberValue()
		if err != nil {
			return 0, err
		}
		sum += v
	}
	return sum, nil
}

func evalMinus(args []ast.Constant) (int64, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("empty argument list")
	}
	diff, err := args[0].NumberValue()
	if err != nil {
		return 0, err
	}
	if len(args) == 1 {
		return -diff, nil // "unary minus"
	}
	for _, num := range args[1:] {
		v, err := num.NumberValue()
		if err != nil {
			return 0, err
		}
		diff -= v
	}
	return diff, nil
}

// EvalReduceFn evaluates a combiner (reduce) function.
func EvalReduceFn(reduceFn ast.ApplyFn, rows []ast.ConstSubstList) (ast.Constant, error) {
	distinct := false
	rowsIter := func(v ast.Variable) iter.Seq[ast.Constant] {
		return func(yield func(ast.Constant) bool) {
			for _, subst := range rows {
				if num, ok := subst.Get(v).(ast.Constant); ok {
					if ok := yield(num); !ok {
						return
					}
				}
			}
		}
	}

	switch reduceFn.Function.Symbol {
	case symbols.CollectDistinct.Symbol:
		distinct = true
		fallthrough
	case symbols.Collect.Symbol:
		tuples := &ast.ListNil

		seen := make(map[uint64][]ast.Constant)
	rowloop:
		for i := len(rows) - 1; i >= 0; i-- {
			subst := rows[i]
			tuple := make([]ast.BaseTerm, 0, len(reduceFn.Args))
			for _, v := range reduceFn.Args {
				v, err := EvalExpr(v, subst)
				if err != nil {
					continue rowloop
				}
				constant, ok := v.(ast.Constant)
				if !ok {
					continue rowloop
				}
				tuple = append(tuple, constant)
			}
			head, err := EvalApplyFn(ast.ApplyFn{symbols.Tuple, tuple}, subst)
			if err != nil {
				continue rowloop
			}
			if !distinct {
				next := ast.ListCons(&head, tuples)
				tuples = &next
				continue
			}

			previousConstants, ok := seen[head.Hash()]
			if ok {
				for _, prev := range previousConstants {
					if prev.Equals(head) {
						continue rowloop
					}
				}
				seen[head.Hash()] = append(seen[head.Hash()], head)
			} else {
				seen[head.Hash()] = []ast.Constant{head}
			}
			next := ast.ListCons(&head, tuples)
			tuples = &next
		}
		return *tuples, nil
	case symbols.CollectToMap.Symbol:
		if len(reduceFn.Args) != 2 {
			return ast.Constant{}, fmt.Errorf("collect_to_map requires exactly 2 arguments (key, value), got %d", len(reduceFn.Args))
		}

		kvMap := make(map[*ast.Constant]*ast.Constant)
		seen := make(map[uint64][]ast.Constant)

	mapRowLoop:
		for _, subst := range rows {
			keyTerm, err := EvalExpr(reduceFn.Args[0], subst)
			if err != nil {
				continue mapRowLoop
			}
			keyConstant, ok := keyTerm.(ast.Constant)
			if !ok {
				continue mapRowLoop
			}

			valueTerm, err := EvalExpr(reduceFn.Args[1], subst)
			if err != nil {
				continue mapRowLoop
			}
			valueConstant, ok := valueTerm.(ast.Constant)
			if !ok {
				continue mapRowLoop
			}

			// Check if we've seen this key before (for deduplication)
			keyHash := keyConstant.Hash()
			if existingKeys, exists := seen[keyHash]; exists {
				shouldSkip := false
				for _, existingKey := range existingKeys {
					if existingKey.Equals(keyConstant) {
						shouldSkip = true
						break
					}
				}
				if shouldSkip {
					continue mapRowLoop
				}
				seen[keyHash] = append(existingKeys, keyConstant)
			} else {
				seen[keyHash] = []ast.Constant{keyConstant}
			}

			kvMap[&keyConstant] = &valueConstant
		}

		return *ast.Map(kvMap), nil
	case symbols.Count.Symbol:
		return ast.Number(int64(len(rows))), nil
	case symbols.CountDistinct.Symbol:
		domain := rows[0].Domain()
		var numDistinct int64 = 0
		seen := make(map[uint64][]ast.ConstSubstList)
		for _, row := range rows {
			newTuple := row.GetRow(domain)
			rowHash := ast.HashConstants(newTuple)
			slot := seen[rowHash]
			shouldAdd := true
			if slot != nil {
				shouldAdd = false
				// Check for collisions.
				for _, subst := range slot {
					existing := subst.GetRow(domain)
					if !ast.EqualsConstants(existing, newTuple) {
						shouldAdd = true
						break
					}
				}
			}
			if shouldAdd {
				seen[rowHash] = append(slot, row)
				numDistinct++
			}
		}
		return ast.Number(int64(numDistinct)), nil
	case symbols.Avg.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalAvg(rowsIter(v))

	case symbols.Max.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalMax(rowsIter(v))
	case symbols.FloatMax.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalFloatMax(rowsIter(v))
	case symbols.Min.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalMin(rowsIter(v))
	case symbols.FloatMin.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalFloatMin(rowsIter(v))
	case symbols.Sum.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalSum(rowsIter(v))
	case symbols.FloatSum.Symbol:
		v := reduceFn.Args[0].(ast.Variable)
		return evalFloatSum(rowsIter(v))
	default:
		return ast.Constant{}, fmt.Errorf("unknown reducer %v", reduceFn.Function)
	}
}

// EvalAtom returns an atom with any apply-expressions evaluated.
func EvalAtom(a ast.Atom, subst ast.Subst) (ast.Atom, error) {
	args, err := EvalExprsBase(a.Args, subst)
	if err != nil {
		return ast.Atom{}, err
	}
	return ast.Atom{a.Predicate, args}, nil
}

// EvalBaseTermPair evaluates a pair of base terms.
func EvalBaseTermPair(left, right ast.BaseTerm, subst ast.Subst) (ast.BaseTerm, ast.BaseTerm, error) {
	var err error
	left, err = EvalExpr(left, subst)
	if err != nil {
		return nil, nil, err
	}
	right, err = EvalExpr(right, subst)
	if err != nil {
		return nil, nil, err
	}
	return left, right, nil
}

func evalToString(val ast.Constant) (ast.Constant, error) {
	if val.Type == ast.StringType {
		return val, nil
	}

	var toStringFun ast.FunctionSym
	switch val.Type {
	case ast.NameType:
		toStringFun = symbols.NameToString
	case ast.NumberType:
		toStringFun = symbols.NumberToString
	case ast.Float64Type:
		toStringFun = symbols.Float64ToString
	default:
		return ast.Constant{}, fmt.Errorf("cannot convert constant to string constant: no conversion for value %v defined", val)
	}

	term := ast.ApplyFn{toStringFun, []ast.BaseTerm{val}}
	res, err := EvalApplyFn(term, ast.ConstSubstMap{})
	if err != nil {
		return ast.Constant{}, err
	}

	if res.Type != ast.StringType {
		return ast.Constant{}, fmt.Errorf("cannot convert constant to string constant: return value of conversion is not of type string")
	}
	return res, nil
}

func evalMax(it iter.Seq[ast.Constant]) (ast.Constant, error) {
	max := int64(math.MinInt64)
	for c := range it {
		num, err := c.NumberValue()
		if err != nil {
			return ast.Constant{}, err
		}
		if num > max {
			max = num
		}
	}
	return ast.Number(max), nil
}

func evalFloatMax(it iter.Seq[ast.Constant]) (ast.Constant, error) {
	max := -1 * math.MaxFloat64
	for c := range it {
		num, err := c.Float64Value()
		if err != nil {
			return ast.Constant{}, err
		}
		if num > max {
			max = num
		}
	}
	return ast.Float64(max), nil
}

func evalMin(it iter.Seq[ast.Constant]) (ast.Constant, error) {
	min := int64(math.MaxInt64)
	for c := range it {
		num, err := c.NumberValue()
		if err != nil {
			return ast.Constant{}, err
		}
		if num < min {
			min = num
		}
	}
	return ast.Number(min), nil
}

func evalFloatMin(it iter.Seq[ast.Constant]) (ast.Constant, error) {
	min := math.MaxFloat64
	for c := range it {
		floatNum, err := c.Float64Value()
		if err != nil {
			return ast.Constant{}, err
		}
		if floatNum < min {
			min = floatNum
		}
	}
	return ast.Float64(min), nil
}

func evalSum(it iter.Seq[ast.Constant]) (ast.Constant, error) {
	var sum int64
	for c := range it {
		num, err := c.NumberValue()
		if err != nil {
			return ast.Constant{}, err
		}
		sum += num
	}
	return ast.Number(sum), nil
}

func evalFloatSum(it iter.Seq[ast.Constant]) (ast.Constant, error) {
	var sum float64
	for c := range it {
		num, err := c.Float64Value()
		if err != nil {
			return ast.Constant{}, err
		}
		sum += num
	}
	return ast.Float64(sum), nil
}

func evalAvg(it iter.Seq[ast.Constant]) (ast.Constant, error) {
	var sum float64
	var n int
	for c := range it {
		n++
		num, err := c.Float64Value()
		if err != nil {
			fnum, err := c.NumberValue()
			if err != nil {
				return ast.Constant{}, err
			}
			num = float64(fnum)
		}
		sum += num
	}
	if n == 0 {
		return ast.Float64(math.NaN()), nil
	}
	return ast.Float64(sum / float64(n)), nil
}
