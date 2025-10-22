// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package examples. examples demonstrates practical usage patterns for the enhanced Result type.
package examples

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/seyedali-dev/gopherbox/rusty/result"
)

// -------------------------------------------- Domain Types --------------------------------------------

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Order struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type Config struct {
	DatabaseURL string
	CacheURL    string
	APITimeout  time.Duration
}

// -------------------------------------------- Error Definitions --------------------------------------------

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrInvalidEmail   = errors.New("invalid email format")
	ErrDatabaseDown   = errors.New("database unavailable")
	ErrCacheMiss      = errors.New("cache miss")
	ErrAPITimeout     = errors.New("API request timeout")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrInvalidInput   = errors.New("invalid input")
	ErrConfigNotFound = errors.New("configuration not found")
)

// -------------------------------------------- Example 1: Simple Database Query with Fallback --------------------------------------------

// FindUserWithFallback demonstrates using Try() with CatchWith for cache fallback.
// If database fails, it automatically tries the cache.
func FindUserWithFallback(db *sql.DB, cache Cache, userID int) (res result.Result[User]) {
	defer result.Catch(&res)

	// If database fails, try cache
	defer result.CatchWith(&res, func(err error) User {
		return cache.GetUser(userID).Try()
	}, ErrDatabaseDown)

	// Try database first
	var user User
	err := db.QueryRow("SELECT id, email, name, created_at FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return result.Err[User](ErrUserNotFound)
	}
	if err != nil {
		return result.Err[User](ErrDatabaseDown)
	}

	return result.Ok(user)
}

// -------------------------------------------- Example 2: Multi-Step Order Processing --------------------------------------------

// ProcessOrderPipeline demonstrates chaining multiple operations with Try().
// Each step can fail and will automatically propagate errors.
func ProcessOrderPipeline(db *sql.DB, orderID int) (res result.Result[string]) {
	defer result.Catch(&res)

	// Each Try() call will early-return if there's an error
	order := fetchOrder(db, orderID).Try()
	user := fetchUser(db, order.UserID).Try()
	validateOrderAmount(order).Try()
	payment := chargePayment(user, order).Try()
	receipt := generateReceipt(payment).Try()

	return result.Ok(receipt)
}

// -------------------------------------------- Example 3: Configuration Loading with Defaults --------------------------------------------

// LoadConfiguration demonstrates using Fallback for default values.
func LoadConfiguration(configPath string) (res result.Result[Config]) {
	defer result.Catch(&res)

	// Use default config if file not found
	defer result.Fallback(&res, Config{
		DatabaseURL: "localhost:5432",
		CacheURL:    "localhost:6379",
		APITimeout:  30 * time.Second,
	}, ErrConfigNotFound)

	config := result.Wrap(loadConfigFromFile(configPath)).Try()
	return result.Ok(config)
}

// -------------------------------------------- Example 4: Multi-Layer Data Fetching --------------------------------------------

// FetchDataMultiLayer demonstrates cascading fallbacks through multiple data sources.
// Tries: Memory -> Local Cache -> Database -> Remote API -> Default
func FetchDataMultiLayer(id int) (res result.Result[string]) {
	defer result.Catch(&res)

	// Ultimate fallback for any error
	defer result.Fallback(&res, "default-value")

	// Try remote API if database fails
	defer result.CatchWith(&res, func(err error) string {
		return result.Wrap(fetchFromRemoteAPI(id)).Try()
	}, ErrDatabaseDown)

	// Try database if cache fails
	defer result.CatchWith(&res, func(err error) string {
		return result.Wrap(fetchFromDatabase(id)).Try()
	}, ErrCacheMiss)

	// Try local cache if memory fails
	defer result.CatchWith(&res, func(err error) string {
		return result.Wrap(fetchFromCache(id)).Try()
	}, sql.ErrNoRows)

	// Try memory first (fastest)
	return result.Wrap(fetchFromMemory(id))
}

// -------------------------------------------- Example 5: HTTP Handler with CatchErr --------------------------------------------

// HandleGetUser demonstrates using CatchErr to adapt to traditional (value, error) signatures.
// Useful for HTTP handlers and interface implementations.
func HandleGetUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		var err error

		// CatchErr adapts Result to (value, error)
		defer result.CatchErr(&user, &err)

		userID := extractUserID(r).Try()
		user = fetchUser(db, userID).Try()

		if err != nil {
			handleHTTPError(w, err)
			return
		}

		json.NewEncoder(w).Encode(user)
	}
}

// -------------------------------------------- Example 6: Validation Chain --------------------------------------------

// ValidateUserInput demonstrates chaining validations with Try().
func ValidateUserInput(email, password, username string) (res result.Result[User]) {
	defer result.Catch(&res)

	validEmail := validateEmail(email).Try()
	_ = validatePassword(password).Try()
	validUsername := validateUsername(username).Try()

	return result.Ok(User{
		Email: validEmail,
		Name:  validUsername,
	})
}

// -------------------------------------------- Example 7: Transaction Handling --------------------------------------------

// ExecuteTransaction demonstrates error handling in database transactions.
func ExecuteTransaction(db *sql.DB, userID int, amount float64) (res result.Result[string]) {
	defer result.Catch(&res)

	tx := result.Wrap(db.Begin()).Try()

	// Rollback on any error
	defer func() {
		if res.IsErr() {
			tx.Rollback()
		}
	}()

	// Execute transaction steps
	updateBalance(tx, userID, amount).Try()
	recordTransaction(tx, userID, amount).Try()

	//result.Wrap(nil, tx.Commit()).Try()
	//return result.Ok("transaction completed")
	panic("TODO: fix my error")
}

// -------------------------------------------- Example 8: Context-Aware Operations --------------------------------------------

// FetchWithTimeout demonstrates context cancellation handling.
func FetchWithTimeout(ctx context.Context, url string) (res result.Result[[]byte]) {
	defer result.Catch(&res)

	// Check context before starting
	if ctx.Err() != nil {
		return result.Err[[]byte](ctx.Err())
	}

	req := result.Wrap(http.NewRequestWithContext(ctx, "GET", url, nil)).Try()

	client := &http.Client{Timeout: 30 * time.Second}
	resp := result.Wrap(client.Do(req)).Try()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result.Err[[]byte](fmt.Errorf("HTTP %d", resp.StatusCode))
	}

	data := result.Wrap(io.ReadAll(resp.Body)).Try()
	return result.Ok(data)
}

// -------------------------------------------- Example 9: Retry Logic --------------------------------------------

// FetchWithRetry demonstrates implementing retry logic with error handlers.
func FetchWithRetry(url string, maxRetries int) (res result.Result[[]byte]) {
	var attempts int

	for attempts = 0; attempts < maxRetries; attempts++ {
		res = fetchData(url)
		if res.IsOk() {
			return res
		}

		// Wait before retry
		time.Sleep(time.Duration(attempts+1) * time.Second)
	}

	return res // Return last error
}

// -------------------------------------------- Example 10: Combining Traditional and Result Patterns --------------------------------------------

// MigrateToResult shows gradual migration from traditional error handling.
func MigrateToResult(db *sql.DB, userID int) (res result.Result[User]) {
	defer result.Catch(&res)

	// Traditional function call - wrap with result.Wrap
	user, err := legacyFetchUser(db, userID)
	if err != nil {
		return result.Err[User](err)
	}

	// New Result-based function - use Try()
	enrichedUser := enrichUserData(user).Try()

	return result.Ok(enrichedUser)
}

// -------------------------------------------- Helper Functions --------------------------------------------

type Cache interface {
	GetUser(id int) result.Result[User]
	SetUser(user User) result.Result[bool]
}

func fetchOrder(db *sql.DB, orderID int) result.Result[Order] {
	var order Order
	err := db.QueryRow("SELECT id, user_id, total_price, status FROM orders WHERE id = ?", orderID).
		Scan(&order.ID, &order.UserID, &order.TotalPrice, &order.Status)
	return result.Wrap(order, err)
}

func fetchUser(db *sql.DB, userID int) result.Result[User] {
	var user User
	err := db.QueryRow("SELECT id, email, name FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Email, &user.Name)
	return result.Wrap(user, err)
}

func validateOrderAmount(order Order) result.Result[Order] {
	if order.TotalPrice <= 0 {
		return result.Err[Order](ErrInvalidInput)
	}
	return result.Ok(order)
}

func chargePayment(user User, order Order) result.Result[string] {
	// Simulate payment processing
	return result.Ok(fmt.Sprintf("payment-%d", order.ID))
}

func generateReceipt(paymentID string) result.Result[string] {
	return result.Ok(fmt.Sprintf("receipt-%s", paymentID))
}

func loadConfigFromFile(path string) (Config, error) {
	// Simulate config loading
	return Config{}, ErrConfigNotFound
}

func fetchFromMemory(id int) (string, error) {
	return "", sql.ErrNoRows
}

func fetchFromCache(id int) (string, error) {
	return "", ErrCacheMiss
}

func fetchFromDatabase(id int) (string, error) {
	return "", ErrDatabaseDown
}

func fetchFromRemoteAPI(id int) (string, error) {
	return fmt.Sprintf("data-%d", id), nil
}

func extractUserID(r *http.Request) result.Result[int] {
	// Simulate ID extraction
	return result.Ok(123)
}

func handleHTTPError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
	case errors.Is(err, ErrUnauthorized):
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func validateEmail(email string) result.Result[string] {
	if email == "" {
		return result.Err[string](ErrInvalidEmail)
	}
	return result.Ok(email)
}

func validatePassword(password string) result.Result[string] {
	if len(password) < 8 {
		return result.Err[string](ErrInvalidInput)
	}
	return result.Ok(password)
}

func validateUsername(username string) result.Result[string] {
	if username == "" {
		return result.Err[string](ErrInvalidInput)
	}
	return result.Ok(username)
}

func updateBalance(tx *sql.Tx, userID int, amount float64) result.Result[bool] {
	_, err := tx.Exec("UPDATE accounts SET balance = balance + ? WHERE user_id = ?", amount, userID)
	return result.Wrap(true, err)
}

func recordTransaction(tx *sql.Tx, userID int, amount float64) result.Result[bool] {
	_, err := tx.Exec("INSERT INTO transactions (user_id, amount) VALUES (?, ?)", userID, amount)
	return result.Wrap(true, err)
}

func fetchData(url string) result.Result[[]byte] {
	//return result.Wrap(http.Get(url))
	panic("implement me and fix my error")
}

func legacyFetchUser(db *sql.DB, userID int) (User, error) {
	var user User
	err := db.QueryRow("SELECT id, email, name FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Email, &user.Name)
	return user, err
}

func enrichUserData(user User) result.Result[User] {
	// Simulate data enrichment
	return result.Ok(user)
}

// -------------------------------------------- Example 11: Comparison of Patterns --------------------------------------------

// TraditionalStyle shows standard Go error handling.
func TraditionalStyle(db *sql.DB, orderID int) (string, error) {
	order, err := fetchOrderTraditional(db, orderID)
	if err != nil {
		return "", err
	}

	user, err := fetchUserTraditional(db, order.UserID)
	if err != nil {
		return "", err
	}

	if order.TotalPrice <= 0 {
		return "", ErrInvalidInput
	}

	payment, err := chargePaymentTraditional(user, order)
	if err != nil {
		return "", err
	}

	receipt, err := generateReceiptTraditional(payment)
	if err != nil {
		return "", err
	}

	return receipt, nil
}

// ResultStyleWithTry shows the new Try() pattern (recommended for sequential operations).
func ResultStyleWithTry(db *sql.DB, orderID int) (res result.Result[string]) {
	defer result.Catch(&res)

	order := fetchOrder(db, orderID).Try()
	user := fetchUser(db, order.UserID).Try()
	validateOrderAmount(order).Try()
	payment := chargePayment(user, order).Try()
	receipt := generateReceipt(payment).Try()

	return result.Ok(receipt)
}

// ResultStyleWithAndThen shows the functional chaining pattern.
func ResultStyleWithAndThen(db *sql.DB, orderID int) result.Result[string] {
	orderResult := fetchOrder(db, orderID)
	orderResult = result.AndThen(orderResult, func(order Order) result.Result[Order] {
		return validateOrderAmount(order)
	})
	userResult := result.AndThen(orderResult, func(order Order) result.Result[User] {
		return fetchUser(db, order.UserID)
	})
	receipt := result.AndThen(userResult, func(user User) result.Result[string] {
		// Need order here but it's not available - limitation of this pattern
		return result.Ok("receipt")
	})
	return receipt
}

func fetchOrderTraditional(db *sql.DB, orderID int) (Order, error) {
	return Order{}, nil
}

func fetchUserTraditional(db *sql.DB, userID int) (User, error) {
	return User{}, nil
}

func chargePaymentTraditional(user User, order Order) (string, error) {
	return "payment-id", nil
}

func generateReceiptTraditional(paymentID string) (string, error) {
	return "receipt", nil
}
