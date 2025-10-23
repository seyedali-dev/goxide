// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package chain. chain provides fluent method chaining for Result and Option types.
// This enables a more readable, pipeline-style programming pattern similar to Rust's method chaining.
//
// Benefits:
//   - Fluent API: Chain multiple operations in a readable sequence
//   - Type safety: Compiler tracks type transformations through the chain
//   - No nesting: Avoids deep nesting of Map/AndThen calls
//   - Better readability: Operations read left-to-right like a pipeline
//
// Example - Traditional vs Chained:
//
//	// Traditional nested style
//	result := result.AndThen(
//	    result.Map(
//	        findUser(123),
//	        func(u User) string { return u.Name },
//	    ),
//	    func(name string) Result[Profile] { return findProfile(name) },
//	)
//
//	// Fluent chained style
//	result := chain.Chain(findUser(123)).
//	    Map(func(u User) string { return u.Name }).
//	    AndThen(findProfile).
//	    Unwrap()
package chain

import (
	"github.com/seyedali-dev/goxide/rusty/result"
)

// -------------------------------------------- Result Chaining --------------------------------------------

// ApplyToResult [Out, In] is the first step in a Result chaining pipeline.
// It holds a Result[In] and provides methods that transform it to Result[Out].
type ApplyToResult[Out, In any] struct {
	result result.Result[In]
}

// Chain starts a new chaining pipeline with a Result[In].
// Use this as the entry point for fluent Result operations.
//
// Example:
//
//	 userResult := findUser(123)
//		chain.Chain(userResult).
//		    Map(func(u User) string { return u.Name }).
//		    AndThen(validateName).
//		    Unwrap()
func Chain[Out, T any](r result.Result[T]) *ApplyToResult[Out, T] {
	return &ApplyToResult[Out, T]{result: r}
}

// Map transforms the value inside the Result using fn.
// Returns a new ApplyToResult that can continue the chain.
//
// Example:
//
//	chain.Chain(Ok(42)).
//	    Map(func(x int) string { return fmt.Sprintf("value: %d", x) }).
//	    Unwrap() // Result[string] with "value: 42"
func (applyToResult *ApplyToResult[Out, In]) Map(fn func(In) Out) result.Result[Out] {
	return result.Map(applyToResult.result, fn)
}

// AndThen chains a Result-returning function.
// Similar to Map but for functions that can fail.
//
// Example:
//
//	chain.Chain(validateEmail("test@example.com")).
//	    AndThen(func(email string) result.Result[User] {
//	        return createUser(email)
//	    }).
//	    Unwrap()
func (applyToResult *ApplyToResult[Out, In]) AndThen(fn func(In) result.Result[Out]) result.Result[Out] {
	return result.AndThen(applyToResult.result, fn)
}

// MapError transforms the error if the Result is in error state.
//
// Example:
//
//	chain.Chain(dbQueryResult).
//	    MapError(func(err error) error {
//	        return fmt.Errorf("database error: %w", err)
//	    }).
//	    Unwrap()
func (applyToResult *ApplyToResult[Out, In]) MapError(fn func(error) error) *ApplyToResult[Out, In] {
	return &ApplyToResult[Out, In]{
		result: applyToResult.result.MapError(fn),
	}
}

// Unwrap terminates the chain and returns the final Result.
// This is usually the last call in a chain.
func (applyToResult *ApplyToResult[Out, In]) Unwrap() result.Result[Out] {
	// We need to handle the case where Out != In (after transformations)
	// This is a type-safe way to extract the final result
	if applyToResult.result.IsErr() {
		return result.Err[Out](applyToResult.result.Err())
	}

	// If we're at the end of a chain where types match, return directly
	if out, ok := any(applyToResult.result).(result.Result[Out]); ok {
		return out
	}

	// This should never happen with proper type tracking
	panic("type mismatch in chain unwrap")
}

// OrElse terminates the chain and returns the value or fallback.
func (applyToResult *ApplyToResult[Out, In]) OrElse(fallback Out) Out {
	return applyToResult.Unwrap().UnwrapOr(fallback)
}

// OrElseGet terminates the chain and returns the value or computed fallback.
func (applyToResult *ApplyToResult[Out, In]) OrElseGet(fn func(error) Out) Out {
	return applyToResult.Unwrap().UnwrapOrElse(fn)
}
