// Package chain_test. chain2_test demonstrates fluent method chaining for Result and Option types.
package chain_test

import (
	"testing"

	"github.com/seyedali-dev/gopherbox/rusty/chain"
	"github.com/seyedali-dev/gopherbox/rusty/result"
)

func TestResultChain_MultipleOperations(t *testing.T) {
	validateEmail := func(email string) result.Result[string] {
		if len(email) > 0 {
			return result.Ok(email)
		}
		return result.Err[string](ErrInvalidEmail)
	}

	createUser := func(email string) result.Result[User] {
		return result.Ok(User{ID: 1, Email: email, Name: "Test User"})
	}

	getUserName := func(user User) string {
		return user.Name
	}

	// Complex chain with multiple transformations
	chainResult := chain.Chain2[string, User, string](validateEmail("test@example.com")).
		AndThen(createUser).
		Map(getUserName)

	if chainResult.IsErr() {
		t.Fatalf("expected success, got error: %v", chainResult.Err())
	}
	if chainResult.Unwrap() != "Test User" {
		t.Fatalf("expected %q, got %q", "Test User", chainResult.Unwrap())
	}
}
