// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package result_test demonstrates the enhanced Result type with Try/Catch functionality.
//
// Tests done on:
//
//	goos: linux
//	goarch: amd64
//	pkg: github.com/seyedali-dev/goxide/rusty/result
//	cpu: 11th Gen Intel(R) Core(TM) i5-11400H @ 2.70GHz
package result_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/seyedali-dev/goxide/rusty/result"
)

// -------------------------------------------- Error Definitions --------------------------------------------

var (
	ErrDatabaseDown    = errors.New("database connection failed")
	ErrCacheMiss       = errors.New("cache miss")
	ErrNotFound        = errors.New("resource not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrTimeout         = errors.New("operation timeout")
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrConfigMissing   = errors.New("configuration missing")
	ErrDivideByZero    = errors.New("divide by zero")
	ErrMemoryNotFound  = errors.New("memory not found")
	ErrRemoteAPIFailed = errors.New("remote API failed")
)

// -------------------------------------------- Helper Functions --------------------------------------------

func divide(x, y int) (int, error) {
	if y == 0 {
		return 0, ErrDivideByZero
	}
	return x / y, nil
}

func findInMemory(id int) (string, error) {
	return "", ErrMemoryNotFound
}

func findInDatabase(id int) (string, error) {
	return "", ErrDatabaseDown
}

func findInRemoteAPI(id int) (string, error) {
	return fmt.Sprintf("user-%d", id), nil
}

// -------------------------------------------- Test Cases: Basic Try/Catch --------------------------------------------

func TestTry_Success(t *testing.T) {
	compute := func() (res result.Result[int]) {
		defer result.Catch(&res)

		val := result.Wrap(divide(10, 2)).BubbleUp()
		return result.Ok(val * 2)
	}

	res := compute()
	if res.IsErr() {
		t.Fatalf("expected Ok, got Err: %v", res.Err())
	}
	if res.Unwrap() != 10 {
		t.Fatalf("expected 10, got %d", res.Unwrap())
	}
}

func TestTry_Error(t *testing.T) {
	compute := func() (res result.Result[int]) {
		defer result.Catch(&res)
		val := result.Wrap(divide(10, 0)).BubbleUp()
		return result.Ok(val * 2)
	}

	res := compute()
	if res.IsOk() {
		t.Fatal("expected Err, got Ok")
	}
	if !errors.Is(res.Err(), ErrDivideByZero) {
		t.Fatalf("expected ErrDivideByZero, got %v", res.Err())
	}
}

func TestTry_MultipleOperations(t *testing.T) {
	compute := func() (res result.Result[int]) {
		defer result.Catch(&res)
		val1 := result.Wrap(divide(100, 2)).BubbleUp()  // 50
		val2 := result.Wrap(divide(val1, 5)).BubbleUp() // 10
		val3 := result.Wrap(divide(val2, 2)).BubbleUp() // 5
		return result.Ok(val3)
	}

	res := compute()
	if res.IsErr() {
		t.Fatalf("expected Ok, got Err: %v", res.Err())
	}
	if res.Unwrap() != 5 {
		t.Fatalf("expected 5, got %d", res.Unwrap())
	}
}

func TestTry_EarlyReturn(t *testing.T) {
	compute := func() (res result.Result[int]) {
		defer result.Catch(&res)
		val1 := result.Wrap(divide(100, 2)).BubbleUp()  // 50
		val2 := result.Wrap(divide(val1, 0)).BubbleUp() // Error here - early return
		val3 := result.Wrap(divide(val2, 2)).BubbleUp() // Never reached
		return result.Ok(val3)
	}

	res := compute()
	if res.IsOk() {
		t.Fatal("expected Err, got Ok")
	}
	if !errors.Is(res.Err(), ErrDivideByZero) {
		t.Fatalf("expected ErrDivideByZero, got %v", res.Err())
	}
}

// -------------------------------------------- Test Cases: CatchWith --------------------------------------------

func TestCatchWith_SpecificError(t *testing.T) {
	fetchData := func() (res result.Result[string]) {
		defer result.Catch(&res)
		defer result.CatchWith(&res, func(err error) string {
			return "cached-fallback"
		}, ErrDatabaseDown)

		return result.Wrap(findInDatabase(123))
	}

	res := fetchData()
	if res.IsErr() {
		t.Fatalf("expected Ok with fallback, got Err: %v", res.Err())
	}
	if res.Unwrap() != "cached-fallback" {
		t.Fatalf("expected 'cached-fallback', got %s", res.Unwrap())
	}
}

func TestCatchWith_ChainedFallbacks(t *testing.T) {
	fetchData := func() (res result.Result[string]) {
		defer result.Catch(&res)
		// Try remote API if database fails
		defer result.CatchWith(&res, func(err error) string {
			return result.Wrap(findInRemoteAPI(123)).BubbleUp()
		}, ErrDatabaseDown)
		// Try database if memory fails
		defer result.CatchWith(&res, func(err error) string {
			return result.Wrap(findInDatabase(123)).BubbleUp()
		}, ErrMemoryNotFound)

		return result.Wrap(findInMemory(123))
	}

	res := fetchData()
	if res.IsErr() {
		t.Fatalf("expected Ok, got Err: %v", res.Err())
	}
	if res.Unwrap() != "user-123" {
		t.Fatalf("expected 'user-123', got %s", res.Unwrap())
	}
}

func TestCatchWith_MultipleErrors(t *testing.T) {
	fetchData := func() (res result.Result[string]) {
		defer result.Catch(&res)
		defer result.CatchWith(&res, func(err error) string {
			return "default-fallback"
		}, ErrMemoryNotFound, ErrDatabaseDown, ErrCacheMiss)

		return result.Wrap(findInDatabase(123))
	}

	res := fetchData()
	if res.IsErr() {
		t.Fatalf("expected Ok with fallback, got Err: %v", res.Err())
	}
	if res.Unwrap() != "default-fallback" {
		t.Fatalf("expected 'default-fallback', got %s", res.Unwrap())
	}
}

func TestCatchWith_AnyError(t *testing.T) {
	fetchData := func() (res result.Result[int]) {
		defer result.Catch(&res)
		defer result.CatchWith(&res, func(err error) int {
			return -1 // Default value for any error
		}) // No specific errors = catches all

		return result.Err[int](ErrTimeout)
	}

	res := fetchData()
	if res.IsErr() {
		t.Fatalf("expected Ok with fallback, got Err: %v", res.Err())
	}
	if res.Unwrap() != -1 {
		t.Fatalf("expected -1, got %d", res.Unwrap())
	}
}

func TestCatchWith_NoMatch(t *testing.T) {
	fetchData := func() (res result.Result[string]) {
		defer result.Catch(&res)
		defer result.CatchWith(&res, func(err error) string {
			return "cached-fallback"
		}, ErrCacheMiss) // Only handles cache miss

		return result.Wrap(findInDatabase(123)) // Returns ErrDatabaseDown
	}

	res := fetchData()
	if res.IsOk() {
		t.Fatal("expected Err to propagate, got Ok")
	}
	if !errors.Is(res.Err(), ErrDatabaseDown) {
		t.Fatalf("expected ErrDatabaseDown, got %v", res.Err())
	}
}

// -------------------------------------------- Test Cases: Fallback --------------------------------------------

func TestFallback_Success(t *testing.T) {
	getConfig := func() (res result.Result[int]) {
		defer result.Catch(&res)
		defer result.Fallback(&res, 30, ErrConfigMissing)

		return result.Err[int](ErrConfigMissing)
	}

	res := getConfig()
	if res.IsErr() {
		t.Fatalf("expected Ok with fallback, got Err: %v", res.Err())
	}
	if res.Unwrap() != 30 {
		t.Fatalf("expected 30, got %d", res.Unwrap())
	}
}

func TestFallback_MultipleErrors(t *testing.T) {
	getTimeout := func() (res result.Result[int]) {
		defer result.Catch(&res)
		defer result.Fallback(&res, 60, ErrConfigMissing, ErrTimeout, ErrInvalidInput)

		return result.Err[int](ErrTimeout)
	}

	res := getTimeout()
	if res.IsErr() {
		t.Fatalf("expected Ok with fallback, got Err: %v", res.Err())
	}
	if res.Unwrap() != 60 {
		t.Fatalf("expected 60, got %d", res.Unwrap())
	}
}

func TestFallback_AnyError(t *testing.T) {
	getValue := func() (res result.Result[bool]) {
		defer result.Catch(&res)
		defer result.Fallback(&res, false) // No specific errors = handles all

		return result.Err[bool](errors.New("some random error"))
	}

	res := getValue()
	if res.IsErr() {
		t.Fatalf("expected Ok with fallback, got Err: %v", res.Err())
	}
	if res.Unwrap() != false {
		t.Fatalf("expected false, got %v", res.Unwrap())
	}
}

func TestFallback_NoMatch(t *testing.T) {
	getConfig := func() (res result.Result[int]) {
		defer result.Catch(&res)
		defer result.Fallback(&res, 30, ErrConfigMissing) // Only handles config missing

		return result.Err[int](ErrTimeout) // Different error
	}

	res := getConfig()
	if res.IsOk() {
		t.Fatal("expected Err to propagate, got Ok")
	}
	if !errors.Is(res.Err(), ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", res.Err())
	}
}

// -------------------------------------------- Test Cases: CatchErr --------------------------------------------

func TestCatchErr_Success(t *testing.T) {
	compute := func() (val int, err error) {
		defer result.CatchErr(&val, &err)
		result1 := result.Wrap(divide(10, 2)).BubbleUp()
		return result1 * 2, nil
	}

	val, err := compute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != 10 {
		t.Fatalf("expected 10, got %d", val)
	}
}

func TestCatchErr_Error(t *testing.T) {
	compute := func() (val int, err error) {
		defer result.CatchErr(&val, &err)

		result1 := result.Wrap(divide(10, 0)).BubbleUp()
		return result1 * 2, nil
	}

	val, err := compute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrDivideByZero) {
		t.Fatalf("expected ErrDivideByZero, got %v", err)
	}
	if val != 0 {
		t.Fatalf("expected 0 for error case, got %d", val)
	}
}

// -------------------------------------------- Test Cases: Real-World Examples --------------------------------------------

func TestRealWorld_FileProcessing(t *testing.T) {
	processFile := func(filename string) (res result.Result[[]byte]) {
		defer result.Catch(&res)
		defer result.Fallback(&res, []byte("default content"), os.ErrNotExist)

		file := result.Wrap(os.Open(filename)).BubbleUp()
		defer file.Close()

		buffer := make([]byte, 100)
		n := result.Wrap(file.Read(buffer)).BubbleUp()
		return result.Ok(buffer[:n])
	}

	// Test with non-existent file
	res := processFile("non-existent-file.txt")
	if res.IsErr() {
		t.Fatalf("expected fallback, got Err: %v", res.Err())
	}
	if string(res.Unwrap()) != "default content" {
		t.Fatalf("expected 'default content', got %s", string(res.Unwrap()))
	}
}

func TestRealWorld_DatabaseWithCacheFallback(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	findUser := func(id int) (res result.Result[User]) {
		defer result.Catch(&res)
		// Try cache if database fails
		defer result.CatchWith(&res, func(err error) User {
			// Simulate cache lookup
			return User{ID: id, Name: "cached-user"}
		}, ErrDatabaseDown)

		// Simulate database lookup that fails
		return result.Err[User](ErrDatabaseDown)
	}

	res := findUser(123)
	if res.IsErr() {
		t.Fatalf("expected cached user, got Err: %v", res.Err())
	}
	user := res.Unwrap()
	if user.Name != "cached-user" {
		t.Fatalf("expected 'cached-user', got %s", user.Name)
	}
}

func TestRealWorld_OrderProcessingPipeline(t *testing.T) {
	type Order struct {
		ID     int
		Amount float64
	}

	type Payment struct {
		OrderID int
		Paid    bool
	}

	type Receipt struct {
		PaymentID int
		Total     float64
	}

	validateOrder := func(order Order) result.Result[Order] {
		if order.Amount <= 0 {
			return result.Err[Order](ErrInvalidInput)
		}
		return result.Ok(order)
	}

	processPayment := func(order Order) result.Result[Payment] {
		return result.Ok(Payment{OrderID: order.ID, Paid: true})
	}

	generateReceipt := func(payment Payment) result.Result[Receipt] {
		return result.Ok(Receipt{PaymentID: payment.OrderID, Total: 100.00})
	}

	processOrder := func(order Order) (res result.Result[Receipt]) {
		defer result.Catch(&res)

		validOrder := validateOrder(order).BubbleUp()
		payment := processPayment(validOrder).BubbleUp()
		receipt := generateReceipt(payment).BubbleUp()

		return result.Ok(receipt)
	}

	// Test successful processing
	res := processOrder(Order{ID: 1, Amount: 100.00})
	if res.IsErr() {
		t.Fatalf("expected Ok, got Err: %v", res.Err())
	}
	receipt := res.Unwrap()
	if receipt.Total != 100.00 {
		t.Fatalf("expected 100.00, got %f", receipt.Total)
	}

	// Test validation failure
	res = processOrder(Order{ID: 2, Amount: -10.00})
	if res.IsOk() {
		t.Fatal("expected Err, got Ok")
	}
	if !errors.Is(res.Err(), ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", res.Err())
	}
}

func TestRealWorld_MultiLayerFallback(t *testing.T) {
	getData := func(id int) (res result.Result[string]) {
		defer result.Catch(&res)
		// Final fallback for any unhandled error
		defer result.Fallback(&res, "ultimate-default")
		// Try remote API if database fails
		defer result.CatchWith(&res, func(err error) string {
			return result.Wrap(findInRemoteAPI(id)).BubbleUp()
		}, ErrDatabaseDown)
		// Try database if memory fails
		defer result.CatchWith(&res, func(err error) string {
			return result.Wrap(findInDatabase(id)).BubbleUp()
		}, ErrMemoryNotFound)

		// Try memory first
		return result.Wrap(findInMemory(id))
	}

	res := getData(123)
	if res.IsErr() {
		t.Fatalf("expected Ok, got Err: %v", res.Err())
	}
	// Should cascade through memory -> database -> remote API
	if res.Unwrap() != "user-123" {
		t.Fatalf("expected 'user-123', got %s", res.Unwrap())
	}
}

// -------------------------------------------- Test Cases: Edge Cases --------------------------------------------

func TestEdgeCase_PanicRecovery(t *testing.T) {
	compute := func() (res result.Result[int]) {
		defer result.Catch(&res)

		// Regular panic (not from Try) should be re-raised
		defer func() {
			if r := recover(); r != nil {
				panic(r) // Re-panic if it's something else
			}
		}()

		panic("regular panic")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic to be re-raised")
		}
	}()

	_ = compute()
}

func TestEdgeCase_NestedTryCalls(t *testing.T) {
	inner := func(x, y int) (res result.Result[int]) {
		defer result.Catch(&res)
		return result.Wrap(divide(x, y))
	}

	outer := func() (res result.Result[int]) {
		defer result.Catch(&res)
		val1 := inner(10, 2).BubbleUp()
		val2 := inner(val1, 0).BubbleUp() // This will error
		return result.Ok(val2)
	}

	res := outer()
	if res.IsOk() {
		t.Fatal("expected Err, got Ok")
	}
	if !errors.Is(res.Err(), ErrDivideByZero) {
		t.Fatalf("expected ErrDivideByZero, got %v", res.Err())
	}
}

func TestEdgeCase_CatchWithReThrow(t *testing.T) {
	compute := func() (res result.Result[string]) {
		defer result.Catch(&res)
		// Handler that re-throws a different error
		defer result.CatchWith(&res, func(err error) string {
			// Transform error and re-throw
			result.Err[string](ErrRemoteAPIFailed).BubbleUp()
			return ""
		}, ErrDatabaseDown)

		return result.Err[string](ErrDatabaseDown)
	}

	res := compute()
	if res.IsOk() {
		t.Fatal("expected Err, got Ok")
	}
	// Should have the transformed error
	if !errors.Is(res.Err(), ErrRemoteAPIFailed) {
		t.Fatalf("expected ErrRemoteAPIFailed, got %v", res.Err())
	}
}

// -------------------------------------------- Benchmark Tests --------------------------------------------

// Test result:
//
//	BenchmarkTraditionalErrorHandling    	1000000000	         0.2497 ns/op	       0 B/op
func BenchmarkTraditionalErrorHandling(b *testing.B) {
	compute := func() (int, error) {
		val1, err := divide(100, 2)
		if err != nil {
			return 0, err
		}
		val2, err := divide(val1, 5)
		if err != nil {
			return 0, err
		}
		val3, err := divide(val2, 2)
		if err != nil {
			return 0, err
		}
		return val3, nil
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = compute()
	}
}

// Test result:
//
//	BenchmarkResultWithTry    	25683512	        45.96 ns/op	      32 B/op	       4 allocs/op
func BenchmarkResultWithTry(b *testing.B) {
	compute := func() (res result.Result[int]) {
		defer result.Catch(&res)
		val1 := result.Wrap(divide(100, 2)).BubbleUp()
		val2 := result.Wrap(divide(val1, 5)).BubbleUp()
		val3 := result.Wrap(divide(val2, 2)).BubbleUp()
		return result.Ok(val3)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = compute()
	}
}

// Test result:
//
//	BenchmarkResultWithAndThen    	28387802	        41.85 ns/op	      24 B/op	       3 allocs/op
func BenchmarkResultWithAndThen(b *testing.B) {
	compute := func() result.Result[int] {
		wrappedResult := result.Wrap(divide(100, 2))
		wrappedDivideResult := result.AndThen(wrappedResult, func(v int) result.Result[int] {
			return result.Wrap(divide(v, 5))
		})
		wrappedDivideResult = result.AndThen(wrappedResult, func(v int) result.Result[int] {
			return result.Wrap(divide(v, 2))
		})
		return wrappedDivideResult
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = compute()
	}
}
