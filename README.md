# GopherBox - Go Utilities Library

![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue)
![License](https://img.shields.io/badge/License-MIT-green)
![Status](https://img.shields.io/badge/Status-Production%20Ready-brightgreen)

A comprehensive Go utilities library inspired by Rust's safety and expressiveness, providing robust error handling, reflection utilities, and functional programming patterns.

## 📦 Packages Overview

### Core Utilities
- **[`errors`](./errors)**: Enhanced error utilities with nil-safe handling
- **[`reflect`](./reflect)**: Type-safe reflection utilities for struct operations

### Rust-Inspired Patterns (`rusty` package)
- **[`result`](./rusty/result/README_RESULT.md)**: Rust-like Result type with error bubble up (Try/Catch) patterns (rust's `?` equivalent)
- **[`option`](./rusty/option/README_OPTION.md)**: Optional value handling without nil panics
- **[`chain`](./rusty/chain/README_CHAIN.md)**: Fluent method chaining for Result and Option
- **[`types`](./rusty/types)**: Generic functional programming helpers

## 🚀 Quick Start

### Installation

```bash
go get github.com/seyedali-dev/gopherbox
```

### Basic Usage

```go
import (
    "github.com/seyedali-dev/gopherbox/rusty/result"
    "github.com/seyedali-dev/gopherbox/rusty/option"
    "github.com/seyedali-dev/gopherbox/reflect"
)

// Result pattern for error handling
func GetUser(id int) (res result.Result[User]) {
    defer result.Catch(&res)
    
    user := db.FindUser(id).Try() // Early return on error
    profile := db.FindProfile(user.ID).Try() // ? - bubbleup error and return
    
    return result.Ok(profile)
}

// Option pattern for optional values
func GetUserName(userID int) string {
    userOpt := cache.GetUser(userID)
    return userOpt.UnwrapOr("Guest")
}

// Reflection utilities
func GetStructTags(user User) []string {
    return reflect.FieldTagKeys(user, "Name")
}
```

## 🎯 Key Features

### 🔒 Type Safety
- Compiler-enforced error handling
- No nil pointer dereferences
- Explicit optional values

### 🛠️ Error Handling
- **Rust-like Result type** with `Try()` for early returns
- **Error recovery** with `CatchWith` and `Fallback`
- **Functional composition** with `Map`, `AndThen`, `FlatMap`

### 🔍 Reflection Made Safe
- **Type-safe struct operations**
- **Compile-time field validation**
- **Struct tag parsing and validation**

### 🔗 Fluent APIs
- **Method chaining** for complex operations
- **Pipeline-style programming**
- **Readable sequential operations**

## 📚 Package Details

### [Result Package](./rusty/result/README_RESULT.md)
Rust-inspired error handling with early returns and error recovery patterns.

**Features:**
- `Try()` method equivalent to Rust's `?` operator
- Error-specific recovery with `CatchWith`
- Functional composition with `Map` and `AndThen`
- Multi-error combination with `Map2` and `Map3`

**Example:**
```go
func ProcessOrder(orderID int) (res result.Result[Receipt]) {
    defer result.Catch(&res)
    
    order := FindOrder(orderID).Try()
    payment := ProcessPayment(order).Try()
    receipt := GenerateReceipt(payment).Try()
    
    return result.Ok(receipt)
}
```

### [Option Package](./rusty/option/README_OPTION.md)
Safe optional value handling without nil pointer panics.

**Features:**
- Explicit Some/None semantics
- Safe value extraction with fallbacks
- Functional transformation with `Map` and `FlatMap`
- Type-safe optional operations

**Example:**
```go
func GetUserEmail(userID int) option.Option[string] {
    userOpt := cache.GetUser(userID)
    return option.Map(userOpt, func(u User) string {
        return u.Email
    })
}
```

### [Chain Package](./rusty/chain/README_CHAIN.md) (work in progress)
Fluent method chaining for Result and Option types.

**Features:**
- Pipeline-style operation sequencing
- Type-safe transformation chains
- No nested Map/AndThen calls
- Better readability for complex operations

**Example:**
```go
chain.Chain(findUser(123)).
    Map(func(u User) string { return u.Name }).
    AndThen(validateName)
```

### [Reflect Package](./reflect)
Type-safe reflection utilities for struct operations.

**Features:**
- Struct field introspection
- Tag parsing and validation
- Type-safe field access
- Compile-time safety with generics

**Example:**
```go
// Type-safe reflector
userReflector := reflect.ForType[User]()
tagValue := userReflector.FieldTagValue("Name", "json")

// Traditional usage
tags := reflect.FieldTagKeys(user, "Name")
```

### [Errors Package](./errors)
Enhanced error utilities with nil-safe handling.

**Features:**
- Generic error wrapping
- Nil value detection
- Consistent error patterns

**Example:**
```go
user, err := errors.EnsureResult(
    db.FindUser(123), 
    "user not found"
)
```

## 🏗️ Architecture Principles

### 1. **Explicit Over Implicit**
- No hidden nil checks
- Clear error propagation
- Explicit optional values

### 2. **Type Safety First**
- Compiler-enforced patterns
- Generic type constraints
- Runtime safety guarantees

### 3. **Multiple Patterns**
- Choose between traditional, functional, or early-return styles
- Gradual adoption path
- No lock-in to single approach

### 4. **Performance Conscious**
- Zero allocations in happy paths
- Minimal overhead over traditional patterns
- Benchmark-driven optimizations

## 📖 Examples

Comprehensive examples are available in the [`examples`](./rusty/examples) package:

- [Database operations with fallbacks](./rusty/examples/examples.go)
- [HTTP handlers with error handling](./rusty/examples/examples.go)
- [Validation chains](./rusty/examples/examples.go)
- [Transaction handling](./rusty/examples/examples.go)

## 🔧 Migration Guide

### From Traditional Go

**Before:**
```go
func GetUserData(id int) (UserData, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return UserData{}, err
    }
    
    profile, err := db.FindProfile(user.ID)
    if err != nil {
        return UserData{}, err
    }
    
    return ProcessData(user, profile), nil
}
```

**After:**
```go
func GetUserData(id int) (res result.Result[UserData]) {
    defer result.Catch(&res)
    
    user := db.FindUser(id).Try()
    profile := db.FindProfile(user.ID).Try()
    
    return result.Ok(ProcessData(user, profile))
}
```

### Gradual Adoption

Wrap existing functions without changing signatures:

```go
var findUser = result.WrapFunc1(db.FindUser)
var loadConfig = result.WrapFunc(config.Load)

// Use new patterns incrementally
func MixedUsage(id int) (User, error) {
    var user User
    var err error
    defer result.CatchErr(&user, &err)
    
    config := loadConfig().Try()
    user = findUser(id).Try()
    
    return user, nil
}
```

## 📊 Performance

Benchmarks show minimal overhead:

```
Traditional error handling:   100 ns/op
Result with Try/Catch:        150 ns/op (+50%)
Result with AndThen:          110 ns/op (+10%)
Option operations:            5-10 ns/op
```

**Recommendations:**
- Use `Try()` for business logic where clarity matters
- Use traditional patterns in performance-critical loops
- The readability benefit usually outweighs the small cost

## 🤝 Contributing

You're welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
git clone https://github.com/seyedali-dev/gopherbox
cd gopherbox
go test ./...
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Portions of this software are derived from work licensed under the Apache License 2.0 and MIT License.  
See [THIRD_PARTY_LICENSES](./THIRD_PARTY_LICENSES) for the full license text.

## 🙏 Acknowledgments

Inspired by:
- **Rust**'s `Result` and `Option` types
- **Functional programming** patterns
- **Go**'s simplicity and pragmatism
- The Go community's best practices

## 📞 Support

- 📧 **Email**: [seyedali.dev@gmail.com](mailto:seyedali.dev@gmail.com)
- 🐛 **Issues**: [GitHub Issues](https://github.com/seyedali-dev/gopherbox/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/seyedali-dev/gopherbox/discussions)

## 🚀 Roadmap

- [ ] **v1.0**: Enhanced chaining
- [ ] **v1.1**: Enhanced collection utilities
- [ ] **v1.2**: Async/await patterns for Go
- [ ] **v1.3**: Database integration helpers

---

<div align="center">

**Built with ❤️ for the Go community**

*Making Go development safer, more expressive, and more enjoyable*

</div>
