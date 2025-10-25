// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package result. result provides a Rust-like Result[T] type for Go to handle operations that can fail.
// It replaces the traditional Go pattern of returning (T, error) with a single Result[T] that is either Ok(value) or Err(error).
//
// Benefits over traditional (T, error) pattern:
//   - Type-safe: Compiler enforces error checking before accessing values
//   - Composable: Chain multiple fallible operations using Map, FlatMap, AndThen
//   - Explicit: Cannot accidentally use a zero value when an error occurred
//   - Functional: Enables railway-oriented programming patterns
//   - Early returns: BubbleUp() enables Rust-like ? operator behavior with deferred error handling
//
// Common use cases:
//   - Database operations that can fail (queries, inserts, updates)
//   - API calls that may return errors
//   - File I/O operations
//   - Validation logic that can fail
//   - Business operations with multiple failure modes
//
// Example - Traditional Go vs Result:
//
//	// Traditional Go
//	user, err := repo.FindByID(123)
//	if err != nil {
//	    return nil, err
//	}
//	// Easy to forget error check and use zero value!
//
//	// With Result
//	userResult := repo.FindByID(123)
//	// Cannot access value without explicitly handling error case
//
// Example - Using BubbleUp() for early returns:
//
//	func ProcessOrder(orderID int) (res Result[Receipt]) {
//	    defer Catch(&res) // Captures panics from BubbleUp()
//	    order := FindOrder(orderID).BubbleUp()
//	    payment := ProcessPayment(order).BubbleUp()
//	    receipt := GenerateReceipt(payment).BubbleUp()
//	    return Ok(receipt)
//	}
package result

import (
	"errors"
	"fmt"

	"github.com/seyedali-dev/goxide/rusty/option"
	"github.com/seyedali-dev/goxide/rusty/types"
)

// -------------------------------------------- Types --------------------------------------------

// Result [T] represents the outcome of an operation that can fail.
// Every Result is either Ok(value) containing a successful result, or Err(error) containing an error.
// This eliminates the ambiguity of Go's (T, error) pattern where you must remember to check error before using T.
//
// Key benefits:
//   - Cannot access the value without handling the error case
//   - No nil pointer dereferences from forgotten error checks
//   - Chainable operations that short-circuit on first error
//   - Clear API showing which operations can fail
//   - Early return support via BubbleUp() method
type Result[T any] struct {
	value option.Option[T]
	err   error
}

// tryError wraps errors raised by BubbleUp() to distinguish them from other panics.
type tryError struct {
	error
}

// -------------------------------------------- Constants --------------------------------------------

// ErrEmptyResult is returned when a Result is in error state but was initialized with nil error.
// This prevents nil error values from propagating through your application.
var ErrEmptyResult = fmt.Errorf("result is error but error was nil")

// -------------------------------------------- Public Functions --------------------------------------------

// Ok wraps a successful value into a Result[T].
// Use this when an operation succeeds and you have a valid result to return.
//
// When to use:
//   - When an operation completes successfully
//   - When validation passes and you have a valid value
//   - When converting from traditional (T, error) where error is nil
//
// Example - Successful database query:
//
//	func (r *Repository) GetUser(id int) Result[User] {
//	    user, err := r.db.QueryUser(id)
//	    if err != nil {
//	        return Err[User](err)
//	    }
//	    return Ok(user) // Success case
//	}
func Ok[T any](value T) Result[T] {
	return Result[T]{
		value: option.Some(value),
		err:   nil,
	}
}

// Err creates a Result[T] representing a failed operation.
// Use this when an operation fails and you need to propagate the error.
//
// When to use:
//   - When an operation fails (database error, network error, etc.)
//   - When validation fails
//   - When business rules are violated
//   - When converting from traditional (T, error) where error is not nil
//
// Note: If you pass nil as error, accessing Err() will return ErrEmptyResult instead.
//
// Example - Database query failure:
//
//	func (r *Repository) FindUser(email string) Result[User] {
//	    user := User{}
//	    err := r.db.QueryRow("SELECT * FROM users WHERE email = ?", email).Scan(&user)
//	    if err == sql.ErrNoRows {
//	        return Err[User](fmt.Errorf("user not found: %s", email))
//	    }
//	    if err != nil {
//	        return Err[User](fmt.Errorf("database error: %w", err))
//	    }
//	    return Ok(user)
//	}
func Err[T any](err error) Result[T] {
	return Result[T]{
		err: err,
	}
}

// IsOk reports whether the Result contains a successful value.
// Use this for explicit checking before accessing the value.
//
// When to use:
//   - When you need to branch on success/failure
//   - When logging or debugging operation outcomes
//   - When you need to perform different actions based on success
//
// Example - Conditional processing based on result:
//
//	func ProcessUser(userID int) {
//	    userResult := repo.FindByID(userID)
//	    if userResult.IsOk() {
//	        user := userResult.Unwrap()
//	        SendWelcomeEmail(user)
//	        log.Printf("Processed user: %s", user.Email)
//	    } else {
//	        log.Printf("Failed to process user %d: %v", userID, userResult.Err())
//	    }
//	}
func (r Result[T]) IsOk() bool {
	return r.value.IsSome()
}

// IsErr reports whether the Result contains an error.
// Equivalent to !IsOk() but reads more naturally when checking for failure.
//
// When to use:
//   - When early-returning on errors
//   - When you want to handle errors first
//   - When checking for failure is more natural than checking for success
//
// Example - Early return on error:
//
//	func CreateOrder(userID int, items []Item) Result[Order] {
//	    userResult := repo.FindUser(userID)
//	    if userResult.IsErr() {
//	        return result.Err[Order](fmt.Errorf("user validation failed: %w", userResult.Err()))
//	    }
//	    // Continue with order creation...
//	}
func (r Result[T]) IsErr() bool {
	return !r.IsOk()
}

// Value returns an Option[T] containing the value if Ok, or None if Err.
// Use this when you want to work with the value in an Option context.
//
// When to use:
//   - When you want to treat the success value as optional
//   - When converting Result to Option (discarding error details)
//   - When chaining with Option operations
//
// Example - Converting Result to Option when error details don't matter:
//
//	func TryGetCachedUser(userID int) option.Option[User] {
//	    cacheResult := FetchFromCache(userID)
//	    return cacheResult.Value() // Some(user) if ok, None if error
//	}
func (r Result[T]) Value() option.Option[T] {
	return r.value
}

// Err returns the error if Result is in error state, otherwise returns nil.
// If Result was initialized as Err(nil), this returns ErrEmptyResult to prevent nil errors.
//
// When to use:
//   - When you need to access the error for logging
//   - When wrapping errors for context
//   - When propagating errors up the call stack
//
// Example - Error logging with context:
//
//	func UpdateUser(user User) Result[User] {
//	    result := repo.Update(user)
//	    if result.IsErr() {
//	        log.Printf("Failed to update user %d: %v", user.ID, result.Err())
//	        return result.Err[User](fmt.Errorf("update failed: %w", result.Err()))
//	    }
//	    return result
//	}
func (r Result[T]) Err() error {
	if r.IsErr() {
		if r.err == nil {
			return ErrEmptyResult
		}
		return r.err
	}
	return nil
}

// BubbleUp returns the value if Ok, or panics with a tryError if Err.
// This enables Rust-like ? operator behavior when combined with Catch().
// The panic will be recovered by Catch() and converted back to a Result.
//
// When to use:
//   - When you want early returns without verbose if-err-return patterns
//   - When building sequential operations that should stop on first error
//   - When the calling function uses defer Catch(&res)
//
// IMPORTANT: The calling function MUST use defer Catch(&res) to recover the panic.
//
// Example - Sequential operations with early returns:
//
//	func ProcessOrder(orderID int) (res Result[Receipt]) {
//	    defer Catch(&res)
//	    order := FindOrder(orderID).BubbleUp()       // Returns early if error
//	    payment := ProcessPayment(order).BubbleUp()  // Returns early if error
//	    receipt := GenerateReceipt(payment).BubbleUp() // Returns early if error
//	    return Ok(receipt)
//	}
//
// Example - Chained database operations:
//
//	func GetUserEmail(userID int) (res Result[string]) {
//	    defer Catch(&res)
//	    user := repo.FindUser(userID).BubbleUp()
//	    profile := repo.FindProfile(user.ProfileID).BubbleUp()
//	    return Ok(profile.Email)
//	}
func (r Result[T]) BubbleUp() T {
	if r.IsErr() {
		panic(tryError{r.Err()})
	}
	return r.Unwrap()
}

// Catch recovers from panics raised by BubbleUp() and populates the Result pointer.
// This must be deferred at the beginning of functions that use BubbleUp().
//
// When to use:
//   - ALWAYS defer this when using BubbleUp() in a function
//   - Place as first defer statement to ensure it runs last
//   - Pass pointer to named Result return value
//
// Note: Panics that are not from BubbleUp() will be re-raised.
//
// Example - Basic usage:
//
//	func DoWork() (res Result[Data]) {
//	    defer Catch(&res) // Must be first defer
//	    data := FetchData().BubbleUp()
//	    return Ok(data)
//	}
//
// Example - With error handlers:
//
//	func GetUser(id int) (res Result[User]) {
//	    defer Catch(&res)
//	    defer CatchWith(&res, func(err error) User {
//	        log.Printf("Using cache fallback: %v", err)
//	        return GetCachedUser(id).BubbleUp()
//	    }, ErrDatabaseDown)
//	    return repo.FindUser(id)
//	}
func Catch[T any](res *Result[T]) {
	if r := recover(); r != nil {
		err, ok := r.(tryError)
		if !ok {
			// Re-panic if not a tryError
			panic(r)
		}
		*res = Err[T](err.error)
	}
}

// CatchWith recovers from specific errors and applies a handler function.
// This enables error-specific recovery strategies similar to match expressions in Rust.
// Must be deferred AFTER Catch() to handle errors before they bubble up.
//
// When to use:
//   - When you want to handle specific error types with custom logic
//   - When implementing fallback strategies (cache, retry, default values)
//   - When you need to transform or recover from known error conditions
//
// Note: Handler can call BubbleUp() which may succeed or propagate a new error.
// If no when errors specified, handler applies to ALL errors.
//
// Example - Cache fallback on database error:
//
//	func GetUser(id int) (res Result[User]) {
//	    defer Catch(&res)
//	    defer CatchWith(&res, func(err error) User {
//	        return GetCachedUser(id).BubbleUp()
//	    }, ErrDatabaseDown)
//	    return repo.FindUser(id)
//	}
//
// Example - Multiple error handlers:
//
//	func FetchData() (res Result[Data]) {
//	    defer Catch(&res)
//	    defer CatchWith(&res, func(err error) Data {
//	        return FetchFromRemote().BubbleUp()
//	    }, ErrCacheMiss)
//	    defer CatchWith(&res, func(err error) Data {
//	        return GetFromCache().BubbleUp()
//	    }, ErrDatabaseTimeout)
//	    return repo.QueryData()
//	}
//
// Example - Handle any error:
//
//	func GetConfig() (res Result[Config]) {
//	    defer Catch(&res)
//	    defer CatchWith(&res, func(err error) Config {
//	        log.Printf("Using defaults: %v", err)
//	        return DefaultConfig()
//	    }) // No when errors = handles all errors
//	    return LoadConfigFile()
//	}
func CatchWith[T any](res *Result[T], handler func(error) T, when ...error) {
	defer func() {
		if res.IsOk() {
			return
		}

		err := res.Err()
		// No specific errors means handle all errors
		if len(when) == 0 {
			*res = Ok(handler(err))
			return
		}

		// Check if error matches any of the specified errors
		for _, target := range when {
			if errors.Is(err, target) {
				*res = Ok(handler(err))
				return
			}
		}
	}()
	defer Catch(res)
	if r := recover(); r != nil {
		panic(r)
	}
}

// Fallback provides a default value when specific errors occur.
// Simpler alternative to CatchWith when you just need a constant fallback.
//
// When to use:
//   - When you have a simple default value for error cases
//   - When you don't need custom error handling logic
//   - When the fallback doesn't require computation
//
// Note: If no when errors specified, fallback applies to ALL errors.
//
// Example - Configuration with defaults:
//
//	func GetTimeout() (res Result[int]) {
//	    defer Catch(&res)
//	    defer Fallback(&res, 30, ErrConfigMissing, ErrInvalidConfig)
//	    return LoadTimeoutConfig()
//	}
//
// Example - User data with guest fallback:
//
//	func GetUserName(id int) (res Result[string]) {
//	    defer Catch(&res)
//	    defer Fallback(&res, "Guest", ErrUserNotFound)
//	    user := repo.FindUser(id).BubbleUp()
//	    return Ok(user.Name)
//	}
//
// Example - Fallback for any error:
//
//	func GetFeatureFlag(name string) (res Result[bool]) {
//	    defer Catch(&res)
//	    defer Fallback(&res, false) // Default to false for any error
//	    return config.GetFlag(name)
//	}
func Fallback[T any](res *Result[T], fallback T, when ...error) {
	defer CatchWith(res, func(_ error) T { return fallback }, when...)
	if r := recover(); r != nil {
		panic(r)
	}
}

// CatchErr adapts Catch for functions returning (T, error) signature.
// Useful when implementing interfaces or overriding methods with traditional Go signatures.
//
// When to use:
//   - When implementing interfaces that require (T, error) returns
//   - When overriding methods in generated code
//   - When integrating with code that expects traditional Go error handling
//
// Example - HTTP handler implementation:
//
//	func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) (user User, err error) {
//	    defer CatchErr(&user, &err)
//	    userID := ParseUserID(r).BubbleUp()
//	    return repo.FindUser(userID).BubbleUp(), nil
//	}
//
// Example - Interface implementation:
//
//	func (s *Service) FetchData(ctx context.Context) (data Data, err error) {
//	    defer CatchErr(&data, &err)
//	    config := LoadConfig().BubbleUp()
//	    return QueryAPI(ctx, config).BubbleUp(), nil
//	}
func CatchErr[T any](out *T, err *error) {
	var res Result[T]
	defer func() {
		if r := recover(); r != nil {
			// Only handle tryError panics, re-panic others
			if tryErr, ok := r.(tryError); ok {
				*err = tryErr.error     // a BubbleUp panic - convert to error return
				*out = types.Value[T]() // Set to zero value
				return
			} else {
				panic(r) // Re-panic non-tryError panics
			}
		}
		// If no panic occurred, don't modify the return values
	}()
	defer Catch(&res)

	if r := recover(); r != nil {
		panic(r)
	}
}

// Expect returns the value if Ok, or panics with the provided message if Err.
// Use ONLY in tests or when the error represents an unrecoverable programming error.
//
// When to use:
//   - In unit tests where failure indicates a test problem
//   - After validation where error is impossible
//   - When error represents an invariant violation
//
// AVOID in production code paths where errors are expected business outcomes.
//
// Example - Test assertion:
//
//	func TestCreateUser(t *testing.T) {
//	    result := service.CreateUser("test@example.com", "password")
//	    user := result.Expect("user creation should succeed in test")
//	    assert.Equal(t, "test@example.com", user.Email)
//	}
//
// Example - Critical initialization:
//
//	func InitDatabase() {
//	    result := ConnectDB(config)
//	    db := result.Expect("database connection is required for application startup")
//	}
func (r Result[T]) Expect(panicMsg string) T {
	return r.Value().Expect(panicMsg)
}

// Wrap converts a (value, error) pair into a Result[T].
// This bridges traditional Go error handling with Result patterns.
//
// When to use:
//   - When calling existing functions that return (T, error)
//   - When migrating code to use Result without changing function signatures
//   - When integrating with libraries that use traditional error handling
//
// Example - Wrapping database operations:
//
//	func FindUserByID(id int) Result[*User] {
//	    user, err := db.QueryUser(id)
//	    return result.Wrap(user, err)
//	}
//
// Example - Wrapping file operations:
//
//	func ReadConfigFile(path string) Result[[]byte] {
//	    data, err := os.ReadFile(path)
//	    return result.Wrap(data, err)
//	}
func Wrap[T any](value T, err error) Result[T] {
	if err != nil {
		return Err[T](err)
	}
	return Ok(value)
}

// WrapPtr converts a (pointer, error) pair into Result[*T], treating nil pointers as errors.
// Use when nil values represent failure conditions.
//
// When to use:
//   - When functions return (*T, error) and nil pointers indicate failure
//   - When you want to enforce non-nil return values
//   - When dealing with lookup operations that should find entities
//
// Example - Wrapping entity lookups:
//
//	func FindOrganization(name string) Result[*Organization] {
//	    org, err := db.FindOrgByName(name)
//	    return result.WrapPtr(org, err)
//	}
//
// Example - Wrapping cache lookups:
//
//	func GetCachedUser(userID int) Result[*User] {
//	    user, err := cache.GetUser(userID)
//	    return result.WrapPtr(user, err)
//	}
func WrapPtr[T any](value *T, err error) Result[*T] {
	if err != nil {
		return Err[*T](err)
	}
	if value == nil {
		return Err[*T](fmt.Errorf("nil value returned"))
	}
	return Ok(value)
}

// WrapFunc wraps a function returning (T, error) into a function returning Result[T].
// Use to adapt existing functions to Result patterns.
//
// When to use:
//   - When you want to create reusable Result-based versions of existing zero-argument functions
//   - When building higher-level APIs with Result patterns
//   - When composing multiple traditional functions together
//
// Example - Creating Result-based API clients:
//
//	var loadConfig = result.WrapFunc(config.LoadConfig)
//	// Now loadConfig() returns Result[Config] instead of (Config, error)
//
// Example - Wrapping a no-arg database connection:
//
//	var connectDB = result.WrapFunc(db.Connect)
//
//	// Usage:
//	func Initialize() Result[Database] {
//	    config := loadConfig()        // Result[Config]
//	    db := result.AndThen(config, func(c Config) Result[Database] {
//	        return connectDB()        // Result[Database]
//	    })
//	    return db
//	}
func WrapFunc[T any](fn func() (T, error)) func() Result[T] {
	return func() Result[T] {
		return Wrap(fn())
	}
}

// WrapPtrFunc wraps a zero-argument function returning (*T, error)
// into a function returning Result[*T], treating nil pointers as errors.
//
// When to use:
//   - When you have a function like func() (*User, error) and want a Result-based version
//   - When you want to reuse the adapter across multiple call sites
//   - When building railway-oriented pipelines with lookup functions
//
// Example:
//
//	var findUser = result.WrapPtrFunc(db.FindUserByID)
//	// Now findUser() returns Result[*User], not (*User, error)
//
//	func ProcessUser(userID int) Result[Profile] {
//	    user := findUser() // Result[*User]
//	    return user.AndThen(func(u *User) Result[Profile] {
//	        return findProfile(u.ID) // another Result-returning function
//	    })
//	}
func WrapPtrFunc[T any](fn func() (*T, error)) func() Result[*T] {
	return func() Result[*T] {
		ptr, err := fn()
		return WrapPtr(ptr, err)
	}
}

// WrapFunc1 wraps a single-argument function returning (T, error) into a function returning Result[T].
//
// When to use:
//   - When adapting single-argument functions to Result patterns
//   - When creating reusable adapters for common operations
//
// Example - Database operation adapters:
//
//	var findUserByID = result.WrapFunc1(db.FindUserByID)
//	var findOrgByName = result.WrapFunc1(db.FindOrgByName)
//
//	// Usage:
//	userResult := findUserByID(123)
//	orgResult := findOrgByName("acme")
func WrapFunc1[A, T any](fn func(A) (T, error)) func(A) Result[T] {
	return func(a A) Result[T] {
		return Wrap(fn(a))
	}
}

// WrapPtrFunc1 wraps a single-argument function returning (*T, error)
// into a function returning Result[*T], treating nil pointers as errors.
//
// When to use:
//   - When you have a function like func(id int) (*User, error) and want a Result-based version
//   - When you want to reuse the adapter across multiple call sites
//   - When chaining multiple pointer-returning lookups
//
// Example:
//
//	var findUser = result.WrapPtrFunc1(db.FindUserByID)
//	var findProfile = result.WrapPtrFunc1(db.FindProfileByUserID)
//
//	func GetCompleteUser(userID int) Result[CompleteUser] {
//	    return findUser(userID).
//	        AndThen(func(u *User) Result[Profile] {
//	            return findProfile(u.ID)
//	        }).
//	        Map2(func(u *User, p *Profile) CompleteUser {
//	            return CompleteUser{User: u, Profile: p}
//	        })
//	}
func WrapPtrFunc1[A, T any](fn func(A) (*T, error)) func(A) Result[*T] {
	return func(a A) Result[*T] {
		ptr, err := fn(a)
		return WrapPtr(ptr, err)
	}
}

// Unwrap returns the value if Ok, or panics with a generic message if Err.
// Shorthand for Expect with a default panic message.
//
// When to use:
//   - In tests for quick assertions
//   - When prototyping
//
// PREFER UnwrapOr, UnwrapOrElse, or proper error handling in production.
//
// Example - Quick test validation:
//
//	func TestFindUser(t *testing.T) {
//	    result := repo.FindByID(123)
//	    user := result.Unwrap() // Panic if error - acceptable in tests
//	    assert.NotEmpty(t, user.Name)
//	}
func (r Result[T]) Unwrap() T {
	return r.Expect("value should be present when unwrap is called")
}

// UnwrapOr returns the value if Ok, otherwise returns the provided default.
// This is the SAFE way to extract values with a fallback when errors occur.
//
// When to use:
//   - When you have a reasonable default value for error cases
//   - When you want to continue execution despite errors
//   - When the error doesn't need to be propagated
//
// Example - Configuration with defaults:
//
//	func GetMaxConnections() int {
//	    result := LoadConfigValue("max_connections")
//	    return result.UnwrapOr(100) // Default to 100 if config load fails
//	}
//
// Example - Cache with fallback:
//
//	func GetUserName(userID int) string {
//	    result := cache.Get(userID)
//	    user := result.UnwrapOr(User{Name: "Guest"})
//	    return user.Name
//	}
func (r Result[T]) UnwrapOr(defaultValue T) T {
	return If(r, types.Id[T], types.Return[error](defaultValue))
}

// UnwrapOrElse returns the value if Ok, otherwise computes a fallback from the error.
// Use when you need to generate a fallback based on the error that occurred.
//
// When to use:
//   - When the fallback depends on the error type
//   - When you want to log the error and provide a default
//   - When the fallback is expensive and should only be computed on error
//
// Example - Error-specific fallback:
//
//	func LoadUserProfile(userID int) User {
//	    result := db.FindUser(userID)
//	    return result.UnwrapOrElse(func(err error) User {
//	        log.Printf("Failed to load user %d: %v, returning guest profile", userID, err)
//	        return GuestUser()
//	    })
//	}
//
// Example - Fallback with error categorization:
//
//	func GetCachedData(key string) []byte {
//	    result := cache.Fetch(key)
//	    return result.UnwrapOrElse(func(err error) []byte {
//	        if errors.Is(err, ErrCacheExpired) {
//	            return RefreshCache(key)
//	        }
//	        return []byte{}
//	    })
//	}
func (r Result[T]) UnwrapOrElse(fn func(error) T) T {
	return If(r, types.Id[T], fn)
}

// ExpectErr returns the error if Result is Err, or panics if Result is Ok.
// Use this in tests when you expect an operation to fail.
//
// When to use:
//   - In tests validating error conditions
//   - When you need to assert that an operation failed
//   - When testing validation logic
//
// Example - Testing validation errors:
//
//	func TestInvalidEmail(t *testing.T) {
//	    result := ValidateEmail("invalid-email")
//	    err := result.ExpectErr("validation should fail for invalid email")
//	    assert.Contains(t, err.Error(), "invalid email")
//	}
//
// Example - Testing authorization:
//
//	func TestUnauthorizedAccess(t *testing.T) {
//	    result := service.AccessResource(unauthorizedUser, resourceID)
//	    err := result.ExpectErr("should reject unauthorized access")
//	    assert.ErrorIs(t, err, ErrUnauthorized)
//	}
func (r Result[T]) ExpectErr(panicMsg string) error {
	if r.IsErr() {
		return r.Err()
	}
	panic(panicMsg)
}

// MapError transforms the error if Result is Err, leaving Ok values unchanged.
// Use this to wrap, annotate, or transform errors while preserving successful values.
//
// When to use:
//   - When adding context to errors (wrapping with fmt.Errorf)
//   - When converting between error types
//   - When sanitizing errors for external APIs
//
// Example - Adding context to database errors:
//
//	func GetUser(id int) Result[User] {
//	    result := repo.FindByID(id)
//	    return result.MapError(func(err error) error {
//	        return fmt.Errorf("failed to get user %d: %w", id, err)
//	    })
//	}
//
// Example - Converting internal errors to API errors:
//
//	func APIGetUser(id int) Result[User] {
//	    result := db.FindUser(id)
//	    return result.MapError(func(err error) error {
//	        if errors.Is(err, sql.ErrNoRows) {
//	            return ErrUserNotFound // Convert to API error
//	        }
//	        return ErrInternalServer // Hide internal details
//	    })
//	}
func (r Result[T]) MapError(fn func(e error) error) Result[T] {
	return If(r, Ok[T], types.Compose(fn, Err[T]))
}

// Ok copies the value into out if successful, returning nil.
// Returns the error if Result is Err. This provides Go-idiomatic error handling.
//
// When to use:
//   - When integrating with traditional Go code expecting (value, error)
//   - When you want Go's "comma ok" style extraction
//   - When converting from Result back to traditional Go patterns
//
// Example - Traditional Go error handling style:
//
//	func ProcessUser(userID int) error {
//	    var user User
//	    if err := repo.FindByID(userID).Ok(&user); err != nil {
//	        return fmt.Errorf("failed to load user: %w", err)
//	    }
//	    // Use user...
//	    return nil
//	}
//
// Example - HTTP handler with Result:
//
//	func HandleGetUser(w http.ResponseWriter, r *http.Request) {
//	    var user User
//	    err := service.GetUser(userID).Ok(&user)
//	    if err != nil {
//	        http.Error(w, err.Error(), http.StatusInternalServerError)
//	        return
//	    }
//	    json.NewEncoder(w).Encode(user)
//	}
func (r Result[T]) Ok(out *T) error {
	if r.IsOk() {
		*out = r.Unwrap()
		return nil
	}
	return r.Err()
}

// If applies okFn if Result is Ok, otherwise applies errFn.
// This is a functional-style conditional that transforms Result into a non-Result type.
//
// When to use:
//   - When you need to transform Result into a different type
//   - When both success and error cases need the same return type
//   - When building functional pipelines
//
// Example - HTTP status code from Result:
//
//	func GetStatusCode(result Result[User]) int {
//	    return result.If(
//	        result,
//	        func(u User) int { return 200 },
//	        func(e error) int {
//	            if errors.Is(e, ErrNotFound) {
//	                return 404
//	            }
//	            return 500
//	        },
//	    )
//	}
//
// Example - Response message from Result:
//
//	func FormatResponse(result Result[Order]) string {
//	    return result.If(
//	        result,
//	        func(o Order) string {
//	            return fmt.Sprintf("Order %d created successfully", o.ID)
//	        },
//	        func(e error) string {
//	            return fmt.Sprintf("Order creation failed: %v", e)
//	        },
//	    )
//	}
func If[Out, T any](r Result[T], okFn func(T) Out, errFn func(error) Out) Out {
	if r.IsOk() {
		return okFn(r.Value().Unwrap())
	}
	return errFn(r.Err())
}

// Map transforms the value if Ok, leaving Err unchanged.
// Use this to transform successful results while automatically propagating errors.
//
// When to use:
//   - When you want to transform a successful value
//   - When building transformation pipelines
//   - When converting between types (e.g., DTO to entity)
//
// Example - Transform user to DTO:
//
//	func GetUserDTO(userID int) Result[UserDTO] {
//	    userResult := repo.FindByID(userID)
//	    return result.Map(userResult, func(u User) UserDTO {
//	        return UserDTO{
//	            ID:    u.ID,
//	            Email: u.Email,
//	            Name:  u.FullName(),
//	        }
//	    })
//	}
//
// Example - Extract specific field:
//
//	func GetUserEmail(userID int) Result[string] {
//	    userResult := repo.FindByID(userID)
//	    return result.Map(userResult, func(u User) string {
//	        return u.Email
//	    })
//	}
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	return If(r, types.Compose(fn, Ok[U]), Err[U])
}

// FlatMap chains a Result-returning function, flattening nested Results.
// Use when your transformation might fail and return a Result (prevents Result[Result[T]]).
//
// When to use:
//   - When the transformation itself can fail
//   - When chaining multiple fallible operations
//   - When you want errors to short-circuit the chain
//
// Example - Chained database operations:
//
//	func GetUserPrimaryEmail(userID int) Result[string] {
//	    userResult := repo.FindUser(userID)
//	    // FlatMap prevents Result[Result[string]]
//	    return result.FlatMap(userResult, func(u User) Result[string] {
//	        return repo.FindPrimaryEmail(u.ID)
//	    })
//	}
//
// Example - Validation chain:
//
//	func CreateUser(email, password string) Result[User] {
//	    emailResult := ValidateEmail(email)
//	    return result.FlatMap(emailResult, func(validEmail string) Result[User] {
//	        passwordResult := ValidatePassword(password)
//	        return result.FlatMap(passwordResult, func(validPass string) Result[User] {
//	            return repo.CreateUser(validEmail, validPass)
//	        })
//	    })
//	}
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	return If(Map(r, fn), types.Id[Result[U]], Err[U])
}

// AndThen chains a Result-returning function if Ok, short-circuiting on Err.
// Similar to FlatMap but more explicit about sequencing operations.
//
// When to use:
//   - When you have a sequence of operations that depend on each other
//   - When you want to make the chaining intent explicit
//   - When building railway-oriented programming pipelines
//
// Example - Sequential validation and processing:
//
//	func RegisterUser(req RegistrationRequest) Result[User] {
//	    return ValidateEmail(req.Email).
//	        AndThen(func(email string) Result[User] {
//	            return ValidatePassword(req.Password)
//	        }).
//	        AndThen(func(_ string) Result[User] {
//	            return CreateUserAccount(req)
//	        }).
//	        AndThen(func(user User) Result[User] {
//	            return SendVerificationEmail(user)
//	        })
//	}
//
// Example - Multi-step order processing:
//
//	func ProcessOrder(orderReq OrderRequest) Result[Receipt] {
//	    return ValidateOrder(orderReq).
//	        AndThen(func(order Order) Result[Payment] {
//	            return ChargePayment(order)
//	        }).
//	        AndThen(func(payment Payment) Result[Receipt] {
//	            return GenerateReceipt(payment)
//	        })
//	}
func AndThen[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	return If(r, fn, Err[U])
}

// Map2 combines two Results by applying fn if both are Ok, otherwise returns the first error.
// Use when you need to combine two independent operations.
//
// When to use:
//   - When you have two independent operations that both must succeed
//   - When you want to combine results from parallel operations
//   - When you need both values before proceeding
//
// Example - Combining user and permissions:
//
//	func GetUserWithPermissions(userID int) Result[UserWithPerms] {
//	    userResult := repo.FindUser(userID)
//	    permsResult := repo.FindPermissions(userID)
//	    return result.Map2(userResult, permsResult, func(u User, p Permissions) UserWithPerms {
//	        return UserWithPerms{User: u, Permissions: p}
//	    })
//	}
//
// Example - Validating multiple fields:
//
//	func ValidateRegistration(email, password string) Result[Registration] {
//	    emailResult := ValidateEmail(email)
//	    passwordResult := ValidatePassword(password)
//	    return result.Map2(emailResult, passwordResult, func(e, p string) Registration {
//	        return Registration{Email: e, Password: p}
//	    })
//	}
func Map2[T, U, V any](r Result[T], s Result[U], fn func(T, U) V) Result[V] {
	if r.IsErr() {
		return Err[V](r.Err())
	}
	if s.IsErr() {
		return Err[V](s.Err())
	}
	return Ok(fn(r.Value().Unwrap(), s.Value().Unwrap()))
}

// Map3 combines three Results by applying fn if all are Ok, otherwise returns the first error.
// Use when you need to combine three independent operations.
//
// When to use:
//   - When you have three independent operations that all must succeed
//   - When building complex validation with multiple checks
//   - When aggregating data from multiple sources
//
// Example - Complete user profile with multiple queries:
//
//	func GetCompleteProfile(userID int) Result[CompleteProfile] {
//	    userResult := repo.FindUser(userID)
//	    addressResult := repo.FindAddress(userID)
//	    prefsResult := repo.FindPreferences(userID)
//	    return result.Map3(userResult, addressResult, prefsResult,
//	        func(u User, a Address, p Preferences) CompleteProfile {
//	            return CompleteProfile{
//	                User:        u,
//	                Address:     a,
//	                Preferences: p,
//	            }
//	        })
//	}
//
// Example - Multi-field validation:
//
//	func ValidateUserInput(email, password, username string) Result[ValidatedInput] {
//	    emailResult := ValidateEmail(email)
//	    passResult := ValidatePassword(password)
//	    usernameResult := ValidateUsername(username)
//	    return result.Map3(emailResult, passResult, usernameResult,
//	        func(e, p, u string) ValidatedInput {
//	            return ValidatedInput{Email: e, Password: p, Username: u}
//	        })
//	}
func Map3[T, U, V, W any](r Result[T], s Result[U], t Result[V], fn func(T, U, V) W) Result[W] {
	if r.IsErr() {
		return Err[W](r.Err())
	}
	if s.IsErr() {
		return Err[W](s.Err())
	}
	if t.IsErr() {
		return Err[W](t.Err())
	}
	return Ok(fn(r.Value().Unwrap(), s.Value().Unwrap(), t.Value().Unwrap()))
}
