# Result Package - Rust-like Error Handling for Go

<!-- TOC -->
* [Result Package - Rust-like Error Handling for Go](#result-package---rust-like-error-handling-for-go)
  * [Overview](#overview)
  * [Key Features](#key-features)
    * [1. Try() - Early Return Pattern (Rust's `?` operator)](#1-try---early-return-pattern-rusts--operator)
    * [2. CatchWith - Error-Specific Recovery](#2-catchwith---error-specific-recovery)
    * [3. Fallback - Simple Default Values](#3-fallback---simple-default-values)
    * [4. CatchErr - Adapt to Traditional Signatures](#4-catcherr---adapt-to-traditional-signatures)
  * [Quick Start](#quick-start)
    * [Installation](#installation)
    * [Basic Usage](#basic-usage)
  * [Usage Patterns](#usage-patterns)
    * [Pattern 1: Sequential Operations (Use Try)](#pattern-1-sequential-operations-use-try)
    * [Pattern 2: Multi-Layer Fallback](#pattern-2-multi-layer-fallback)
    * [Pattern 3: Transaction Handling](#pattern-3-transaction-handling)
    * [Pattern 4: Validation Chain](#pattern-4-validation-chain)
    * [Pattern 5: Gradual Migration](#pattern-5-gradual-migration)
  * [Comparison: Try vs AndThen](#comparison-try-vs-andthen)
    * [Use Try() when:](#use-try-when)
    * [Use AndThen() when:](#use-andthen-when)
  * [Best Practices](#best-practices)
    * [✅ DO: Always defer Catch() first](#-do-always-defer-catch-first)
    * [✅ DO: Use named return values with Try()](#-do-use-named-return-values-with-try)
    * [✅ DO: Order deferred handlers carefully](#-do-order-deferred-handlers-carefully)
    * [✅ DO: Use CatchWith for error transformation](#-do-use-catchwith-for-error-transformation)
    * [❌ DON'T: Forget to defer Catch()](#-dont-forget-to-defer-catch)
    * [❌ DON'T: Use Try() outside functions with Catch()](#-dont-use-try-outside-functions-with-catch)
    * [❌ DON'T: Mix Try() with traditional panic/recover](#-dont-mix-try-with-traditional-panicrecover)
  * [Performance Considerations](#performance-considerations)
  * [Migration Guide](#migration-guide)
    * [Step 1: Start with new functions](#step-1-start-with-new-functions)
    * [Step 2: Wrap legacy functions](#step-2-wrap-legacy-functions)
    * [Step 3: Gradually refactor](#step-3-gradually-refactor)
  * [API Reference](#api-reference)
    * [Core Functions](#core-functions)
    * [Early Return Pattern](#early-return-pattern)
    * [Inspection](#inspection)
    * [Unwrapping](#unwrapping)
    * [Transformation](#transformation)
    * [Combination](#combination)
  * [Examples](#examples)
  * [Benchmarks](#benchmarks)
  * [README_OPTION.md](#readme_optionmd)
* [Option Package - Safe Optional Values for Go](#option-package---safe-optional-values-for-go)
  * [Overview](#overview-1)
  * [Why Use Option Instead of Pointers?](#why-use-option-instead-of-pointers)
  * [Quick Start](#quick-start-1)
    * [Installation](#installation-1)
    * [Basic Usage](#basic-usage-1)
  * [Core Concepts](#core-concepts)
    * [Creating Options](#creating-options)
    * [Checking Presence](#checking-presence)
    * [Safe Value Extraction](#safe-value-extraction)
  * [Usage Patterns](#usage-patterns-1)
    * [Pattern 1: Database Queries](#pattern-1-database-queries)
    * [Pattern 2: Configuration with Defaults](#pattern-2-configuration-with-defaults)
    * [Pattern 3: Cache Lookups](#pattern-3-cache-lookups)
    * [Pattern 4: Optional Function Parameters](#pattern-4-optional-function-parameters)
    * [Pattern 5: Chaining Transformations](#pattern-5-chaining-transformations)
  * [API Reference](#api-reference-1)
    * [Creation](#creation)
    * [Inspection](#inspection-1)
    * [Unwrapping](#unwrapping-1)
    * [Transformation](#transformation-1)
    * [Functional Helpers](#functional-helpers)
  * [Best Practices](#best-practices-1)
    * [✅ DO: Use UnwrapOr for Safe Defaults](#-do-use-unwrapor-for-safe-defaults)
    * [✅ DO: Use Map for Transformations](#-do-use-map-for-transformations)
    * [✅ DO: Use FlatMap for Optional Operations](#-do-use-flatmap-for-optional-operations)
    * [✅ DO: Use Some() for Go Idiomatic Access](#-do-use-some-for-go-idiomatic-access)
    * [❌ DON'T: Use Unwrap() in Production Code](#-dont-use-unwrap-in-production-code)
    * [❌ DON'T: Use Option for Error Handling](#-dont-use-option-for-error-handling)
  * [Migration from Pointers](#migration-from-pointers)
    * [Before: Using Pointers](#before-using-pointers)
    * [After: Using Option](#after-using-option)
  * [Integration with Result](#integration-with-result)
  * [Examples](#examples-1)
<!-- TOC -->

## Overview

The `result` package provides a Rust-like `Result[T]` type for Go with enhanced error handling capabilities including:

- **Type-safe error handling** - Compiler enforces error checking
- **Early returns** - `Try()` method enables Rust-like `?` operator behavior
- **Deferred error handling** - `CatchWith` and `Fallback` for elegant error recovery
- **Functional composition** - `Map`, `FlatMap`, `AndThen` for chaining operations
- **Zero boilerplate** - Eliminate repetitive `if err != nil` checks
- **Multiple patterns** - Choose from traditional, functional, or early-return styles

## Key Features

### 1. Try() - Early Return Pattern (Rust's `?` operator)

The `Try()` method enables Rust-like `?` operator behavior by panicking on errors. When combined with `Catch()`, this provides clean early returns without verbose if-err-return patterns.

```go
func ProcessOrder(orderID int) (res result.Result[Receipt]) {
    defer result.Catch(&res) // MUST defer this first

    // Each Try() returns early on error
    order := FindOrder(orderID).Try()
    payment := ProcessPayment(order).Try()
    receipt := GenerateReceipt(payment).Try()

    return result.Ok(receipt)
}
```

**Comparison with Traditional Go:**

```go
// Traditional Go - verbose and repetitive
func ProcessOrder(orderID int) (Receipt, error) {
    order, err := FindOrder(orderID)
    if err != nil {
        return Receipt{}, err
    }

    payment, err := ProcessPayment(order)
    if err != nil {
        return Receipt{}, err
    }

    receipt, err := GenerateReceipt(payment)
    if err != nil {
        return Receipt{}, err
    }

    return receipt, nil
}

// With Try() - clean and succinct
func ProcessOrder(orderID int) (res result.Result[Receipt]) {
    defer result.Catch(&res)

    order := FindOrder(orderID).Try()
    payment := ProcessPayment(order).Try()
    receipt := GenerateReceipt(payment).Try()

    return result.Ok(receipt)
}
```

### 2. CatchWith - Error-Specific Recovery

The `CatchWith` function allows handling specific errors with custom recovery logic. It must be deferred **after** `Catch()`.

```go
func GetUser(id int) (res result.Result[User]) {
    defer result.Catch(&res)

    // Handle database errors with cache fallback
    defer result.CatchWith(&res, func(err error) User {
        log.Printf("Database down, using cache: %v", err)
        return GetCachedUser(id).Try()
    }, ErrDatabaseDown)

    return repo.FindUser(id)
}
```

**Multiple Error Handlers:**

```go
func FetchData(id int) (res result.Result[string]) {
    defer result.Catch(&res)

    // Handlers are checked in reverse order (LIFO)
    defer result.CatchWith(&res, func (err error) string {
        return FetchFromRemote(id).Try()
    }, ErrCacheMiss)

    defer result.CatchWith(&res, func (err error) string {
        return GetFromCache(id).Try()
    }, ErrDatabaseDown)

    return repo.QueryData(id)
}
```

**Handle Any Error:**

```go
func GetConfig() (res result.Result[Config]) {
    defer result.Catch(&res)

    // No error list = handles all errors
    defer result.CatchWith(&res, func(err error) Config {
        log.Printf("Using defaults: %v", err)
        return DefaultConfig()
    })

    return LoadConfigFile()
}
```

### 3. Fallback - Simple Default Values

`Fallback` provides a simpler alternative to `CatchWith` when you just need a constant default value.

```go
func GetTimeout() (res result.Result[int]) {
    defer result.Catch(&res)
    defer result.Fallback(&res, 30, ErrConfigMissing, ErrInvalidConfig)

    return LoadTimeoutConfig()
}
```

**Fallback for Any Error:**

```go
func GetFeatureFlag(name string) (res result.Result[bool]) {
    defer result.Catch(&res)
    defer result.Fallback(&res, false) // Default to false for any error

    return config.GetFlag(name)
}
```

### 4. CatchErr - Adapt to Traditional Signatures

`CatchErr` adapts the Result pattern to traditional `(value, error)` signatures, useful for interface implementations.

```go
func HandleRequest(w http.ResponseWriter, r *http.Request) (user User, err error) {
    defer result.CatchErr(&user, &err)

    userID := ExtractUserID(r).Try()
    user = repo.FindUser(userID).Try()

    return user, nil
}
```

## Quick Start

### Installation

```go
import "github.com/seyedali-dev/gopherbox/rusty/result"
```

### Basic Usage

```go
// Traditional function wrapping
func FindUser(id int) result.Result[*User] {
    user, err := db.FindUserByID(id)
    return result.Wrap(user, err)
}

// Using Try() for early returns
func GetUserProfile(userID int) (res result.Result[Profile]) {
    defer result.Catch(&res)
    
    user := FindUser(userID).Try()
    profile := FindProfile(user.ProfileID).Try()
    
    return result.Ok(profile)
}

// Functional composition
func ProcessUser(email string) result.Result[User] {
    return ValidateEmail(email).
        AndThen(CreateUser).
        Map(SendWelcomeEmail)
}
```

## Usage Patterns

### Pattern 1: Sequential Operations (Use Try)

When you have a sequence of operations that depend on each other:

```go
func CreateAccount(email, password string) (res result.Result[Account]) {
    defer result.Catch(&res)

    validEmail := ValidateEmail(email).Try()
    hashedPass := HashPassword(password).Try()
    account := repo.CreateAccount(validEmail, hashedPass).Try()
    SendWelcomeEmail(account).Try()

    return result.Ok(account)
}
```

### Pattern 2: Multi-Layer Fallback

Cascade through multiple data sources with fallbacks:

```go
func GetData(id int) (res result.Result[string]) {
    defer result.Catch(&res)

    // Deferred handlers execute in reverse order (LIFO)
    defer result.Fallback(&res, "default-value")
    defer result.CatchWith(&res, func (err error) string {
        return FetchRemote(id).Try()
    }, ErrDatabaseDown)
    defer result.CatchWith(&res, func (err error) string {
        return GetFromDB(id).Try()
    }, ErrCacheMiss)

    return GetFromCache(id) // Try cache first
}
```

### Pattern 3: Transaction Handling

Handle database transactions with automatic rollback:

```go
func Transfer(from, to int, amount float64) (res result.Result[string]) {
    defer result.Catch(&res)

    tx := result.Wrap(db.Begin()).Try()

    defer func () {
        if res.IsErr() {
            tx.Rollback()
        }
    }()

    DebitAccount(tx, from, amount).Try()
    CreditAccount(tx, to, amount).Try()
    result.Wrap(tx.Commit()).Try()

    return result.Ok("transfer completed")
}
```

### Pattern 4: Validation Chain

Chain multiple validations together:

```go
func ValidateRegistration(req RegistrationRequest) (res result.Result[ValidatedUser]) {
    defer result.Catch(&res)

    email := ValidateEmail(req.Email).Try()
    password := ValidatePassword(req.Password).Try()
    username := ValidateUsername(req.Username).Try()
    age := ValidateAge(req.Age).Try()

    return result.Ok(ValidatedUser{
        Email:    email,
        Password: password,
        Username: username,
        Age:      age,
    })
}
```

### Pattern 5: Gradual Migration

Mix traditional and Result patterns during migration:

```go
func MigrateFunction(db *sql.DB, id int) (res result.Result[Data]) {
    defer result.Catch(&res)

    // Legacy function - wrap with Wrap()
    user, err := legacyFetchUser(db, id)
    if err != nil {
        return result.Err[Data](err)
    }

    // New Result-based function - use Try()
    enrichedData := enrichData(user).Try()

    return result.Ok(enrichedData)
}
```

## Comparison: Try vs AndThen

### Use Try() when:

- Operations are sequential and depend on each other
- You want imperative, readable code
- You need access to multiple previous results

```go
func ProcessOrder(orderID int) (res result.Result[Receipt]) {
    defer result.Catch(&res)

    order := FindOrder(orderID).Try()
    user := FindUser(order.UserID).Try() // Uses order
    payment := Charge(user, order).Try() // Uses both user and order

    return result.Ok(GenerateReceipt(payment))
}
```

### Use AndThen() when:

- Building functional pipelines
- Each step only needs the previous result
- You want explicit composition

```go
func ProcessOrder(orderID int) result.Result[Receipt] {
    return FindOrder(orderID).
        AndThen(func (order Order) result.Result[Payment] {
            return ChargeOrder(order)
        }).
        AndThen(func (payment Payment) result.Result[Receipt] {
            return GenerateReceipt(payment)
        })
}
```

## Best Practices

### ✅ DO: Always defer Catch() first

```go
func DoWork() (res result.Result[Data]) {
    defer result.Catch(&res) // MUST be first
    defer result.Fallback(&res, defaultData)

    data := FetchData().Try()
    return result.Ok(data)
}
```

### ✅ DO: Use named return values with Try()

```go
// ✅ Correct - named return
func GetData() (res result.Result[Data]) {
    defer result.Catch(&res)
    return FetchData()
}

// ❌ Incorrect - unnamed return
func GetData() result.Result[Data] {
    defer result.Catch(&res) // res is not defined!
    return FetchData()
}
```

### ✅ DO: Order deferred handlers carefully

Remember that deferred functions execute in LIFO (Last In, First Out) order:

```go
func GetData() (res result.Result[string]) {
    defer result.Catch(&res) // Executes LAST (catches all)
    defer result.Fallback(&res, "default") // Executes 3rd
    defer result.CatchWith(&res, handler2, ErrB) // Executes 2nd
    defer result.CatchWith(&res, handler1, ErrA) // Executes FIRST

    return FetchData()
}
```

### ✅ DO: Use CatchWith for error transformation

```go
func APIGetUser(id int) (res result.Result[User]) {
    defer result.Catch(&res)

    // Transform internal errors to API errors
    defer result.CatchWith(&res, func(err error) User {
        if errors.Is(err, sql.ErrNoRows) {
            result.Err[User](ErrUserNotFound).Try()
        }
        result.Err[User](ErrInternalServer).Try()
        return User{}
    })

    return db.FindUser(id)
}
```

### ❌ DON'T: Forget to defer Catch()

```go
// ❌ This will panic and crash your program!
func GetData() result.Result[Data] {
    data := FetchData().Try() // Panics, no Catch() to recover
    return result.Ok(data)
}
```

### ❌ DON'T: Use Try() outside functions with Catch()

```go
// ❌ Don't use Try() at package level or in init()
var globalData = FetchData().Try() // Will panic!

// ✅ Use traditional unwrapping instead
var globalData = FetchData().UnwrapOr(defaultData)
```

### ❌ DON'T: Mix Try() with traditional panic/recover

```go
// ❌ Don't mix patterns
func GetData() (res result.Result[Data]) {
    defer func () {
        if r := recover(); r != nil {
            // This might interfere with Catch()
        }
    }()
    defer result.Catch(&res)

    return FetchData()
}
```

## Performance Considerations

The Try/Catch pattern uses panic/recover internally, which has some overhead:

```go
// Benchmark results (approximate):
// Traditional:  100 ns/op
// Try/Catch:    150 ns/op  (~50% slower)
// AndThen:      110 ns/op  (~10% slower)
```

**Recommendations:**

- Use Try/Catch for business logic where clarity matters
- Use traditional patterns for tight loops or performance-critical code
- Use functional composition (AndThen/Map) for simple pipelines
- The readability benefit usually outweighs the small performance cost

## Migration Guide

### Step 1: Start with new functions

```go
// New function - use Result from the start
func CreateUser(email string) (res result.Result[User]) {
    defer result.Catch(&res)

    validEmail := ValidateEmail(email).Try()
    user := repo.Create(validEmail).Try()

    return result.Ok(user)
}
```

### Step 2: Wrap legacy functions

```go
// Wrap existing functions without changing them
var findUser = result.WrapFunc1(db.FindUserByID)
var loadConfig = result.WrapFunc(config.Load)

// Now use them with Try()
func DoWork() (res result.Result[Data]) {
    defer result.Catch(&res)

    user := findUser(123).Try()
    config := loadConfig().Try()

    return ProcessData(user, config)
}
```

### Step 3: Gradually refactor

```go
// Before: Traditional Go
func ProcessUser(id int) (Data, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return Data{}, err
    }

    profile, err := db.FindProfile(user.ProfileID)
    if err != nil {
        return Data{}, err
    }

    return Process(user, profile), nil
}

// After: With Result
func ProcessUser(id int) (res result.Result[Data]) {
    defer result.Catch(&res)

    user := db.FindUser(id).Try()
    profile := db.FindProfile(user.ProfileID).Try()

    return result.Ok(Process(user, profile))
}
```

## API Reference

### Core Functions

- `Ok[T](value T) Result[T]` - Create successful Result
- `Err[T](err error) Result[T]` - Create error Result
- `Wrap[T](value T, err error) Result[T]` - Convert (T, error) to Result
- `WrapPtr[T](value *T, err error) Result[*T]` - Convert (*T, error) with nil check

### Early Return Pattern

- `Try() T` - Return value or panic with error (requires Catch)
- `Catch[T](res *Result[T])` - Recover panics from Try()
- `CatchWith[T](res *Result[T], handler func(error) T, when ...error)` - Handle specific errors
- `Fallback[T](res *Result[T], fallback T, when ...error)` - Provide default for specific errors
- `CatchErr[T](out *T, err *error)` - Adapt to (T, error) signatures

### Inspection

- `IsOk() bool` - Check if Result is Ok
- `IsErr() bool` - Check if Result is Err
- `Err() error` - Get error or nil
- `Value() Option[T]` - Get value as Option

### Unwrapping

- `Unwrap() T` - Get value or panic
- `Expect(msg string) T` - Get value or panic with message
- `UnwrapOr(default T) T` - Get value or default
- `UnwrapOrElse(fn func(error) T) T` - Get value or compute default
- `ExpectErr(msg string) error` - Get error or panic if Ok

### Transformation

- `Map[T, U](r Result[T], fn func(T) U) Result[U]` - Transform value
- `FlatMap[T, U](r Result[T], fn func(T) Result[U]) Result[U]` - Chain fallible operations
- `AndThen[T, U](r Result[T], fn func(T) Result[U]) Result[U]` - Alias for FlatMap
- `MapError(fn func(error) error) Result[T]` - Transform error

### Combination

- `Map2[T, U, V](r Result[T], s Result[U], fn func(T, U) V) Result[V]` - Combine two Results
- `Map3[T, U, V, W](r Result[T], s Result[U], t Result[V], fn func(T, U, V) W) Result[W]` - Combine three Results

## Examples

See the [examples](../examples/examples.go) package for comprehensive real-world usage patterns including:

- Database queries with cache fallback
- Multi-step order processing pipelines
- Configuration loading with defaults
- Multi-layer data fetching with cascading fallbacks
- HTTP handlers with proper error responses
- Validation chains
- Transaction handling with automatic rollback
- Context-aware operations with timeout
- Retry logic
- Gradual migration from traditional patterns

## Benchmarks

Run the benchmarks to see performance characteristics:

```bash
go test -bench=. -benchmem ./rusty/result/...
```

Typical results:
- Traditional error handling: ~100 ns/op
- Try/Catch pattern: ~150 ns/op (+50%)
- Functional composition: ~110 ns/op (+10%)
