# Types Package - Generic Functional Helpers

<!-- TOC -->
* [Types Package - Generic Functional Helpers](#types-package---generic-functional-helpers)
  * [Overview](#overview)
  * [Quick Start](#quick-start)
    * [Installation](#installation)
  * [Core Functions](#core-functions)
    * [Identity Function](#identity-function)
      * [`Id[T any](t T) T`](#idt-anyt-t-t)
    * [Constant Functions](#constant-functions)
      * [`Return[In any, T any](t T) func(In) T`](#returnin-any-t-anyt-t-funcin-t)
      * [`Return0[T any](t T) func() T`](#return0t-anyt-t-func-t)
      * [`Value[T any]() T`](#valuet-any-t)
    * [Function Composition](#function-composition)
      * [`Compose[T, U, V any](fn1 func(T) U, fn2 func(U) V) func(T) V`](#composet-u-v-anyfn1-funct-u-fn2-funcu-v-funct-v)
  * [Usage Patterns](#usage-patterns)
    * [Pattern 1: Default Placeholder Functions](#pattern-1-default-placeholder-functions)
    * [Pattern 2: Lazy Default Values](#pattern-2-lazy-default-values)
    * [Pattern 3: Function Pipelines](#pattern-3-function-pipelines)
    * [Pattern 4: Higher-Order Function Parameters](#pattern-4-higher-order-function-parameters)
  * [Integration with Result and Option](#integration-with-result-and-option)
    * [With Result Type](#with-result-type)
    * [With Option Type](#with-option-type)
  * [API Reference](#api-reference)
    * [Identity and Values](#identity-and-values)
      * [`Id[T any](t T) T`](#idt-anyt-t-t-1)
      * [`Value[T any]() T`](#valuet-any-t-1)
    * [Constant Functions](#constant-functions-1)
      * [`Return[In any, T any](t T) func(In) T`](#returnin-any-t-anyt-t-funcin-t-1)
      * [`Return0[T any](t T) func() T`](#return0t-anyt-t-func-t-1)
    * [Function Composition](#function-composition-1)
      * [`Compose[T, U, V any](fn1 func(T) U, fn2 func(U) V) func(T) V`](#composet-u-v-anyfn1-funct-u-fn2-funcu-v-funct-v-1)
  * [Best Practices](#best-practices)
    * [✅ DO: Use for Higher-Order Functions](#-do-use-for-higher-order-functions)
    * [✅ DO: Use Compose for Readable Pipelines](#-do-use-compose-for-readable-pipelines)
    * [✅ DO: Use Return0 for Lazy Evaluation](#-do-use-return0-for-lazy-evaluation)
    * [❌ DON'T: Overuse in Simple Cases](#-dont-overuse-in-simple-cases)
    * [❌ DON'T: Sacrifice Readability](#-dont-sacrifice-readability)
  * [Examples](#examples)
<!-- TOC -->

## Overview

The `types` package provides small, generic functional helpers inspired by functional programming patterns. These utilities are designed to make higher-order function composition and default handling more expressive in Go.

**Philosophy**: These helpers are most useful in generic libraries or when building composable abstractions. Avoid them when plain inline functions are clearer — Go favors explicitness.

## Quick Start

### Installation

```go
import "github.com/seyedali-dev/gopherbox/rusty/types"
```

## Core Functions

### Identity Function

#### `Id[T any](t T) T`

Returns its input unchanged. Useful as a default function or placeholder in higher-order contexts.

```go
// Basic usage
x := types.Id(42)           // x == 42
name := types.Id("hello")   // name == "hello"

// As a default transformation
result := processValue(42, types.Id[int])  // No transformation
```

### Constant Functions

#### `Return[In any, T any](t T) func(In) T`

Creates a function that ignores its input and always returns `t`. Use when you need a constant function of type `func(In) T`.

```go
// Create constant functions
alwaysFive := types.Return[string, int](5)
fmt.Println(alwaysFive("ignored"))  // 5

alwaysHello := types.Return[int, string]("hello")
fmt.Println(alwaysHello(123))       // "hello"
```

#### `Return0[T any](t T) func() T`

Creates a zero-argument function that always returns `t`. Useful for lazy defaults or callbacks that require `func() T`.

```go
// Lazy evaluation
lazyConfig := types.Return0(loadDefaultConfig())
// Config not loaded until called

config := lazyConfig()  // Now loadDefaultConfig() is called

// As a fallback provider
fallback := types.Return0("default-value")
```

#### `Value[T any]() T`

Returns the zero value of type `T`. Use when you need an explicit zero value without constructing it manually.

```go
// Get zero values
var zeroInt int = types.Value[int]()        // 0
var zeroStr string = types.Value[string]()  // ""
var zeroSlice []int = types.Value[[]int]()  // nil

// Useful in generic contexts
func Initialize[T any]() T {
    return types.Value[T]()  // Returns appropriate zero value
}
```

### Function Composition

#### `Compose[T, U, V any](fn1 func(T) U, fn2 func(U) V) func(T) V`

Chains two functions: `fn1` followed by `fn2`. Returns a new function that applies `fn1`, then `fn2`.

```go
// Create composed functions
trimAndUpper := types.Compose(strings.TrimSpace, strings.ToUpper)
result := trimAndUpper("  hello  ")  // "HELLO"

parseAndValidate := types.Compose(parseNumber, validatePositive)
result := parseAndValidate("42")     // Result after parsing and validation
```

## Usage Patterns

### Pattern 1: Default Placeholder Functions

Use `Id` as a default when callers can provide custom transformations:

```go
type Processor struct {
    transform func(string) string
}

func NewProcessor(transform func(string) string) *Processor {
    if transform == nil {
        transform = types.Id[string]  // Default: no transformation
    }
    return &Processor{transform: transform}
}

// Usage - with custom transform
p1 := NewProcessor(strings.ToUpper)
// Usage - with default (no transform)
p2 := NewProcessor(nil)  // Uses types.Id internally
```

### Pattern 2: Lazy Default Values

Use `Return0` for expensive defaults that should only be computed when needed:

```go
type Service struct {
    loadConfig func() Config
}

func NewService(configProvider func() Config) *Service {
    if configProvider == nil {
        // Lazy loading of default config
        configProvider = types.Return0(loadDefaultConfig())
    }
    return &Service{loadConfig: configProvider}
}

// Config not loaded until first use
service := NewService(nil)
// ... later, when config is needed
config := service.loadConfig()  // loadDefaultConfig() called here
```

### Pattern 3: Function Pipelines

Use `Compose` to build reusable transformation pipelines:

```go
// Build data processing pipeline
var processInput = types.Compose(
    strings.TrimSpace,      // Step 1: trim
    strings.ToLower,        // Step 2: lowercase  
    func(s string) string { // Step 3: custom logic
        if s == "" {
            return "default"
        }
        return s
    },
)

result := processInput("  HELLO  ")  // "hello"
result2 := processInput("   ")       // "default"
```

### Pattern 4: Higher-Order Function Parameters

Use the helpers when working with functions that accept other functions:

```go
func TransformSlice[T any](slice []T, transform func(T) T) []T {
    if transform == nil {
        transform = types.Id[T]  // Default: identity
    }
    
    result := make([]T, len(slice))
    for i, v := range slice {
        result[i] = transform(v)
    }
    return result
}

// Usage with custom transform
numbers := []int{1, 2, 3}
doubled := TransformSlice(numbers, func(x int) int { return x * 2 })
// [2, 4, 6]

// Usage with default (no transform)
same := TransformSlice(numbers, nil)  // Uses types.Id
// [1, 2, 3]
```

## Integration with Result and Option

### With Result Type

The helpers work seamlessly with the `Result` type:

```go
import (
    "github.com/seyedali-dev/gopherbox/rusty/result"
    "github.com/seyedali-dev/gopherbox/rusty/types"
)

// Identity in Result transformations
userResult := result.Map(findUser(123), types.Id[User])
// Equivalent to not mapping at all, but useful in generic code

// Constant functions for fallbacks
defaultUser := types.Return[error, User](User{Name: "Guest"})
user := userResult.UnwrapOrElse(defaultUser)

// Composition with Result operations
processUser := types.Compose(validateUser, result.WrapFunc1(saveUser))
// Returns func(User) Result[User]
```

### With Option Type

Similarly, they work well with `Option`:

```go
import (
    "github.com/seyedali-dev/gopherbox/rusty/option"
    "github.com/seyedali-dev/gopherbox/rusty/types"
)

// Identity in Option transformations
nameOpt := option.Map(userOpt, types.Id[string])

// Constant fallbacks
fallback := types.Return0("default-name")
name := nameOpt.UnwrapOrElse(fallback)

// Composition pipeline
formatName := types.Compose(
    strings.TrimSpace,
    strings.Title,
    func(s string) string { return "Mr. " + s },
)

formattedOpt := option.Map(nameOpt, formatName)
```

## API Reference

### Identity and Values

#### `Id[T any](t T) T`
Returns the input value unchanged.

**Usage:**
```go
x := types.Id(42)        // 42
s := types.Id("hello")   // "hello"
```

#### `Value[T any]() T`
Returns the zero value for type T.

**Usage:**
```go
zero := types.Value[int]()        // 0
empty := types.Value[string]()    // ""
nilSlice := types.Value[[]int]()  // nil
```

### Constant Functions

#### `Return[In any, T any](t T) func(In) T`
Creates a function that ignores input and returns constant value.

**Usage:**
```go
always42 := types.Return[string, int](42)
result := always42("ignore")  // 42

alwaysHi := types.Return[int, string]("hi")
result := alwaysHi(123)       // "hi"
```

#### `Return0[T any](t T) func() T`
Creates a zero-argument function that returns constant value.

**Usage:**
```go
getConfig := types.Return0(loadConfig())
config := getConfig()  // loadConfig() is called here
```

### Function Composition

#### `Compose[T, U, V any](fn1 func(T) U, fn2 func(U) V) func(T) V`
Chains two functions together.

**Usage:**
```go
trimUpper := types.Compose(strings.TrimSpace, strings.ToUpper)
result := trimUpper("  hello  ")  // "HELLO"
```

## Best Practices

### ✅ DO: Use for Higher-Order Functions

```go
// ✅ Useful in generic higher-order functions
func WithFallback[T any](primary func() T, fallback func() T) func() T {
    if primary == nil {
        primary = types.Return0(types.Value[T]())
    }
    return func() T {
        // implementation
    }
}

// ❌ Overkill for simple cases
func AddOne(x int) int {
    return types.Id(x) + 1  // Just use x + 1
}
```

### ✅ DO: Use Compose for Readable Pipelines

```go
// ✅ Clear transformation pipeline
var processUserInput = types.Compose(
    sanitizeInput,
    validateInput, 
    normalizeInput,
)

// ❌ Manual composition (harder to read)
func processUserInput(input string) string {
    return normalizeInput(validateInput(sanitizeInput(input)))
}
```

### ✅ DO: Use Return0 for Lazy Evaluation

```go
// ✅ Lazy expensive operations
func NewService() *Service {
    return &Service{
        configLoader: types.Return0(loadExpensiveConfig()),
    }
}

// ❌ Eager loading (wasteful)
func NewService() *Service {
    config := loadExpensiveConfig()  // Loaded even if never used
    return &Service{
        configLoader: func() Config { return config },
    }
}
```

### ❌ DON'T: Overuse in Simple Cases

```go
// ✅ Simple and clear
func GetName(user User) string {
    return user.Name
}

// ❌ Unnecessary abstraction
func GetName(user User) string {
    return types.Id(user.Name)
}
```

### ❌ DON'T: Sacrifice Readability

```go
// ✅ Clear and explicit
func Process(data string) string {
    trimmed := strings.TrimSpace(data)
    lower := strings.ToLower(trimmed)
    return strconv.Quote(lower)
}

// ❌ Overly abstracted (hard to follow)
var Process = types.Compose(
    strings.TrimSpace,
    strings.ToLower,
    strconv.Quote,
)
```

## Examples

Here are practical examples from common use cases:

```go
// Configuration with lazy defaults
type AppConfig struct {
    // ... fields
}

func loadDefaultConfig() AppConfig {
    // Expensive operation
    return AppConfig{/* ... */}
}

func NewApp(configProvider func() AppConfig) *App {
    if configProvider == nil {
        configProvider = types.Return0(loadDefaultConfig())
    }
    return &App{configProvider: configProvider}
}

// Data processing pipeline
var cleanAndValidate = types.Compose(
    strings.TrimSpace,
    strings.ToLower,
    func(s string) string {
        if len(s) > 100 {
            return s[:100]
        }
        return s
    },
)

// Generic processor with identity default
func ProcessItems[T any](items []T, processor func(T) T) []T {
    if processor == nil {
        processor = types.Id[T]
    }
    
    result := make([]T, len(items))
    for i, item := range items {
        result[i] = processor(item)
    }
    return result
}
```

These functional helpers provide the building blocks for creating expressive, composable APIs while maintaining Go's philosophy of simplicity and clarity.
