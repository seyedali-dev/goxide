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

### Option[T] Examples

#### Basic Usage
```go
import "github.com/seyedali-dev/gopherbox/rusty/option"

// Creating Option values
someValue := option.Some(42)
noValue := option.None[int]()

// Checking presence
if someValue.IsSome() {
    fmt.Println("Value is present:", someValue.Unwrap())
}

if noValue.IsNone() {
    fmt.Println("No value present")
}
```

#### Safe Value Extraction
```go
// Using UnwrapOr with a default value
result := someValue.UnwrapOr(0)  // Returns 42
fallback := noValue.UnwrapOr(100)  // Returns 100

// Using UnwrapOrElse with a computation
expensiveDefault := option.None[int]().UnwrapOrElse(func() int {
    // Expensive computation only runs when value is absent
    return calculateDefault()
})

// Go-idiomatic extraction
var extracted int
if someValue.Some(&extracted) {
    fmt.Printf("Extracted: %d\n", extracted)
}
```

#### Functional Composition
```go
// Transforming values with Map
numbersOpt := option.Some(5)
stringOpt := option.Map(numbersOpt, func(n int) string {
    return fmt.Sprintf("Number: %d", n)
}) // Some("Number: 5")

// Chaining with FlatMap (when transformation can return Option)
userOpt := option.Some(User{Email: "user@example.com"})
emailOpt := option.FlatMap(userOpt, func(u User) option.Option[string] {
    if u.Email != "" {
        return option.Some(u.Email)
    }
    return option.None[string]()
})

// Type casting from interface{}
data := interface{}("hello")
stringOpt = option.Cast[string](data) // Some("hello")
intOpt = option.Cast[int](data)       // None[int]
```

### Result[T] Examples

#### Basic Error Handling
```go
import "github.com/seyedali-dev/gopherbox/rusty/result"

// Creating successful and error results
success := result.Ok(42)
failure := result.Err[int](fmt.Errorf("something went wrong"))

// Checking status
if success.IsOk() {
    value := success.Unwrap()
    fmt.Println("Success:", value)
}

if failure.IsErr() {
    err := failure.Err()
    fmt.Println("Error:", err)
}
```

#### Safe Value Extraction
```go
// With default values
value1 := success.UnwrapOr(0)     // Returns 42
value2 := failure.UnwrapOr(-1)    // Returns -1

// With error-based computation
value3 := failure.UnwrapOrElse(func(err error) int {
    log.Printf("Error occurred: %v", err)
    return -1
})

// Go-idiomatic error handling
var extracted int
if err := success.Ok(&extracted); err != nil {
    fmt.Printf("Error: %v\n", err)
} else {
    fmt.Printf("Extracted: %d\n", extracted)
}
```

#### Functional Composition
```go
// Transforming values with Map
numbersResult := result.Ok(5)
stringResult := result.Map(numbersResult, func(n int) string {
    return fmt.Sprintf("Number: %d", n)
}) // Ok("Number: 5")

// Chaining operations with FlatMap
processUser := func(id int) result.Result[User] {
    if id > 0 {
        return result.Ok(User{ID: id, Name: "John"})
    }
    return result.Err[User](fmt.Errorf("invalid user ID"))
}

userResult := result.FlatMap(numbersResult, processUser)

// Sequential operations with AndThen
validateEmail := func(email string) result.Result[string] {
    if strings.Contains(email, "@") {
        return result.Ok(email)
    }
    return result.Err[string](fmt.Errorf("invalid email"))
}

registerUser := func(email string) result.Result[User] {
    // Registration logic
    return result.Ok(User{Email: email})
}

// Chain operations safely
userResult := validateEmail("test@example.com").
    AndThen(registerUser)
```

#### Combining Multiple Results
```go
// Combining two results
configResult := LoadConfig()
dbResult := ConnectDB()
combined := result.Map2(configResult, dbResult, func(cfg Config, db DB) AppState {
    return AppState{Config: cfg, Database: db}
})

// Combining three results
userResult := GetUser(123)
permsResult := GetPermissions(123)
settingsResult := GetSettings(123)
fullProfile := result.Map3(userResult, permsResult, settingsResult, 
    func(u User, p Permissions, s Settings) UserProfile {
        return UserProfile{User: u, Permissions: p, Settings: s}
    })
```

#### Error Transformation
```go
// Transforming errors while preserving success values
userResult := FetchUser(123)
detailedResult := userResult.MapError(func(err error) error {
    return fmt.Errorf("failed to fetch user: %w", err)
})
```

### Reflection Examples

#### Basic Reflection
```go
import "github.com/seyedali-dev/gopherbox/reflect"

// Value comparison
equal := reflect.IsEqual(5, 5.0)        // true
equal = reflect.IsEqual("hello", "world") // false

// Checking for empty values
isEmpty := reflect.IsEmpty("")           // true
isEmpty = reflect.IsEmpty([]int{})       // true
isEmpty = reflect.IsEmpty(0)             // true
isEmpty = reflect.IsEmpty(false)         // true
isEmpty = reflect.IsEmpty("hello")       // false

// Type inference
val, err := reflect.InferType[int]("123")  // Attempts to convert string to int
if err != nil {
    // Handle conversion error
}
```

#### Struct Reflection - Field Access
```go
type Person struct {
    Name string `json:"name" validate:"required"`
    Age  int    `json:"age" validate:"min=0"`
}

person := Person{Name: "John", Age: 30}

// Get field by name
field := reflect.Field(person, "Name")
fmt.Println(field.Name)  // "Name"

// Get field value
fieldValue, ok := reflect.FieldValue(&person, "Name")
if ok {
    fmt.Println(fieldValue.String())  // "John"
}

// Set field value
err := reflect.FieldSet(&person, "Name", "Jane")
if err != nil {
    // Handle error
}
fmt.Println(person.Name)  // "Jane"
```

#### Struct Reflection - Tag Operations
```go
// Get tag value
tagValue := reflect.FieldTagValue(person, "Name", "json", "")  // "name"

// Check if field has multiple tags
hasTags := reflect.FieldHasTags(person, "Name", []string{"json", "validate"})  // true

// Get all tag keys for a field
tagKeys := reflect.FieldTagKeys(person, "Name")  // ["json", "validate"]

// Get tag key-value pair
key, value, found := reflect.FieldTagKeyValue(person, "Name", "validate", "")  // "validate", "required", true

// Find field by tag value
fieldName := reflect.FieldNameByTagValue(person, "validate", "required")  // "Name"
fieldNames := reflect.FieldNamesByTagValue(person, "json", "age")  // ["Age"]

// Check for specific tag value in multi-value tags
type Config struct {
    Field string `permissions:"read,write,admin"`
}
config := Config{}
hasPermission := reflect.FieldHasTagValue(config, "Field", "permissions", "admin", ",")  // true
```

#### Type-Safe Reflector Pattern
```go
// Creating a type-safe reflector for compile-time safety
personReflector := reflect.ForType[Person]()

// All operations are now type-safe
tagValue := personReflector.FieldTagValue("Name", "json", "")  // Compile-time check
hasTags := personReflector.FieldHasTags("Name", []string{"json", "validate"})  // Compile-time check
fieldInfo := personReflector.Field("Age")  // Compile-time check

// This would cause a compile error if "NonExistentField" doesn't exist in Person
// badTag := personReflector.FieldTagValue("NonExistentField", "json", "")
```

#### Advanced Struct Operations
```go
// Get all fields
person := Person{Name: "John", Age: 30}
fields := reflect.Fields(person)
for _, field := range fields {
    fmt.Printf("Field: %s, Type: %s\n", field.Name, field.Type)
}

// Get all field values
values := reflect.FieldValues(person)
for i, value := range values {
    fmt.Printf("Value %d: %v\n", i, value.Interface())
}

// Get struct type name
typeName := reflect.StructTypeName(person)  // "main.Person"
```

### Error Handling Examples

#### Using EnsureResult
```go
import "github.com/seyedali-dev/gopherbox/errors"

// Typical function that returns (value, error)
func GetUser(id int) (*User, error) {
    // Implementation that might return nil, err
    if id <= 0 {
        return nil, fmt.Errorf("invalid ID")
    }
    return &User{ID: id}, nil
}

// Using EnsureResult for consistent error handling
user, err := GetUser(123)
user, err = errors.EnsureResult(user, err, "user not found or invalid")

// If err is not nil, or user is nil, EnsureResult returns a new error with the message
if err != nil {
    // Handle error
    return err
}
```

### Real-World Use Cases

#### Option in Cache Operations
```go
type Cache struct {
    data map[string]interface{}
    mu   sync.RWMutex
}

func (c *Cache) Get(key string) option.Option[interface{}] {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if val, exists := c.data[key]; exists {
        return option.Some(val)
    }
    return option.None[interface{}]()
}

// Usage
cacheResult := cache.Get("user:123")
user, ok := cacheResult.Some(&User{})
if ok {
    fmt.Printf("Cache hit: %v\n", user)
} else {
    // Load from database
    user = loadUserFromDB(123)
    cache.Set("user:123", user)
}
```

#### Result in API Client Operations
```go
type APIClient struct {
    baseURL string
    client  *http.Client
}

func (c *APIClient) GetUser(id int) result.Result[User] {
    url := fmt.Sprintf("%s/users/%d", c.baseURL, id)
    resp, err := c.client.Get(url)
    if err != nil {
        return result.Err[User](fmt.Errorf("http request failed: %w", err))
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return result.Err[User](fmt.Errorf("api returned status: %d", resp.StatusCode))
    }
    
    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return result.Err[User](fmt.Errorf("json decode failed: %w", err))
    }
    
    return result.Ok(user)
}

// Chaining API calls safely
func (c *APIClient) GetUserWithPermissions(userID int) result.Result[UserWithPermissions] {
    return result.AndThen(c.GetUser(userID), func(user User) result.Result[UserWithPermissions] {
        permsResult := c.GetUserPermissions(userID)
        return result.Map(permsResult, func(perms []Permission) UserWithPermissions {
            return UserWithPermissions{User: user, Permissions: perms}
        })
    })
}
```

#### Complex Reflection with Struct Tags
```go
type APIRequest struct {
    UserID    int    `json:"user_id" validate:"required,min=1"`
    Name      string `json:"name" validate:"required,max=50"`
    Email     string `json:"email" validate:"required,email"`
    OptField  string `json:"opt_field,omitempty" validate:"max=100"`
}

func ValidateRequest(req interface{}) error {
    // Use reflection to validate based on struct tags
    v := reflect.ValueOf(req)
    t := reflect.TypeOf(req)
    
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
        t = t.Elem()
    }
    
    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)
        fieldType := t.Field(i)
        
        // Check required tags
        if reflect.FieldHasTagValue(req, fieldType.Name, "validate", "required", ",") {
            if reflect.IsEmpty(field.Interface()) {
                return fmt.Errorf("field %s is required", fieldType.Name)
            }
        }
        
        // Get validation rules
        rules, _ := reflect.FieldTagValues(req, fieldType.Name, "validate", ",")
        for _, rule := range rules {
            parts := strings.Split(rule, "=")
            switch parts[0] {
            case "min":
                if field.Kind() == reflect.Int {
                    min, _ := strconv.Atoi(parts[1])
                    if field.Int() < int64(min) {
                        return fmt.Errorf("field %s must be at least %d", fieldType.Name, min)
                    }
                }
            case "max":
                if field.Kind() == reflect.String || field.Kind() == reflect.Slice {
                    max, _ := strconv.Atoi(parts[1])
                    if field.Len() > max {
                        return fmt.Errorf("field %s exceeds max length %d", fieldType.Name, max)
                    }
                }
            }
        }
    }
    
    return nil
}

// Usage with type-safe reflector
requestReflector := reflect.ForType[APIRequest]()
// Compile-time safe access to field info
req := APIRequest{UserID: 1, Name: "John", Email: "john@example.com"}
if ValidateRequest(req) == nil {
    fmt.Println("Request is valid")
}
```

#### Configuration Loading with Results
```go
type Config struct {
    Port     int    `env:"PORT" default:"8080"`
    Database string `env:"DB_URL" validate:"required"`
    RedisURL string `env:"REDIS_URL"`
}

// Load configuration with multiple fallbacks
func LoadConfig() result.Result[Config] {
    // First try to load from environment
    envConfig, envErr := loadFromEnv()
    if envErr == nil && !reflect.IsEmpty(envConfig) {
        return result.Ok(envConfig)
    }
    
    // Then try from file
    fileConfig, fileErr := loadFromFile()
    if fileErr == nil && !reflect.IsEmpty(fileConfig) {
        return result.Ok(fileConfig)
    }
    
    // Finally use defaults
    defaultConfig := getDefaultConfig()
    return result.Ok(defaultConfig).MapError(func(err error) error {
        return fmt.Errorf("config loading failed. env: %v, file: %v, using defaults", 
            envErr, fileErr)
    })
}

// Using Map2 and Map3 for combining configuration sources
func LoadCompleteConfig() result.Result[CompleteConfig] {
    envResult := loadFromEnv()
    fileResult := loadFromFile()
    defaultResult := getDefaultConfig()
    
    return result.Map3(envResult, fileResult, defaultResult, 
        func(envConf Config, fileConf Config, defaultConf Config) CompleteConfig {
            // Merge configurations with priority: env > file > default
            return mergeConfigs(envConf, fileConf, defaultConf)
        })
}
```

#### Database Operations with Option and Result
```go
type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) FindByID(id int) option.Option[User] {
    var user User
    err := r.db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return option.None[User]()
        }
        // For other errors, you might want to return a Result instead
        log.Printf("Database error: %v", err)
        return option.None[User]()
    }
    
    return option.Some(user)
}

func (r *UserRepository) Create(user User) result.Result[User] {
    result, err := r.db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
        user.Name, user.Email)
    if err != nil {
        return result.Err[User](fmt.Errorf("failed to create user: %w", err))
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return result.Err[User](fmt.Errorf("failed to get inserted ID: %w", err))
    }
    
    newUser := user
    newUser.ID = int(id)
    return result.Ok(newUser)
}

// Combining repository operations
func (r *UserRepository) CreateUserIfNotExists(email string) result.Result[User] {
    // Check if user exists
    existingUserOpt := r.FindByEmail(email)
    if existingUserOpt.IsSome() {
        return result.Ok(existingUserOpt.Unwrap()) // Return existing user
    }
    
    // Create new user
    newUser := User{Name: extractNameFromEmail(email), Email: email}
    return r.Create(newUser)
}
```
