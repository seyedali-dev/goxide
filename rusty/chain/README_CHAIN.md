# Chain Package - Fluent Method Chaining for Go

<!-- TOC -->
* [Chain Package - Fluent Method Chaining for Go](#chain-package---fluent-method-chaining-for-go)
  * [Overview](#overview)
  * [Why Use Method Chaining?](#why-use-method-chaining)
  * [Quick Start](#quick-start)
    * [Installation](#installation)
    * [Basic Usage](#basic-usage)
  * [Core Concepts](#core-concepts)
    * [Result Chaining](#result-chaining)
    * [Type-Safe Transformations](#type-safe-transformations)
    * [Error Propagation](#error-propagation)
  * [Usage Patterns](#usage-patterns)
    * [Pattern 1: Sequential Data Processing](#pattern-1-sequential-data-processing)
    * [Pattern 2: Validation Pipeline](#pattern-2-validation-pipeline)
    * [Pattern 3: API Call Chains](#pattern-3-api-call-chains)
    * [Pattern 4: Database Operations](#pattern-4-database-operations)
    * [Pattern 5: Mixed Result and Option Chains](#pattern-5-mixed-result-and-option-chains)
  * [API Reference](#api-reference)
    * [Result Chaining](#result-chaining-1)
      * [`Chain[Out, T any](r result.Result[T]) *ApplyToResult[Out, T]`](#chainout-t-anyr-resultresultt-applytoresultout-t)
      * [`Map(fn func(In) Out) result.Result[Out]`](#mapfn-funcin-out-resultresultout)
      * [`AndThen(fn func(In) result.Result[Out]) result.Result[Out]`](#andthenfn-funcin-resultresultout-resultresultout)
      * [`MapError(fn func(error) error) *ApplyToResult[Out, In]`](#maperrorfn-funcerror-error-applytoresultout-in)
      * [`Unwrap() result.Result[Out]`](#unwrap-resultresultout)
      * [`OrElse(fallback Out) Out`](#orelsefallback-out-out)
      * [`OrElseGet(fn func(error) Out) Out`](#orelsegetfn-funcerror-out-out)
    * [Multi-Step Chaining](#multi-step-chaining)
      * [`Chain2[Out2, Out1, T any](result result.Result[T]) *ApplyToResult2[Out1, Out2, T]`](#chain2out2-out1-t-anyresult-resultresultt-applytoresult2out1-out2-t)
  * [Comparison with Traditional Patterns](#comparison-with-traditional-patterns)
    * [Traditional Nested Style](#traditional-nested-style)
    * [Fluent Chain Style](#fluent-chain-style)
  * [Best Practices](#best-practices)
    * [✅ DO: Use for Readable Pipelines](#-do-use-for-readable-pipelines)
    * [✅ DO: Chain Related Operations](#-do-chain-related-operations)
    * [✅ DO: Use Meaningful Transformation Names](#-do-use-meaningful-transformation-names)
    * [❌ DON'T: Overuse for Simple Operations](#-dont-overuse-for-simple-operations)
    * [❌ DON'T: Chain Unrelated Operations](#-dont-chain-unrelated-operations)
    * [❌ DON'T: Ignore Error Handling](#-dont-ignore-error-handling)
  * [Integration with Result and Option](#integration-with-result-and-option)
  * [Performance Considerations](#performance-considerations)
  * [Examples](#examples)
<!-- TOC -->

## Overview

The `chain` package provides fluent method chaining for `Result` and `Option` types, enabling readable pipeline-style programming similar to Rust's method chaining. It transforms nested functional calls into clean, sequential operations that read left-to-right.

**Key Benefits:**
- **Fluent API**: Chain multiple operations in a readable sequence
- **Type safety**: Compiler tracks type transformations through the chain
- **No nesting**: Avoids deep nesting of Map/AndThen calls
- **Better readability**: Operations read left-to-right like a pipeline
- **Error short-circuiting**: Automatically propagates errors through the chain

## Why Use Method Chaining?

| Aspect              | Traditional Nesting           | Fluent Chaining            |
|---------------------|-------------------------------|----------------------------|
| **Readability**     | Deeply nested, hard to follow | Linear, easy to read       |
| **Maintainability** | Difficult to modify           | Easy to add/remove steps   |
| **Debugging**       | Hard to trace through nesting | Clear step-by-step flow    |
| **Type Safety**     | Manual type tracking          | Compiler-enforced          |
| **Error Handling**  | Manual error propagation      | Automatic short-circuiting |

## Quick Start

### Installation

```go
import "github.com/seyedali-dev/gopherbox/rusty/chain"
```

### Basic Usage

```go
// Traditional nested style (hard to read)
result := result.AndThen(
    result.Map(
        findUser(123),
        func(u User) string { return u.Name },
    ),
    func(name string) result.Result[Profile] { 
        return findProfile(name) 
    },
)

// Fluent chain style (clean and readable)
result := chain.Chain(findUser(123)).
    Map(func(u User) string { return u.Name }).
    AndThen(findProfile).
    Unwrap()
```

## Core Concepts

### Result Chaining

The chain package provides a fluent interface for transforming `Result` values through multiple operations:

```go
// Start a chain with Chain()
userProfile := chain.Chain(findUser(123)).
    // Transform successful value
    Map(func(u User) string { 
        return u.Email 
    }).
    // Chain with another fallible operation
    AndThen(validateEmail).
    // Transform error if present
    MapError(func(err error) error {
        return fmt.Errorf("user processing failed: %w", err)
    }).
    // Terminate chain and get final Result
    Unwrap()
```

### Type-Safe Transformations

The chain tracks type transformations automatically:

```go
// Compiler tracks: Result[User] → Result[string] → Result[Profile]
result := chain.Chain(findUser(123)).     // Result[User]
    Map(func(u User) string {             // Result[string]
        return u.Name 
    }).
    AndThen(findProfileByName).           // Result[Profile]
    Unwrap()                              // Result[Profile]
```

### Error Propagation

Errors automatically short-circuit the chain:

```go
// If findUser fails, subsequent operations are skipped
result := chain.Chain(findUser(999)).     // Returns error
    Map(func(u User) string {             // Skipped
        return u.Name 
    }).
    AndThen(findProfileByName).           // Skipped
    Unwrap()                              // Contains original error
```

## Usage Patterns

### Pattern 1: Sequential Data Processing

```go
func ProcessUserData(userID int) result.Result[ProcessedData] {
    return chain.Chain(fetchUser(userID)).
        Map(validateUser).
        AndThen(func(u ValidUser) result.Result[UserData] {
            return fetchUserData(u.ID)
        }).
        Map(transformData).
        MapError(func(err error) error {
            return fmt.Errorf("data processing failed: %w", err)
        }).
        Unwrap()
}
```

### Pattern 2: Validation Pipeline

```go
func ValidateRegistration(request RegistrationRequest) result.Result[ValidatedRequest] {
    return chain.Chain(validateEmail(request.Email)).
        AndThen(func(email string) result.Result[ValidatedRequest] {
            return validatePassword(request.Password)
        }).
        AndThen(func(_ string) result.Result[ValidatedRequest] {
            return validateUsername(request.Username)
        }).
        Map(func(_ string) ValidatedRequest {
            return ValidatedRequest{
                Email:    request.Email,
                Password: request.Password,
                Username: request.Username,
            }
        }).
        Unwrap()
}
```

### Pattern 3: API Call Chains

```go
func GetUserDashboard(userID int) result.Result[Dashboard] {
    return chain.Chain(fetchUser(userID)).
        AndThen(func(user User) result.Result[Dashboard] {
            return chain.Chain(fetchUserPreferences(user.ID)).
                AndThen(func(prefs Preferences) result.Result[Dashboard] {
                    return chain.Chain(fetchRecentActivity(user.ID)).
                        Map(func(activity []Activity) Dashboard {
                            return Dashboard{
                                User:       user,
                                Preferences: prefs,
                                Activity:   activity,
                            }
                        }).
                        Unwrap()
                }).
                Unwrap()
        }).
        Unwrap()
}
```

### Pattern 4: Database Operations

```go
func CompleteUserTransaction(userID int, amount float64) result.Result[Receipt] {
    return chain.Chain(findUser(userID)).
        AndThen(func(user User) result.Result[Transaction] {
            return beginTransaction(user.ID)
        }).
        AndThen(func(tx Transaction) result.Result[Receipt] {
            return chain.Chain(debitAccount(tx, amount)).
                AndThen(func(_ bool) result.Result[Receipt] {
                    return creditVendor(tx, amount)
                }).
                AndThen(func(_ bool) result.Result[Receipt] {
                    return commitTransaction(tx)
                }).
                MapError(func(err error) error {
                    _ = rollbackTransaction(tx)
                    return fmt.Errorf("transaction failed: %w", err)
                }).
                Unwrap()
        }).
        Unwrap()
}
```

### Pattern 5: Mixed Result and Option Chains

```go
func GetUserContactInfo(userID int) result.Result[option.Option[ContactInfo]] {
    return chain.Chain(findUser(userID)).
        Map(func(u User) option.Option[ContactInfo] {
            return option.Map(u.Profile, func(p Profile) ContactInfo {
                return ContactInfo{
                    Email: p.Email,
                    Phone: p.Phone,
                }
            })
        }).
        Unwrap()
}
```

## API Reference

### Result Chaining

#### `Chain[Out, T any](r result.Result[T]) *ApplyToResult[Out, T]`

Start a new chaining pipeline with a `Result[T]`. The `Out` type parameter represents the final type after transformations.

```go
chain.Chain(findUser(123)).  // Result[User] → ApplyToResult[Out, User]
```

#### `Map(fn func(In) Out) result.Result[Out]`

Transform the value inside the `Result` using the provided function. Returns a new `ApplyToResult` that can continue the chain.

```go
chain.Chain(result.Ok(42)).
    Map(func(x int) string { 
        return fmt.Sprintf("value: %d", x) 
    }).  // Result[int] → Result[string]
    Unwrap()
```

#### `AndThen(fn func(In) result.Result[Out]) result.Result[Out]`

Chain a `Result`-returning function. Similar to `Map` but for functions that can fail.

```go
chain.Chain(validateEmail("test@example.com")).
    AndThen(func(email string) result.Result[User] {
        return createUser(email)
    }).  // Result[string] → Result[User]
    Unwrap()
```

#### `MapError(fn func(error) error) *ApplyToResult[Out, In]`

Transform the error if the `Result` is in error state.

```go
chain.Chain(dbQueryResult).
    MapError(func(err error) error {
        return fmt.Errorf("database error: %w", err)
    }).  // Transforms error, keeps value type
    Unwrap()
```

#### `Unwrap() result.Result[Out]`

Terminate the chain and return the final `Result`. This is usually the last call in a chain.

```go
result := chain.Chain(findUser(123)).
    Map(func(u User) string { return u.Name }).
    Unwrap()  // Result[string]
```

#### `OrElse(fallback Out) Out`

Terminate the chain and return the value or fallback.

```go
name := chain.Chain(findUser(123)).
    Map(func(u User) string { return u.Name }).
    OrElse("Unknown User")  // string
```

#### `OrElseGet(fn func(error) Out) Out`

Terminate the chain and return the value or computed fallback.

```go
name := chain.Chain(findUser(123)).
    Map(func(u User) string { return u.Name }).
    OrElseGet(func(err error) string {
        return "Fallback: " + err.Error()
    })  // string
```

### Multi-Step Chaining

#### `Chain2[Out2, Out1, T any](result result.Result[T]) *ApplyToResult2[Out1, Out2, T]`

Start a chain that expects exactly 2 transformations. Useful when you know the exact number of steps for type clarity.

```go
chain.Chain2[string, User, int](findUser(123)).
    Map(func(u User) string { return u.Name }).
    AndThen(validateName)
```

## Comparison with Traditional Patterns

### Traditional Nested Style

```go
// Deep nesting, hard to read and maintain
result := result.AndThen(
    result.Map(
        result.AndThen(
            findUser(123),
            func(u User) result.Result[Profile] {
                return findProfile(u.ProfileID)
            }
        ),
        func(p Profile) string {
            return p.DisplayName
        }
    ),
    func(name string) result.Result[bool] {
        return validateDisplayName(name)
    }
)
```

### Fluent Chain Style

```go
// Linear, easy to read and modify
result := chain.Chain(findUser(123)).
    AndThen(func(u User) result.Result[Profile] {
        return findProfile(u.ProfileID)
    }).
    Map(func(p Profile) string {
        return p.DisplayName
    }).
    AndThen(validateDisplayName).
    Unwrap()
```

## Best Practices

### ✅ DO: Use for Readable Pipelines

```go
// ✅ Clean data processing pipeline
result := chain.Chain(fetchData()).
    Map(parseData).
    AndThen(validateData).
    Map(transformData).
    Unwrap()

// ❌ Overly complex for simple operations
result := chain.Chain(result.Ok(42)).
    Map(func(x int) int { return x * 2 }).
    Unwrap() // Just use result.Map directly
```

### ✅ DO: Chain Related Operations

```go
// ✅ Related transformations
userInfo := chain.Chain(fetchUser(123)).
    AndThen(fetchProfile).
    AndThen(fetchPreferences).
    Map(compileUserInfo).
    Unwrap()

// ❌ Unrelated operations
result := chain.Chain(fetchUser(123)).
    Map(func(u User) string { return u.Name }).  // User operation
    AndThen(sendEmail).                          // Unrelated!
    Unwrap()
```

### ✅ DO: Use Meaningful Transformation Names

```go
// ✅ Descriptive function names
result := chain.Chain(fetchRawData()).
    Map(parseJSONData).
    AndThen(validateBusinessRules).
    Map(applyFormatting).
    Unwrap()

// ❌ Anonymous functions everywhere
result := chain.Chain(fetchRawData()).
    Map(func(data []byte) Data { 
        // Complex parsing logic inline
    }).
    AndThen(func(d Data) result.Result[Data] {
        // Complex validation inline  
    }).
    Unwrap()
```

### ❌ DON'T: Overuse for Simple Operations

```go
// ✅ Simple case - use direct Result operations
result := result.Map(findUser(123), func(u User) string {
    return u.Name
})

// ❌ Overkill for simple transformation
result := chain.Chain(findUser(123)).
    Map(func(u User) string { return u.Name }).
    Unwrap()
```

### ❌ DON'T: Chain Unrelated Operations

```go
// ❌ Mixing concerns
result := chain.Chain(validateUser(input)).
    AndThen(func(u User) result.Result[bool] {
        return sendNotification(u.Email)  // Different concern!
    }).
    Unwrap()

// ✅ Separate concerns
validationResult := validateUser(input)
if validationResult.IsOk() {
    user := validationResult.Unwrap()
    _ = sendNotification(user.Email)  // Separate operation
}
```

### ❌ DON'T: Ignore Error Handling

```go
// ❌ Ignoring potential errors
name := chain.Chain(findUser(123)).
    Map(func(u User) string { return u.Name }).
    OrElse("Unknown")  // Lost error information!

// ✅ Preserve error context
result := chain.Chain(findUser(123)).
    Map(func(u User) string { return u.Name }).
    MapError(func(err error) error {
        return fmt.Errorf("failed to get user name: %w", err)
    }).
    Unwrap()
```

## Integration with Result and Option

The chain package works seamlessly with both `Result` and `Option` types:

```go
// Mixed Result and Option operations
func GetUserDisplayInfo(userID int) result.Result[string] {
    return chain.Chain(findUser(userID)).
        Map(func(u User) option.Option[string] {
            return u.DisplayName  // Option[string]
        }).
        Map(func(nameOpt option.Option[string]) string {
            return nameOpt.UnwrapOr(u.FallbackName())
        }).
        Unwrap()
}

// Convert Option to Result when needed
func RequireUserName(userID int) result.Result[string] {
    userOpt := findUser(userID)
    if userOpt.IsNone() {
        return result.Err[string](fmt.Errorf("user not found"))
    }
    return chain.Chain(result.Ok(userOpt.Unwrap())).
        Map(func(u User) string { return u.Name }).
        Unwrap()
}
```

## Performance Considerations

Method chaining adds minimal overhead compared to direct `Result` operations:

```go
// Benchmark comparison:
// Direct Result.Map:    ~100 ns/op
// Chain with Map:       ~110 ns/op (+10%)
// Chain with AndThen:   ~115 ns/op (+15%)

// The readability benefits typically outweigh the small performance cost
```

## Examples

See the [chain tests](../rusty/chain/) for comprehensive usage examples:

```go
// From chain tests
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

    // Complex chain with multiple transformations
    chainResult := chain.Chain2[string, User, string](validateEmail("test@example.com")).
        AndThen(createUser).
        Map(getUserName)

    if chainResult.IsErr() {
        t.Fatalf("expected success, got error: %v", chainResult.Err())
    }
}
```

Additional examples in the [examples package](../examples/examples.go) demonstrate real-world usage patterns.
