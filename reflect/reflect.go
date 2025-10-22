// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package reflect. reflect provides general reflection utilities.
package reflect

import (
	"fmt"
	"reflect"
)

// ------------------------------------- Public functions -------------------------------------

// IsEqual compares two values of potentially different types.
// - If both values are directly comparable (==), it uses that.
// - Otherwise, it falls back to reflect.DeepEqual.
//
// Example:
//
//	IsEqual(5, 5)               	// true
//	IsEqual("go", "Go")         	// false
//	IsEqual([]int{1,2}, []int{1,2}) // true
func IsEqual[T, V any](v1 T, v2 V) bool {
	rv1 := reflect.ValueOf(v1)
	rv2 := reflect.ValueOf(v2)

	if rv1.Kind() != rv2.Kind() {
		return false
	}
	return rv1.Equal(rv2)
}

// IsEmpty reports whether v is the zero value for its type.
// It supports:
//   - nil pointers/interfaces
//   - empty strings, slices, maps, arrays
//   - zero numbers (int, uint, float, complex)
//   - false booleans
//   - invalid reflect values
func IsEmpty[T any](v T) bool {
	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return rv.IsNil()
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return rv.Complex() == 0
	case reflect.Bool:
		return !rv.Bool()
	default:
		return false
	}
}

// InferType attempts to convert a generic interface{} value into the specified type T.
//
// Behavior:
//   - If intrfc is nil, it returns the zero value of T and an error.
//   - If intrfc is already of type T, it returns the value directly.
//   - Otherwise, it falls back to convertType[T](intrfc) for custom conversion logic.
//
// This is useful when you have a value of unknown dynamic type (e.g., from JSON decoding,
// database queries, or generic containers) and want to safely infer and convert it to a
// specific type.
//
// Usage:
//
//	// Example 1: Conversion from compatible type
//	var x interface{} = "123"
//	num, err := InferType[int](x)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(num) // Output: 123 (assuming convertType handles string→int)
//
//	// Example 2: Nil input
//	var y interface{} = nil
//	_, err = InferType[string](y)
//	if err != nil {
//	    fmt.Println("Error:", err) // Output: Error: cannot infer type from nil interface
//	}
func InferType[T any](intrfc interface{}) (T, error) {
	if intrfc == nil {
		var nothing T
		return nothing, fmt.Errorf("cannot infer type from nil interface")
	}

	// Try direct type assertion
	if val, ok := intrfc.(T); ok {
		return val, nil
	}

	// custom types -> use reflect
	return convertType[T](intrfc)
}

// InferTypeWithPanic is the panic version of InferType on error.
func InferTypeWithPanic[T any](structType any) T {
	val, err := InferType[T](structType)
	if err != nil {
		panic(fmt.Errorf("failed to infer type: %w", err))
	}
	return val
}

// ------------------------------------- Private Helper functions -------------------------------------

// convertType handles type conversions using reflection for more complex cases.
func convertType[T any](intrfc interface{}) (T, error) {
	var zero T
	expectedType := reflect.TypeOf(zero)
	actualValue := reflect.ValueOf(intrfc)

	if !actualValue.IsValid() {
		return zero, fmt.Errorf("invalid value for type conversion")
	}

	isConvertible := actualValue.Type().ConvertibleTo(expectedType)
	if isConvertible {
		return actualValue.
			Convert(expectedType).
			Interface().(T), nil
	}

	return zero, fmt.Errorf("cannot infer type \"%T\" from interface (expected %v) where actual type is \"%v\"", intrfc, expectedType, actualValue.Type())
}
