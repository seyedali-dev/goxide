// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package option. option provides a Rust-like Option[T] type for Go to handle optional values safely.
// It eliminates nil pointer dereferences and makes the absence of values explicit in function signatures.
// Use Option[T] instead of *T when you want to be explicit about "value may not exist" semantics.
//
// Common use cases:
//   - Database queries that may return no rows (instead of returning nil + error)
//   - Configuration values that are optional (instead of using pointers)
//   - Cache lookups that may miss (instead of checking map existence)
//   - Parsing operations that may fail (instead of returning zero values)
package option

import "github.com/seyedali-dev/goxide/rusty/types"

// -------------------------------------------- Types --------------------------------------------

// Option [T] represents an optional value of type T.
// Every Option is either Some(value) containing a value, or None representing absence.
// This eliminates the need for nil checks and makes your API explicit about optionality.
//
// Key benefits:
//   - Type-safe: The compiler forces you to handle both Some and None cases
//   - Explicit: Function signatures clearly show when a value might be absent
//   - Chainable: Methods like Map and FlatMap enable functional composition
type Option[T any] struct {
	isSome bool
	value  *T
}

// -------------------------------------------- Public Functions --------------------------------------------

// Some wraps a non-nil value into an Option[T] that is present.
// Use this when you have a valid value that you want to wrap in an Option.
//
// When to use:
//   - When you successfully found/computed a value
//   - When returning from a function that might sometimes have no value
//   - When you want to convert a regular value into an optional context
//
// Example - Converting database result to Option:
//
//	func GetCachedUser(userID int) Option[User] {
//	    if user, exists := cache.Get(userID); exists {
//	        return option.Some(user) // Value found in cache
//	    }
//	    return option.None[User]() // Cache miss
//	}
func Some[T any](value T) Option[T] {
	return Option[T]{
		isSome: true,
		value:  &value,
	}
}

// None returns an Option[T] representing the absence of a value.
// Use this when you have no value to return (equivalent to returning nil in traditional Go).
//
// When to use:
//   - When a search/lookup operation finds nothing
//   - When a computation cannot produce a valid result
//   - When you want to explicitly represent "no data available"
//
// Note: None() is explicitly identical to the zero value of Option[T].
//
// Example - Repository method that may not find an entity:
//
//	func (r *UserRepository) FindByID(id int) Option[User] {
//	    user, err := r.db.Query("SELECT * FROM users WHERE id = ?", id)
//	    if err != nil || user == nil {
//	        return option.None[User]() // Not found or error occurred
//	    }
//	    return option.Some(*user)
//	}
func None[T any]() Option[T] {
	return Option[T]{}
}

// IsSome reports whether the Option contains a value.
// Use this for explicit checking before accessing the value.
//
// When to use:
//   - When you need to branch on presence/absence in if statements
//   - When you want to validate before unwrapping
//   - When logging or debugging optional states
//
// Example - Conditional processing based on value presence:
//
//	userOpt := FindUserByEmail("user@example.com")
//	if userOpt.IsSome() {
//	    user := userOpt.Unwrap()
//	    log.Printf("Found user: %s", user.Name)
//	    SendWelcomeEmail(user.Email)
//	} else {
//	    log.Println("User not found, creating new account")
//	    CreateNewUser("user@example.com")
//	}
func (optn Option[T]) IsSome() bool {
	return optn.isSome
}

// IsNone reports whether the Option is empty (contains no value).
// Equivalent to !IsSome() but reads more naturally in some contexts.
//
// When to use:
//   - When checking for absence is more natural than checking for presence
//   - When early-returning on missing values
//   - When validating that operations produced no result
//
// Example - Early return pattern for missing configuration:
//
//	func ProcessConfig(configOpt Option[Config]) error {
//	    if configOpt.IsNone() {
//	        return fmt.Errorf("configuration not provided")
//	    }
//	    config := configOpt.Unwrap()
//	    // Continue processing with config...
//	    return nil
//	}
func (optn Option[T]) IsNone() bool {
	return !optn.IsSome()
}

// Expect returns the contained value or panics with your custom message.
// Use ONLY when you are absolutely certain a value exists (e.g., in tests or after IsSome check).
//
// When to use:
//   - In unit tests where absence indicates a test failure
//   - After validation where absence is a programming error
//   - When the absence represents an unrecoverable invariant violation
//
// AVOID in production code paths where absence is a valid business case.
//
// Example - Test assertion where value must exist:
//
//	func TestUserCreation(t *testing.T) {
//	    userOpt := CreateUser("test@example.com")
//	    user := userOpt.Expect("user creation should never fail in test setup")
//	    assert.Equal(t, "test@example.com", user.Email)
//	}
//
// Example - After validation in critical initialization:
//
//	func InitializeApp() {
//	    dbOpt := LoadDatabaseConfig()
//	    if dbOpt.IsNone() {
//	        log.Fatal("database configuration is required")
//	    }
//	    db := dbOpt.Expect("database config validated but still None - programming error")
//	}
func (optn Option[T]) Expect(panicMsg string) T {
	if optn.IsSome() {
		return *optn.value
	}
	panic(panicMsg)
}

// Unwrap returns the contained value or panics with a generic message.
// Shorthand for Expect with a default panic message.
//
// When to use:
//   - In tests where you want quick assertions
//   - When prototyping and you'll add proper handling later
//
// PREFER UnwrapOr or UnwrapOrElse in production code.
//
// Example - Quick test validation:
//
//	func TestFindUser(t *testing.T) {
//	    userOpt := repo.FindByID(123)
//	    user := userOpt.Unwrap() // Panics if not found - acceptable in tests
//	    assert.Equal(t, "John Doe", user.Name)
//	}
func (optn Option[T]) Unwrap() T {
	return optn.Expect("called Unwrap on None value")
}

// UnwrapOr returns the contained value if present, otherwise returns the provided default.
// This is the SAFE way to extract values with a fallback.
//
// When to use:
//   - When you have a reasonable default value
//   - When absence is normal and you want to continue with a fallback
//   - When you want to avoid if-else branching
//
// Example - Configuration with defaults:
//
//	func GetServerPort() int {
//	    portOpt := LoadPortFromConfig()
//	    return portOpt.UnwrapOr(8080) // Default to 8080 if not configured
//	}
//
// Example - User preferences with fallback:
//
//	func GetUserTheme(userID int) string {
//	    themeOpt := FetchUserTheme(userID)
//	    return themeOpt.UnwrapOr("light") // Default theme if user hasn't set one
//	}
func (optn Option[T]) UnwrapOr(defaultValue T) T {
	return If(optn, types.Id[T], types.Return0(defaultValue))
}

// UnwrapOrElse returns the value if present, otherwise computes a fallback using the provided function.
// Use when the fallback is expensive to compute or requires dynamic calculation.
//
// When to use:
//   - When the default value is expensive to create (database call, API request)
//   - When the fallback requires computation that shouldn't run unless needed
//   - When the fallback depends on runtime conditions
//
// Example - Expensive fallback (database query):
//
//	func GetCachedOrFetchUser(userID int) User {
//	    cachedOpt := GetFromCache(userID)
//	    return cachedOpt.UnwrapOrElse(func() User {
//	        // This database query only runs if cache misses
//	        user, _ := FetchUserFromDB(userID)
//	        return user
//	    })
//	}
//
// Example - Dynamic fallback (current timestamp):
//
//	func GetLastLoginTime(userID int) time.Time {
//	    loginOpt := FetchLastLogin(userID)
//	    return loginOpt.UnwrapOrElse(func() time.Time {
//	        return time.Now() // Generate current time only if no last login exists
//	    })
//	}
func (optn Option[T]) UnwrapOrElse(fn func() T) T {
	return If(optn, types.Id[T], fn)
}

// Some copies the value into the provided pointer if present, returning true if successful.
// This is Go's idiomatic way to extract an optional value (similar to map access pattern).
//
// When to use:
//   - When you want Go-style "comma ok" idiom for optional values
//   - When you need to extract into an existing variable
//   - When integrating with code that expects the "value, ok" pattern
//
// Example - Go-idiomatic optional extraction:
//
//	var user User
//	if userOpt.Some(&user) {
//	    log.Printf("Processing user: %s", user.Name)
//	    ProcessUser(user)
//	} else {
//	    log.Println("No user found, skipping processing")
//	}
//
// Example - Multiple optional extractions:
//
//	var config Config
//	var port int
//	if configOpt.Some(&config) && config.Port.Some(&port) {
//	    server.Start(port)
//	}
func (optn Option[T]) Some(out *T) bool {
	if optn.IsSome() {
		*out = optn.Unwrap()
		return true
	}
	return false
}

// If applies someFn if Option contains a value, otherwise applies noneFn.
// This is a functional-style conditional that avoids manual if-else branching.
//
// When to use:
//   - When you want to transform an Option into a non-optional result
//   - When both branches need to return the same type
//   - When building functional pipelines
//
// Example - Conditional string formatting:
//
//	func FormatUserStatus(userOpt Option[User]) string {
//	    return option.If(
//	        userOpt,
//	        func(u User) string {
//	            return fmt.Sprintf("Active user: %s (%s)", u.Name, u.Email)
//	        },
//	        func() string {
//	            return "No user logged in"
//	        },
//	    )
//	}
//
// Example - Conditional HTTP status:
//
//	func GetResponseStatus(resultOpt Option[Result]) int {
//	    return option.If(
//	        resultOpt,
//	        func(r Result) int { return 200 }, // OK if result exists
//	        func() int { return 404 },         // Not Found if absent
//	    )
//	}
func If[Out, T any](r Option[T], someFn func(T) Out, noneFn func() Out) Out {
	if r.IsSome() {
		return someFn(r.Unwrap())
	}
	return noneFn()
}

// Map transforms the contained value if present, otherwise returns None.
// This enables functional-style chaining of transformations without manual unwrapping.
//
// When to use:
//   - When you want to transform a value inside an Option
//   - When building transformation pipelines
//   - When you need to convert one optional type to another
//
// Example - Transform optional user to optional email:
//
//	func GetUserEmail(userOpt Option[User]) Option[string] {
//	    return option.Map(userOpt, func(u User) string {
//	        return u.Email
//	    })
//	}
//
// Example - Price calculation with optional discount:
//
//	func CalculateFinalPrice(priceOpt Option[float64], discountPercent float64) Option[float64] {
//	    return option.Map(priceOpt, func(price float64) float64 {
//	        return price * (1 - discountPercent/100)
//	    })
//	}
//
// Example - String transformation:
//
//	func GetUppercaseUsername(userOpt Option[User]) Option[string] {
//	    return option.Map(userOpt, func(u User) string {
//	        return strings.ToUpper(u.Username)
//	    })
//	}
func Map[T, U any](r Option[T], fn func(T) U) Option[U] {
	return If(r, types.Compose(fn, Some[U]), None[U])
}

// FlatMap chains an Option-returning function, flattening nested Options.
// Use when your transformation function itself returns an Option (prevents Option[Option[T]]).
//
// When to use:
//   - When the transformation might fail and return None
//   - When chaining multiple optional operations
//   - When avoiding nested Option types (Option[Option[T]])
//
// Example - Chained database lookups:
//
//	func GetUserPrimaryAddress(userID int) Option[Address] {
//	    userOpt := FindUserByID(userID)
//	    // FlatMap prevents Option[Option[Address]]
//	    return option.FlatMap(userOpt, func(u User) Option[Address] {
//	        return FindAddressByID(u.PrimaryAddressID)
//	    })
//	}
//
// Example - Safe nested property access:
//
//	func GetCompanyEmailDomain(userOpt Option[User]) Option[string] {
//	    return option.FlatMap(userOpt, func(u User) Option[string] {
//	        if u.Company == nil {
//	            return option.None[string]()
//	        }
//	        emailParts := strings.Split(u.Company.Email, "@")
//	        if len(emailParts) != 2 {
//	            return option.None[string]()
//	        }
//	        return option.Some(emailParts[1])
//	    })
//	}
//
// Example - Pipeline of optional operations:
//
//	func ProcessOrder(orderID int) Option[Receipt] {
//	    orderOpt := FindOrder(orderID)
//	    return option.FlatMap(orderOpt, func(order Order) Option[Receipt] {
//	        paymentOpt := ProcessPayment(order)
//	        return option.FlatMap(paymentOpt, func(payment Payment) Option[Receipt] {
//	            return GenerateReceipt(order, payment)
//	        })
//	    })
//	}
func FlatMap[T, U any](r Option[T], fn func(T) Option[U]) Option[U] {
	return If(Map(r, fn), types.Id[Option[U]], None[U])
}

// Cast attempts to type-assert value to type T, returning Some if successful.
// This is useful for safe downcasting from interface{} or any.
//
// When to use:
//   - When working with interface{} values that might be of type T
//   - When safely extracting concrete types from generic containers
//   - When parsing or deserializing data of unknown types
//
// Example - Safe type extraction from map[string]any:
//
//	func GetConfigValue(config map[string]any, key string) Option[int] {
//	    if val, exists := config[key]; exists {
//	        return option.Cast[int](val) // Returns Some(int) or None if wrong type
//	    }
//	    return option.None[int]()
//	}
//
// Example - API response parsing:
//
//	func ParseUserResponse(data any) Option[User] {
//	    return option.Cast[User](data) // Safe cast from interface{} to User
//	}
//
// Example - Event handler with multiple types:
//
//	func HandleEvent(event any) {
//	    if userEvent := option.Cast[UserCreatedEvent](event); userEvent.IsSome() {
//	        user := userEvent.Unwrap()
//	        log.Printf("User created: %s", user.Name)
//	    } else if orderEvent := option.Cast[OrderPlacedEvent](event); orderEvent.IsSome() {
//	        order := orderEvent.Unwrap()
//	        log.Printf("Order placed: %d", order.ID)
//	    }
//	}
func Cast[T any](value any) Option[T] {
	if t, ok := value.(T); ok {
		return Some(t)
	}
	return None[T]()
}
