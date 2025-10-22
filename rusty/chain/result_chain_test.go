// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package chain_test. chain_test demonstrates fluent method chaining for Result and Option types.
package chain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/seyedali-dev/gopherbox/rusty/chain"
	"github.com/seyedali-dev/gopherbox/rusty/result"
)

// -------------------------------------------- Test Data --------------------------------------------

type User struct {
	ID    int
	Name  string
	Email string
}

type Profile struct {
	UserID   int
	Bio      string
	Settings string
}

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidEmail = errors.New("invalid email")
	ErrDBConnection = errors.New("database connection failed")
)

// -------------------------------------------- Additional Test Helpers --------------------------------------------

func double(x int) int { return x * 2 }

func intToString(x int) string { return fmt.Sprintf("num: %d", x) }

func failOnOdd(x int) result.Result[string] {
	if x%2 == 1 {
		return result.Err[string](errors.New("odd number not allowed"))
	}
	return result.Ok(fmt.Sprintf("even: %d", x))
}

// -------------------------------------------- Result Chain Tests --------------------------------------------

func TestResultChain_Map(t *testing.T) {
	// Traditional style
	traditional := result.Map(
		result.Ok(42),
		func(x int) string { return fmt.Sprintf("value: %d", x) },
	)

	// Fluent chain style
	chained := chain.Chain[string, int](result.Ok(42)).
		Map(func(x int) string { return fmt.Sprintf("value: %d", x) }).
		Unwrap()

	if traditional.Unwrap() != chained {
		t.Fatalf("expected %q, got %q", traditional.Unwrap(), chained)
	}
}

func TestResultChain_Map_ErrorPropagation(t *testing.T) {
	chained := chain.Chain[string, int](result.Err[int](ErrDBConnection)).
		Map(intToString)

	if chained.IsOk() {
		t.Fatal("expected error to propagate through Map")
	}
	if !errors.Is(chained.Err(), ErrDBConnection) {
		t.Errorf("expected %v, got %v", ErrDBConnection, chained.Err())
	}
}

func TestResultChain_Map_ErrorPropagation2(t *testing.T) {
	failingOp := func() result.Result[int] {
		return result.Err[int](ErrDBConnection)
	}

	transform := func(x int) string {
		return fmt.Sprintf("value: %d", x)
	}

	// The chain should short-circuit on error
	chainResult := chain.Chain[string, int](failingOp()).
		Map(transform)

	if chainResult.IsOk() {
		t.Fatal("expected error to propagate through chain")
	}
	if !errors.Is(chainResult.Err(), ErrDBConnection) {
		t.Fatalf("expected %v, got %v", ErrDBConnection, chainResult.Err())
	}
}

func TestResultChain_AndThen(t *testing.T) {
	input := result.Ok(4)

	chained := chain.Chain[string, int](input).
		AndThen(failOnOdd)

	if !chained.IsOk() {
		t.Fatalf("expected Ok, got Err: %v", chained.Err())
	}
	if chained.Unwrap() != "even: 4" {
		t.Errorf("expected 'even: 4', got %q", chained.Unwrap())
	}
}

func TestResultChain_AndThen2(t *testing.T) {
	findUser := func(id int) result.Result[User] {
		if id == 123 {
			return result.Ok(User{ID: 123, Name: "John", Email: "john@example.com"})
		}
		return result.Err[User](ErrUserNotFound)
	}

	getProfile := func(user User) result.Result[Profile] {
		return result.Ok(Profile{UserID: user.ID, Bio: "Software developer", Settings: "default"})
	}

	// Traditional nested style
	traditional := result.AndThen(
		findUser(123),
		func(user User) result.Result[Profile] {
			return getProfile(user)
		},
	)

	// Fluent chain style
	chained := chain.Chain[Profile, User](findUser(123)).
		AndThen(getProfile).
		Unwrap()

	if traditional.Unwrap().UserID != chained.UserID {
		t.Fatalf("expected user ID %d, got %d", traditional.Unwrap().UserID, chained.UserID)
	}
}

func TestResultChain_AndThen3_ErrorFromFunction(t *testing.T) {
	input := result.Ok(3) // odd → will fail in failOnOdd

	chained := chain.Chain[string, int](input).
		AndThen(failOnOdd)

	if chained.IsOk() {
		t.Fatal("expected error from AndThen function")
	}
	if chained.Err().Error() != "odd number not allowed" {
		t.Errorf("unexpected error: %v", chained.Err())
	}
}

func TestResultChain_AndThen4_ErrorPropagation(t *testing.T) {
	input := result.Err[int](ErrUserNotFound)

	chained := chain.Chain[string, int](input).
		AndThen(failOnOdd)

	if chained.IsOk() {
		t.Fatal("expected initial error to propagate")
	}
	if !errors.Is(chained.Err(), ErrUserNotFound) {
		t.Errorf("expected %v, got %v", ErrUserNotFound, chained.Err())
	}
}

func TestResultChain_MapError(t *testing.T) {
	chainResult := chain.Chain[string, int](result.Err[int](ErrDBConnection)).
		MapError(func(err error) error {
			// Did some operation and that operation failed as well
			return fmt.Errorf("wrapped: %w", err)
		}).
		Unwrap()

	if chainResult.IsOk() {
		t.Fatal("expected error")
	}
	if chainResult.Err().Error() != "wrapped: database connection failed" {
		t.Fatalf("expected wrapped error, got %v", chainResult.Err())
	}
}

func TestResultChain_MapError4_TransformsError(t *testing.T) {
	chained := chain.Chain[string, int](result.Err[int](ErrInvalidEmail)).
		MapError(func(err error) error { return fmt.Errorf("validation: %w", err) }).
		Unwrap()

	if chained.IsOk() {
		t.Fatal("expected error")
	}
	expectedMsg := "validation: invalid email"
	if chained.Err().Error() != expectedMsg {
		t.Errorf("expected %q, got %q", expectedMsg, chained.Err().Error())
	}
}
