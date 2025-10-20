# gopherbox
A lightweight collection of Go utility functions for reflection, value checks, error handling, and functional programming patterns with Rust-inspired types.

## Features

### Reflection Utilities (`reflect` package)
- **General Reflection**: Functions like `IsEqual`, `IsEmpty`, `InferType` for working with Go values and types.
- **Struct Reflection**: Advanced struct field manipulation including tag parsing, field access, and modification.
- **Type-Safe Wrappers**: Optional type-safe reflector pattern that provides compile-time safety for struct operations.

### Error Handling (`errors` package)
- **EnsureResult**: Enforces consistent error handling patterns for functions that return (value, error).
- **WrapNilError**: Handles both error checking and "empty" value validation in one call.

### Rust-Style Types (`rusty` package)
- **Option[T]**: A Rust-like optional type to handle values that might be absent without nil checks.
- **Result[T]**: A Rust-like type for operations that can fail, enabling functional error handling.
- **Functional Helpers**: Generic utilities like `Id`, `Compose`, and `Return` for functional programming patterns.

## Packages

### `reflect`
Utilities for Go's reflection system, including:
- Value comparison and emptiness checks
- Type inference from interface{}
- Struct field manipulation, tag parsing, and type-safe reflector patterns

### `errors`
Error handling utilities that provide:
- Consistent error and nil checking patterns
- Centralized error handling logic

### `rusty/option`
Rust-like optional values with methods like:
- `Some(value)` and `None()` constructors
- `IsSome()`, `IsNone()` value checking
- `Map()`, `FlatMap()` functional composition
- `Unwrap()`, `UnwrapOr()`, `UnwrapOrElse()` safe value extraction

### `rusty/result`
Rust-like result type for error handling with methods like:
- `Ok(value)` and `Err(error)` constructors
- `Map()`, `FlatMap()`, `AndThen()` for error propagation
- `Unwrap()`, `UnwrapOr()`, `UnwrapOrElse()` safe value extraction

### `rusty/types`
Generic functional programming helpers including:
- `Id`, `Compose`, `Return` for function manipulation
- Utility functions for creating zero-value functions

## Installation

```bash
go get github.com/seyedali-dev/gopherbox
```

## Usage Examples

### Option[T] Example
```go
import "github.com/seyedali-dev/gopherbox/rusty/option"

// Safe optional value handling
userOpt := FindUserByEmail("user@example.com")
if userOpt.IsSome() {
    user := userOpt.Unwrap()
    fmt.Printf("Found user: %s", user.Name)
} else {
    fmt.Println("User not found")
}

// Functional chaining
emailOpt := option.Map(userOpt, func(u User) string {
    return u.Email
})
```

### Result[T] Example
```go
import "github.com/seyedali-dev/gopherbox/rusty/result"

// Safe error handling
userResult := repo.FindByID(123)
if userResult.IsOk() {
    user := userResult.Unwrap()
    // Process user
} else {
    err := userResult.Err()
    // Handle error
}

// Functional chaining
result := ValidateEmail(email).
    AndThen(func(validEmail string) result.Result[User] {
        return CreateUser(validEmail, password)
    })
```

### Reflection Example
```go
import "github.com/seyedali-dev/gopherbox/reflect"

// Check if values are equal
if reflect.IsEqual(5, 5) {
    fmt.Println("Values are equal")
}

// Check if a value is empty
if reflect.IsEmpty("") {
    fmt.Println("String is empty")
}

// Infer type from interface{}
val, err := reflect.InferType[int]("123")
```
