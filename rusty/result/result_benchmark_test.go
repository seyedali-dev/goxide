// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

package result_test

import (
	"errors"
	"testing"

	"github.com/seyedali-dev/goxide/rusty/chain"
	"github.com/seyedali-dev/goxide/rusty/result"
)

// Traditional Go error handling functions for comparison
func traditionalSuccess() (int, error) {
	return 42, nil
}

func traditionalError() (int, error) {
	return 0, errors.New("error occurred")
}

func traditionalChainedSuccess(val int) (int, error) {
	return val * 2, nil
}

func traditionalChainedError(val int) (int, error) {
	return 0, errors.New("chain error")
}

// Result-based equivalents
func resultSuccess() result.Result[int] {
	return result.Ok(42)
}

func resultError() result.Result[int] {
	return result.Err[int](errors.New("error occurred"))
}

func resultChainedSuccess(val int) result.Result[int] {
	return result.Ok(val * 2)
}

func resultChainedError(val int) result.Result[int] {
	return result.Err[int](errors.New("chain error"))
}

// Benchmark: Simple Traditional Success Case
//
//	BenchmarkTraditionalSuccess    	1000000000	         0.2457 ns/op	       0 B/op	       0 allocs/op
func BenchmarkTraditionalSuccess(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		val, err := traditionalSuccess()
		if err != nil {
			b.Fatal("unexpected error")
		}
		if val != 42 {
			b.Fatal("unexpected value")
		}
	}
}

// Benchmark: Simple Result Success Case
//
//	BenchmarkResultSuccess    	1000000000	         0.2448 ns/op	       0 B/op	       0 allocs/op
func BenchmarkResultSuccess(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res := resultSuccess()
		if res.IsErr() {
			b.Fatal("unexpected error")
		}
		val := res.Unwrap()
		if val != 42 {
			b.Fatal("unexpected value")
		}
	}
}

// BenchmarkResultSuccessUnwrapOr    	88363804	        12.75 ns/op	       8 B/op	       1 allocs/op
func BenchmarkResultSuccessUnwrapOr(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res := resultSuccess()
		val := res.UnwrapOr(0)
		if val != 42 {
			b.Fatal("unexpected value")
		}
	}
}

// Benchmark: Simple Traditional Error Case
//
//	BenchmarkTraditionalError    	1000000000	         0.2427 ns/op	       0 B/op	       0 allocs/op
func BenchmarkTraditionalError(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		val, err := traditionalError()
		if err == nil {
			b.Fatal("expected error")
		}
		if val != 0 {
			b.Fatal("unexpected value")
		}
	}
}

// Benchmark: Simple Result Error Case
//
//	BenchmarkResultError    	56643303	        20.12 ns/op	      16 B/op	       1 allocs/op
func BenchmarkResultError(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res := resultError()
		if res.IsOk() {
			b.Fatal("expected error")
		}
		if res.UnwrapOr(0) != 0 {
			b.Fatal("unexpected value")
		}
	}
}

// Benchmark: Chained Traditional Operations (Success Path)
//
//	BenchmarkTraditionalChainedSuccess    	1000000000	         0.2458 ns/op	       0 B/op	       0 allocs/op
func BenchmarkTraditionalChainedSuccess(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		val1, err := traditionalSuccess()
		if err != nil {
			b.Fatal("unexpected error")
		}

		val2, err := traditionalChainedSuccess(val1)
		if err != nil {
			b.Fatal("unexpected error")
		}

		val3, err := traditionalChainedSuccess(val2)
		if err != nil {
			b.Fatal("unexpected error")
		}

		if val3 != 168 { // 42 * 2 * 2
			b.Fatal("unexpected value")
		}
	}
}

// Benchmark: Chained Result Operations (Success Path)
//
//	BenchmarkResultChainedSuccess-12    	33059582	        34.23 ns/op	      24 B/op	       3 allocs/op
func BenchmarkResultChainedSuccess(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {

		res := chain.Chain2[int, int](resultSuccess()).
			AndThen(resultChainedSuccess).
			AndThen(resultChainedSuccess)

		if res.IsErr() {
			b.Fatal("unexpected error")
		}
		if res.Unwrap() != 168 {
			b.Fatal("unexpected value")
		}
	}
}

//
//func BenchmarkResultChainedSuccessMap(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		res := resultSuccess().
//			Map(func(x int) int { return x * 2 }).
//			Map(func(x int) int { return x * 2 })
//
//		if res.IsErr() {
//			b.Fatal("unexpected error")
//		}
//		if res.Unwrap() != 168 {
//			b.Fatal("unexpected value")
//		}
//	}
//}
//
//// Benchmark: Chained Operations (Error Path)
//func BenchmarkTraditionalChainedError(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		val1, err := traditionalSuccess()
//		if err != nil {
//			b.Fatal("unexpected error")
//		}
//
//		_, err = traditionalChainedError(val1)
//		if err == nil {
//			b.Fatal("expected error")
//		}
//		// Error occurred, no further processing
//	}
//}
//
//func BenchmarkResultChainedError(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		res := resultSuccess().
//			AndThen(resultChainedError).
//			AndThen(resultChainedSuccess) // This won't execute due to error
//
//		if !res.IsErr() {
//			b.Fatal("expected error")
//		}
//	}
//}
//
//// Benchmark: BubbleUp with Catch (Success Path)
//func BenchmarkResultBubbleUpSuccess(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		var res result.Result[int]
//		func() {
//			defer result.Catch(&res)
//			val1 := resultSuccess().BubbleUp()
//			val2 := resultChainedSuccess(val1).BubbleUp()
//			val3 := resultChainedSuccess(val2).BubbleUp()
//			res = result.Ok(val3)
//		}()
//
//		if res.IsErr() {
//			b.Fatal("unexpected error")
//		}
//		if res.Unwrap() != 168 {
//			b.Fatal("unexpected value")
//		}
//	}
//}
//
//// Benchmark: BubbleUp with Catch (Error Path)
//func BenchmarkResultBubbleUpError(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		var res result.Result[int]
//		func() {
//			defer result.Catch(&res)
//			val1 := resultSuccess().BubbleUp()
//			_ = resultChainedError(val1).BubbleUp() // This will panic and be caught
//			// Execution won't reach here
//			res = result.Ok(0)
//		}()
//
//		if !res.IsErr() {
//			b.Fatal("expected error")
//		}
//	}
//}
//
//// Benchmark: MapError
//func BenchmarkTraditionalMapError(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		_, err := traditionalError()
//		if err != nil {
//			// Traditional way of mapping errors
//			err = errors.New("wrapped: " + err.Error())
//		}
//		if err == nil {
//			b.Fatal("expected error")
//		}
//	}
//}
//
//func BenchmarkResultMapError(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		res := resultError().
//			MapError(func(err error) error {
//				return errors.New("wrapped: " + err.Error())
//			})
//
//		if !res.IsErr() {
//			b.Fatal("expected error")
//		}
//	}
//}
//
//// Benchmark: UnwrapOr with default value
//func BenchmarkTraditionalUnwrapOr(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		val, err := traditionalError()
//		resultVal := 0
//		if err != nil {
//			resultVal = 100 // default
//		} else {
//			resultVal = val
//		}
//		if resultVal != 100 {
//			b.Fatal("unexpected value")
//		}
//	}
//}
//
//func BenchmarkResultUnwrapOr(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		res := resultError()
//		val := res.UnwrapOr(100)
//		if val != 100 {
//			b.Fatal("unexpected value")
//		}
//	}
//}
//
//// Benchmark: Multiple value combination (Map2, Map3)
//func BenchmarkTraditionalMultiValue(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		val1, err1 := traditionalSuccess()
//		if err1 != nil {
//			b.Fatal("unexpected error")
//		}
//
//		val2, err2 := traditionalSuccess()
//		if err2 != nil {
//			b.Fatal("unexpected error")
//		}
//
//		result := val1 + val2
//		if result != 84 {
//			b.Fatal("unexpected value")
//		}
//	}
//}
//
//func BenchmarkResultMap2(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		res1 := resultSuccess()
//		res2 := resultSuccess()
//		res := result.Map2(res1, res2, func(a, b int) int {
//			return a + b
//		})
//
//		if res.IsErr() {
//			b.Fatal("unexpected error")
//		}
//		if res.Unwrap() != 84 {
//			b.Fatal("unexpected value")
//		}
//	}
//}
//
//// Benchmark: Wrapping traditional functions
//func BenchmarkTraditionalWrap(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		val, err := traditionalSuccess()
//		if err != nil {
//			b.Fatal("unexpected error")
//		}
//		if val != 42 {
//			b.Fatal("unexpected value")
//		}
//	}
//}
//
//func BenchmarkResultWrap(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		res := result.Wrap(traditionalSuccess())
//		if res.IsErr() {
//			b.Fatal("unexpected error")
//		}
//		if res.Unwrap() != 42 {
//			b.Fatal("unexpected value")
//		}
//	}
//}
//
//// Benchmark: Option value access
//func BenchmarkResultOptionValue(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		res := resultSuccess()
//		opt := res.Value()
//		if opt.IsNone() {
//			b.Fatal("expected some value")
//		}
//		val := opt.Unwrap()
//		if val != 42 {
//			b.Fatal("unexpected value")
//		}
//	}
//}
//
//// Benchmark: Error checking overhead
//func BenchmarkTraditionalErrorCheck(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		_, err := traditionalSuccess()
//		if err != nil {
//			b.Fatal("unexpected error")
//		}
//	}
//}
//
//func BenchmarkResultErrorCheck(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		res := resultSuccess()
//		if res.IsErr() {
//			b.Fatal("unexpected error")
//		}
//	}
//}
