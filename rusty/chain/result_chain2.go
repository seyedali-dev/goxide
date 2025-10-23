// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package chain. chain2 provides fluent method chaining for Result and Option types.
// This enables a more readable, pipeline-style programming pattern similar to Rust's method chaining.
//
// Benefits:
//   - Fluent API: Chain multiple operations in a readable sequence
//   - Type safety: Compiler tracks type transformations through the chain
//   - No nesting: Avoids deep nesting of Map/AndThen calls
//   - Better readability: Operations read left-to-right like a pipeline
package chain

import (
	"github.com/seyedali-dev/goxide/rusty/result"
)

// -------------------------------------------- Multi-Step Result Chaining --------------------------------------------

// ApplyToResult2 [Out1, Out2, In] represents a 2-step transformation pipeline.
type ApplyToResult2[Out1, Out2, In any] struct {
	result result.Result[In]
}

// Chain2 starts a chain that expects exactly 2 transformations.
// Useful when you know the exact number of steps for type clarity.
func Chain2[Out2, Out1, T any](result result.Result[T]) *ApplyToResult2[Out1, Out2, T] {
	return &ApplyToResult2[Out1, Out2, T]{
		result: result,
	}
}

func (applyToResult2 ApplyToResult2[Out1, Out2, T]) AndThen(fn func(T) result.Result[Out1]) *ApplyToResult[Out2, Out1] {
	return Chain[Out2](result.AndThen(applyToResult2.result, fn))
}

func (applyToResult2 ApplyToResult2[Out1, Out2, T]) Map(fn func(T) Out1) *ApplyToResult[Out2, Out1] {
	return Chain[Out2](result.Map(applyToResult2.result, fn))
}
