// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package types. map provides small, generic functional helpers inspired by
// functional programming patterns. These utilities are designed to make
// higher‑order function composition and default handling more expressive.
//
// Avoid them when plain inline functions are clearer — Go favors explicitness,
// so these helpers are most useful in generic libraries or when building
// composable abstractions.
package types

// ------------------------------------- Types -------------------------------------

// Id returns its input unchanged.
// Useful as a default function or placeholder in higher-order contexts.
//
// Example:
//
//	x := types.Id(42) // x == 42
func Id[T any](t T) T {
	return t
}

// Return creates a function that ignores its input and always returns t.
// Use when you need a constant function of type func(In) T.
//
// Example:
//
//	f := types.Return
//	fmt.Println(f("ignored")) // 5
func Return[In any, T any](t T) func(In) T {
	return func(_ In) T {
		return t
	}
}

// Return0 creates a zero-argument function that always returns t.
// Useful for lazy defaults or callbacks that require func() T.
//
// Example:
//
//	f := types.Return0(10)
//	fmt.Println(f()) // 10
func Return0[T any](t T) func() T {
	return func() T {
		return t
	}
}

// Value returns the zero value of type T.
// Use when you need an explicit zero value without constructing it manually.
//
// Example:
//
//	var s string = types.Value[string]() // s == ""
func Value[T any]() (t T) {
	return t
}

// Compose chains two functions: fn1 followed by fn2.
// Returns a new function that applies fn1, then fn2.
//
// Example:
//
//	f := types.Compose(strings.TrimSpace, strings.ToUpper)
//	fmt.Println(f("  hi ")) // "HI"
func Compose[T, U, V any](fn1 func(T) U, fn2 func(U) V) func(T) V {
	return func(t T) V {
		return fn2(fn1(t))
	}
}
