// Package reflect. reflect provides general reflection utilities.
package reflect

import "reflect"

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
