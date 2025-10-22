// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package errors. error provides error utilities.
package errors

import (
	"errors"
	"fmt"

	gophreflect "github.com/seyedali-dev/gopherbox/reflect"
)

// NilError is returned when a value is considered "empty" or nil.
//
// Deprecated
var NilError = errors.New("object is nil")

// WrapNilError checks for an error or an "empty" value and returns a wrapped error if found.
//   - If err is non-nil, it wraps it with msg (if provided) using %w so error unwrapping works.
//   - If err is nil but val fails the provided condition, it returns NilError wrapped with msg (if provided).
//   - If condition is nil, it defaults to IsEmpty.
//
// Returns nil if both err is nil and val passes the condition.
//
// Usage:
//
//	userObj, err := InferType[*models.User](user)
//	if err = WrapNilError(userObj, err, "", IsEmpty); err != nil {
//		return err // <-- no need for checking if userObj == nil
//	}
//
// This avoids repetitive `if err != nil || obj == nil` checks and works with any type T.
//
// Deprecated: redundant condition calls. Use EnsureResult instead.
func WrapNilError[T any](val T, err error, msg string, condition func(T) bool) error {
	if err != nil {
		if msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}

	if condition == nil {
		condition = gophreflect.IsEmpty[T]
	}

	if condition(val) {
		if msg != "" {
			return fmt.Errorf("%w: %s", NilError, msg)
		}
		return NilError
	}

	return nil
}

// EnsureResult enforces a consistent pattern for handling (value, error) returns.
//
// Rules:
//   - If err is non-nil, return the zero value of T and the error.
//   - If err is nil but val is the zero value (including nil pointers), return
//     the zero value of T and a new error with the provided message.
//   - Otherwise, return val and nil.
//
// This centralizes the common "check error, then check nil result" pattern.
func EnsureResult[T any](val T, err error, nilErrMsg string) (T, error) {
	messageFunc := func(e error) error {
		var finErr error
		if nilErrMsg != "" {
			if e != nil {
				finErr = fmt.Errorf("%s: %w", nilErrMsg, e)
			} else {
				finErr = fmt.Errorf(nilErrMsg)
			}
			return finErr
		}
		return e
	}

	var zero T
	if err != nil {
		return zero, messageFunc(err)
	}

	if gophreflect.IsEqual(val, zero) {
		return zero, fmt.Errorf(nilErrMsg)
	}

	return val, nil
}
