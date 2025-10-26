// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package result_test. result_benchmark_with_db_test provides benchmarks for result.Result types.
package result_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/seyedali-dev/goxide/internal/tests"
	"github.com/seyedali-dev/goxide/rusty/chain"
	"github.com/seyedali-dev/goxide/rusty/result"
)

// Test suite setup
var (
	testDB          *sql.DB
	traditionalRepo *TraditionalUserRepo
	resultRepo      *ResultUserRepo
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	tc, err := tests.SetupTestContainer(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to setup test container: %v\n", err)
		os.Exit(1)
	}
	defer tc.Cleanup(ctx)

	testDB = tc.DB

	setupDatabase(ctx)

	exitCode := m.Run()
	os.Exit(exitCode)
}

func setupDatabase(ctx context.Context) {
	// Create users table
	_, err := testDB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to create users table: %v", err))
	}

	// Clear any existing data
	_, err = testDB.ExecContext(ctx, "TRUNCATE TABLE users RESTART IDENTITY")
	if err != nil {
		panic(fmt.Sprintf("failed to truncate users table: %v", err))
	}

	traditionalRepo = NewTraditionalUserRepo(testDB)
	resultRepo = NewResultUserRepo(testDB)
}

func clearUsersTable(ctx context.Context) {
	_, err := testDB.ExecContext(ctx, "TRUNCATE TABLE users RESTART IDENTITY")
	if err != nil {
		panic(fmt.Errorf("failed to truncate users table: %w", err))
	}
}

// Database Benchmarks

// Test results:
//
//	BenchmarkTraditionalDBCreateUser    	     642	   1632707 ns/op	    1133 B/op	      27 allocs/op
//	BenchmarkTraditionalDBCreateUser    	     792	   1452404 ns/op	    1133 B/op	      27 allocs/op
//	BenchmarkTraditionalDBCreateUser    	     799	   1443119 ns/op	    1133 B/op	      27 allocs/op
//	BenchmarkTraditionalDBCreateUser    	     837	   1424413 ns/op	    1133 B/op	      27 allocs/op
//	BenchmarkTraditionalDBCreateUser    	     847	   1412738 ns/op	    1133 B/op	      27 allocs/op
//	BenchmarkTraditionalDBCreateUser    	     850	   1430383 ns/op	    1133 B/op	      27 allocs/op
func BenchmarkTraditionalDBCreateUser(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		email := fmt.Sprintf("user%d@example.com", i)
		id, err := traditionalRepo.CreateUser(ctx, email, "Test User")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		if id <= 0 {
			b.Fatal("expected positive ID")
		}
	}
}

// Test results:
//
//	BenchmarkResultDBCreateUser    	     795	   1576020 ns/op	    1138 B/op	      28 allocs/op
//	BenchmarkResultDBCreateUser    	     800	   1437796 ns/op	    1138 B/op	      28 allocs/op
//	BenchmarkResultDBCreateUser    	     800	   1423525 ns/op	    1138 B/op	      28 allocs/op
//	BenchmarkResultDBCreateUser    	     831	   1627161 ns/op	    1138 B/op	      28 allocs/op
//	BenchmarkResultDBCreateUser    	     836	   1419551 ns/op	    1138 B/op	      28 allocs/op
//	BenchmarkResultDBCreateUser    	     846	   1419076 ns/op	    1138 B/op	      28 allocs/op
func BenchmarkResultDBCreateUser(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		email := fmt.Sprintf("user%d@example.com", i)
		res := resultRepo.CreateUser(ctx, email, "Test User")
		if res.IsErr() {
			b.Fatalf("unexpected error: %v", res.Err())
		}
		id := res.Unwrap()
		if id <= 0 {
			b.Fatal("expected positive ID")
		}
	}
}

// Test results:
//
//	BenchmarkTraditionalDBFindUser    	    8821	    125367 ns/op	    1104 B/op	      27 allocs/op
//	BenchmarkTraditionalDBFindUser    	    9076	    119524 ns/op	    1104 B/op	      27 allocs/op
//	BenchmarkTraditionalDBFindUser    	    9727	    118185 ns/op	    1104 B/op	      27 allocs/op
//	BenchmarkTraditionalDBFindUser    	   10000	    132449 ns/op	    1104 B/op	      27 allocs/op
//	BenchmarkTraditionalDBFindUser    	   10000	    112932 ns/op	    1104 B/op	      27 allocs/op
//	BenchmarkTraditionalDBFindUser    	   10000	    114335 ns/op	    1104 B/op	      27 allocs/op
func BenchmarkTraditionalDBFindUser(b *testing.B) {
	ctx := context.Background()

	// Setup: create a user first
	// Clear table before each iteration to avoid unique constraint violations
	clearUsersTable(ctx)
	id, err := traditionalRepo.CreateUser(ctx, "finduser@example.com", "Find User")
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {

		user, err := traditionalRepo.FindUserByID(ctx, id)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		if user.Email != "finduser@example.com" {
			b.Fatal("unexpected user data")
		}
	}
}

// Test results:
//
//	BenchmarkResultDBFindUser    	    9747	    112171 ns/op	    1112 B/op	      28 allocs/op
//	BenchmarkResultDBFindUser    	    9852	    121999 ns/op	    1112 B/op	      28 allocs/op
//	BenchmarkResultDBFindUser    	   10000	    118915 ns/op	    1112 B/op	      28 allocs/op
//	BenchmarkResultDBFindUser    	   10000	    120670 ns/op	    1112 B/op	      28 allocs/op
//	BenchmarkResultDBFindUser    	   10000	    125825 ns/op	    1112 B/op	      28 allocs/op
//	BenchmarkResultDBFindUser    	   10000	    124088 ns/op	    1112 B/op	      28 allocs/op
func BenchmarkResultDBFindUser(b *testing.B) {
	ctx := context.Background()

	// Setup: create a user first
	// Clear table before each iteration to avoid unique constraint violations
	clearUsersTable(ctx)
	res := resultRepo.CreateUser(ctx, "finduser@example.com", "Find User")
	if res.IsErr() {
		b.Fatalf("setup failed: %v", res.Err())
	}
	id := res.Unwrap()

	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {

		userRes := resultRepo.FindUserByID(ctx, id)
		if userRes.IsErr() {
			b.Fatalf("unexpected error: %v", userRes.Err())
		}
		user := userRes.Unwrap()
		if user.Email != "finduser@example.com" {
			b.Fatal("unexpected user data")
		}
	}
}

// Test results:
//
//	BenchmarkTraditionalDBFindUserNotFound    	     814	   1416564 ns/op	    1122 B/op	      27 allocs/op
//	BenchmarkTraditionalDBFindUserNotFound    	     826	   1469812 ns/op	    1122 B/op	      27 allocs/op
//	BenchmarkTraditionalDBFindUserNotFound    	     846	   1497534 ns/op	    1122 B/op	      27 allocs/op
//	BenchmarkTraditionalDBFindUserNotFound    	     873	   1496532 ns/op	    1122 B/op	      27 allocs/op
//	BenchmarkTraditionalDBFindUserNotFound    	     880	   1364513 ns/op	    1122 B/op	      27 allocs/op
//	BenchmarkTraditionalDBFindUserNotFound    	     890	   1434094 ns/op	    1122 B/op	      27 allocs/op
func BenchmarkTraditionalDBFindUserNotFound(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		user, err := traditionalRepo.FindUserByID(ctx, 999999)
		if err == nil {
			b.Fatal("expected error for non-existent user")
		}
		if user != nil {
			b.Fatal("expected nil user")
		}
	}
}

// Test results:
//
//	BenchmarkResultDBFindUserNotFound    	     502	   2058408 ns/op	    1124 B/op	      27 allocs/op
//	BenchmarkResultDBFindUserNotFound    	     680	   1612772 ns/op	    1123 B/op	      27 allocs/op
//	BenchmarkResultDBFindUserNotFound    	     836	   1794866 ns/op	    1123 B/op	      27 allocs/op
//	BenchmarkResultDBFindUserNotFound    	     853	   1392572 ns/op	    1123 B/op	      27 allocs/op
//	BenchmarkResultDBFindUserNotFound    	     855	   1659572 ns/op	    1122 B/op	      27 allocs/op
//	BenchmarkResultDBFindUserNotFound    	     879	   1465590 ns/op	    1122 B/op	      27 allocs/op
func BenchmarkResultDBFindUserNotFound(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		userRes := resultRepo.FindUserByID(ctx, 999999)
		if userRes.IsOk() {
			b.Fatal("expected error for non-existent user")
		}
		if userRes.UnwrapOr(nil) != nil {
			b.Fatal("expected nil user")
		}
	}
}

// Test results:
//
//	BenchmarkTraditionalDBUpdateUser    	    8726	    130652 ns/op	     368 B/op	      10 allocs/op
//	BenchmarkTraditionalDBUpdateUser    	    9182	    152693 ns/op	     368 B/op	      10 allocs/op
//	BenchmarkTraditionalDBUpdateUser    	    9254	    143603 ns/op	     368 B/op	      10 allocs/op
//	BenchmarkTraditionalDBUpdateUser    	    9532	    127135 ns/op	     368 B/op	      10 allocs/op
//	BenchmarkTraditionalDBUpdateUser    	    9856	    132048 ns/op	     368 B/op	      10 allocs/op
//	BenchmarkTraditionalDBUpdateUser    	   10000	    149509 ns/op	     368 B/op	      10 allocs/op
func BenchmarkTraditionalDBUpdateUser(b *testing.B) {
	ctx := context.Background()

	// Setup: create a user first
	// Clear table before each iteration to avoid unique constraint violations
	clearUsersTable(ctx)
	id, err := traditionalRepo.CreateUser(ctx, "updateuser@example.com", "Old Name")
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {

		newName := fmt.Sprintf("New Name %d", i)
		err := traditionalRepo.UpdateUserName(ctx, id, newName)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// Test results:
//
//	BenchmarkResultDBUpdateUser    	    8686	    125186 ns/op	     376 B/op	      11 allocs/op
//	BenchmarkResultDBUpdateUser    	    8998	    145335 ns/op	     376 B/op	      11 allocs/op
//	BenchmarkResultDBUpdateUser    	    9456	    127883 ns/op	     376 B/op	      11 allocs/op
//	BenchmarkResultDBUpdateUser    	    9338	    138504 ns/op	     376 B/op	      11 allocs/op
//	BenchmarkResultDBUpdateUser    	    9675	    141047 ns/op	     376 B/op	      11 allocs/op
//	BenchmarkResultDBUpdateUser    	    9250	    134451 ns/op	     376 B/op	      11 allocs/op
func BenchmarkResultDBUpdateUser(b *testing.B) {
	ctx := context.Background()

	// Setup: create a user first
	// Clear table before each iteration to avoid unique constraint violations
	clearUsersTable(ctx)
	res := resultRepo.CreateUser(ctx, "updateuser@example.com", "Old Name")
	if res.IsErr() {
		b.Fatalf("setup failed: %v", res.Err())
	}
	id := res.Unwrap()

	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {

		newName := fmt.Sprintf("New Name %d", i)
		updateRes := resultRepo.UpdateUserName(ctx, id, newName)
		if updateRes.IsErr() {
			b.Fatalf("unexpected error: %v", updateRes.Err())
		}
		if !updateRes.Unwrap() {
			b.Fatal("expected update to affect rows")
		}
	}
}

// Test results:
//
//	BenchmarkTraditionalDBGetOrCreateUser    	     528	   2021604 ns/op	    3283 B/op	      76 allocs/op
//	BenchmarkTraditionalDBGetOrCreateUser    	     550	   2351294 ns/op	    3284 B/op	      76 allocs/op
//	BenchmarkTraditionalDBGetOrCreateUser    	     614	   2053341 ns/op	    3284 B/op	      76 allocs/op
//	BenchmarkTraditionalDBGetOrCreateUser    	     619	   2169874 ns/op	    3284 B/op	      76 allocs/op
//	BenchmarkTraditionalDBGetOrCreateUser    	     622	   1970512 ns/op	    3284 B/op	      76 allocs/op
//	BenchmarkTraditionalDBGetOrCreateUser    	     625	   2142520 ns/op	    3284 B/op	      76 allocs/op
func BenchmarkTraditionalDBGetOrCreateUser(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		email := fmt.Sprintf("getorcreate%d@example.com", i)
		user, err := traditionalRepo.GetOrCreateUser(ctx, email, "Test User")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		if user.Email != email {
			b.Fatal("unexpected user email")
		}
	}
}

// Test resuls
//
//	BenchmarkResultDBGetOrCreateUser    	     592	   1799120 ns/op	    3372 B/op	      80 allocs/op
//	BenchmarkResultDBGetOrCreateUser    	     627	   1818443 ns/op	    3371 B/op	      80 allocs/op
//	BenchmarkResultDBGetOrCreateUser    	     651	   1705625 ns/op	    3372 B/op	      80 allocs/op
//	BenchmarkResultDBGetOrCreateUser    	     658	   1813511 ns/op	    3372 B/op	      80 allocs/op
//	BenchmarkResultDBGetOrCreateUser    	     679	   1720880 ns/op	    3372 B/op	      80 allocs/op
//	BenchmarkResultDBGetOrCreateUser    	     685	   1724771 ns/op	    3372 B/op	      80 allocs/op
func BenchmarkResultDBGetOrCreateUser(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		email := fmt.Sprintf("getorcreate%d@example.com", i)
		userRes := resultRepo.GetOrCreateUser(ctx, email, "Test User")
		if userRes.IsErr() {
			b.Fatalf("unexpected error: %v", userRes.Err())
		}
		user := userRes.Unwrap()
		if user.Email != email {
			b.Fatal("unexpected user email")
		}
	}
}

// Complex operation: Multiple chained database operations

// Test results:
//
//	BenchmarkTraditionalDBChainedOperations           614           1926726 ns/op            3699 B/op         90 allocs/op
//	BenchmarkTraditionalDBChainedOperations           636           1913628 ns/op            3699 B/op         90 allocs/op
//	BenchmarkTraditionalDBChainedOperations           637           1809579 ns/op            3698 B/op         90 allocs/op
//	BenchmarkTraditionalDBChainedOperations           650           1793920 ns/op            3698 B/op         90 allocs/op
//	BenchmarkTraditionalDBChainedOperations           654           1840567 ns/op            3699 B/op         90 allocs/op
//	BenchmarkTraditionalDBChainedOperations           662           1886138 ns/op            3698 B/op         90 allocs/op
func BenchmarkTraditionalDBChainedOperations(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		// Create user
		email := fmt.Sprintf("chained%d@example.com", i)
		id, err := traditionalRepo.CreateUser(ctx, email, "Chained User")
		if err != nil {
			b.Fatalf("create failed: %v", err)
		}

		// Find user
		user, err := traditionalRepo.FindUserByID(ctx, id)
		if err != nil {
			b.Fatalf("find failed: %v", err)
		}

		// Update user
		err = traditionalRepo.UpdateUserName(ctx, user.ID, "Updated Name")
		if err != nil {
			b.Fatalf("update failed: %v", err)
		}

		// Find again to verify
		updatedUser, err := traditionalRepo.FindUserByID(ctx, user.ID)
		if err != nil {
			b.Fatalf("second find failed: %v", err)
		}
		if updatedUser.Name != "Updated Name" {
			b.Fatal("update didn't persist")
		}
	}
}

// Test results:
//
//	BenchmarkResultDBChainedOperations                        596           1796836 ns/op            2603 B/op         66 allocs/op
//	BenchmarkResultDBChainedOperations                        660           1701199 ns/op            2603 B/op         66 allocs/op
//	BenchmarkResultDBChainedOperations                        696           1827799 ns/op            2602 B/op         66 allocs/op
//	BenchmarkResultDBChainedOperations                        706           1706408 ns/op            2602 B/op         66 allocs/op
//	BenchmarkResultDBChainedOperations                        710           1699925 ns/op            2602 B/op         66 allocs/op
//	BenchmarkResultDBChainedOperations                        712           1730262 ns/op            2603 B/op         66 allocs/op
func BenchmarkResultDBChainedOperations(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		// Using AndThen for chained operations
		finalResult := chain.Chain2[bool, *User, int](resultRepo.CreateUser(ctx, fmt.Sprintf("chained%d@example.com", i), "Chained User")).
			AndThen(func(id int) result.Result[*User] {
				return resultRepo.FindUserByID(ctx, id)
			}).
			AndThen(func(user *User) result.Result[bool] {
				return resultRepo.UpdateUserName(ctx, user.ID, "Updated Name")
			})

		if finalResult.IsErr() {
			b.Fatalf("chained operations failed: %v", finalResult.Err())
		}
	}
}

// Test results:
//
//	BenchmarkResultDBChainedOperationsBubbleUp                530           1977630 ns/op            3724 B/op         95 allocs/op
//	BenchmarkResultDBChainedOperationsBubbleUp                640           1925477 ns/op            3722 B/op         95 allocs/op
//	BenchmarkResultDBChainedOperationsBubbleUp                646           1820982 ns/op            3722 B/op         95 allocs/op
//	BenchmarkResultDBChainedOperationsBubbleUp                651           1841983 ns/op            3723 B/op         95 allocs/op
//	BenchmarkResultDBChainedOperationsBubbleUp                655           1836290 ns/op            3723 B/op         95 allocs/op
//	BenchmarkResultDBChainedOperationsBubbleUp                667           1888982 ns/op            3722 B/op         95 allocs/op
func BenchmarkResultDBChainedOperationsBubbleUp(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		var finalResult result.Result[*User]

		func() {
			defer result.Catch(&finalResult)

			// Create user
			id := resultRepo.CreateUser(ctx, fmt.Sprintf("bubbleup%d@example.com", i), "BubbleUp User").BubbleUp()

			// Find user
			user := resultRepo.FindUserByID(ctx, id).BubbleUp()

			// Update user
			updated := resultRepo.UpdateUserName(ctx, user.ID, "Updated Name").BubbleUp()
			if !updated {
				panic(errors.New("update failed"))
			}

			// Find again to verify
			finalUser := resultRepo.FindUserByID(ctx, user.ID).BubbleUp()
			finalResult = result.Ok(finalUser)
		}()

		if finalResult.IsErr() {
			b.Fatalf("bubbleup operations failed: %v", finalResult.Err())
		}
	}
}

// Benchmark error handling with fallbacks

// Test results:
//
//	BenchmarkTraditionalDBErrorHandlingWithFallback           637           1748260 ns/op            3216 B/op         75 allocs/op
//	BenchmarkTraditionalDBErrorHandlingWithFallback           646           1763479 ns/op            3216 B/op         75 allocs/op
//	BenchmarkTraditionalDBErrorHandlingWithFallback           661           1761188 ns/op            3215 B/op         75 allocs/op
//	BenchmarkTraditionalDBErrorHandlingWithFallback           693           1674359 ns/op            3216 B/op         75 allocs/op
//	BenchmarkTraditionalDBErrorHandlingWithFallback           693           1727367 ns/op            3216 B/op         75 allocs/op
//	BenchmarkTraditionalDBErrorHandlingWithFallback           721           1788172 ns/op            3216 B/op         75 allocs/op
func BenchmarkTraditionalDBErrorHandlingWithFallback(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		var user *User

		// Try to find existing user
		existingUser, err := traditionalRepo.FindUserByEmail(ctx, "nonexistent@example.com")
		if err != nil {
			// Fallback: create new user
			id, createErr := traditionalRepo.CreateUser(ctx, "fallback@example.com", "Fallback User")
			if createErr != nil {
				b.Fatalf("fallback failed: %v", createErr)
			}

			newUser, findErr := traditionalRepo.FindUserByID(ctx, id)
			if findErr != nil {
				b.Fatalf("find after create failed: %v", findErr)
			}
			user = newUser
		} else {
			user = existingUser
		}

		if user == nil {
			b.Fatal("user should not be nil")
		}
	}
}

// Test results:
//
//	BenchmarkResultDBErrorHandlingWithFallback                631           1723818 ns/op            3307 B/op         79 allocs/op
//	BenchmarkResultDBErrorHandlingWithFallback                667           1692890 ns/op            3306 B/op         79 allocs/op
//	BenchmarkResultDBErrorHandlingWithFallback                678           1923316 ns/op            3307 B/op         79 allocs/op
//	BenchmarkResultDBErrorHandlingWithFallback                681           1729803 ns/op            3306 B/op         79 allocs/op
//	BenchmarkResultDBErrorHandlingWithFallback                684           1785772 ns/op            3307 B/op         79 allocs/op
//	BenchmarkResultDBErrorHandlingWithFallback                708           1734439 ns/op            3306 B/op         79 allocs/op
func BenchmarkResultDBErrorHandlingWithFallback(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		userRes := resultRepo.FindUserByEmail(ctx, "nonexistent@example.com").
			UnwrapOrElse(func(err error) *User {
				// Fallback: create new user
				id := resultRepo.CreateUser(ctx, "fallback@example.com", "Fallback User").Unwrap()
				return resultRepo.FindUserByID(ctx, id).Unwrap()
			})

		if userRes == nil {
			b.Fatal("user should not be nil")
		}
	}
}

// Memory allocation benchmarks

// Test results:
//
// BenchmarkTraditionalDBCreateUserAllocs            798           1521647 ns/op            1133 B/op         27 allocs/op
// BenchmarkTraditionalDBCreateUserAllocs            801           1528827 ns/op            1133 B/op         27 allocs/op
// BenchmarkTraditionalDBCreateUserAllocs            819           1452859 ns/op            1133 B/op         27 allocs/op
// BenchmarkTraditionalDBCreateUserAllocs            836           1437072 ns/op            1133 B/op         27 allocs/op
// BenchmarkTraditionalDBCreateUserAllocs            836           1454589 ns/op            1133 B/op         27 allocs/op
// BenchmarkTraditionalDBCreateUserAllocs            844           1425961 ns/op            1133 B/op         27 allocs/op
func BenchmarkTraditionalDBCreateUserAllocs(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		email := fmt.Sprintf("alloc%d@example.com", i)
		id, err := traditionalRepo.CreateUser(ctx, email, "Test User")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		if id <= 0 {
			b.Fatal("expected positive ID")
		}
	}
}

// Test results:
//
//	BenchmarkResultDBCreateUserAllocs                 694           1551579 ns/op            1138 B/op         28 allocs/op
//	BenchmarkResultDBCreateUserAllocs                 733           1535593 ns/op            1139 B/op         28 allocs/op
//	BenchmarkResultDBCreateUserAllocs                 766           1558525 ns/op            1138 B/op         28 allocs/op
//	BenchmarkResultDBCreateUserAllocs                 810           1569651 ns/op            1138 B/op         28 allocs/op
//	BenchmarkResultDBCreateUserAllocs                 824           1509752 ns/op            1138 B/op         28 allocs/op
//	BenchmarkResultDBCreateUserAllocs                 844           1598817 ns/op            1138 B/op         28 allocs/op
func BenchmarkResultDBCreateUserAllocs(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Clear table before each iteration to avoid unique constraint violations
		clearUsersTable(ctx)

		email := fmt.Sprintf("alloc%d@example.com", i)
		res := resultRepo.CreateUser(ctx, email, "Test User")
		if res.IsErr() {
			b.Fatalf("unexpected error: %v", res.Err())
		}
		id := res.Unwrap()
		if id <= 0 {
			b.Fatal("expected positive ID")
		}
	}
}
