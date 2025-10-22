// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package reflect. reflect_struct provides utility functions for working with reflection and struct tags.
//
// There are two main usage patterns:
//
//  1. Traditional: Call the public functions directly and provide the type for each call.
//  2. Type-safe: Create a reflector for a base type and call reflection functions on it.
//     The underlying calls will preserve the type and provide compile-time safety.
//
// Example 1: Traditional usage
//
//	func TraditionalUsage(user *object.User) {
//	    // Works as before
//	    tagValue := util.FieldTagValue(user, "Name", "step_category")
//	    hasTags := util.FieldHasTags(user, "Name", []string{"step", "step_category"})
//	}
//
// Example 2: Type-safe usage (optional)
//
//	func TypeSafeUsage(user *object.User) {
//	    // Create type-safe reflector once
//	    userReflector := util.ForType[object.User]()
//
//	    // Now get compile-time safety
//	    tagValue := userReflector.FieldTagValue("Name", "step_category")
//	    hasTags := userReflector.FieldHasTags("Name", []string{"step", "step_category"})
//
//	    // This would cause IDE/compiler errors if fields don't exist:
//	    // userReflector.FieldTagValue("NonExistentField", "tag") // Compile error
//	}
//
// Example 3: Mixed usage
//
//	func MixedUsage(user *object.User, authCtx *pipeline.AuthContext) {
//	    // Use type-safe for critical paths
//	    userReflector := util.ForType[object.User]()
//	    userTag := userReflector.FieldTagValue("CompletedStepProfile", "step_category")
//
//	    // Use traditional for one-off or dynamic cases
//	    authTag := util.FieldTagValue(authCtx, "SomeField", "some_tag")
//	}
package reflect

import (
	"fmt"
	"reflect"
	"strings"
)

// StructTagInfo contains information extracted from struct tags.
type StructTagInfo struct {
	FieldName string
	TagValues []string
}

// ------------------------------------- Public functions -------------------------------------

// FieldTagValue returns the value of a specific struct tag.
//
// Example:
//
//	type User struct {
//	    Name string `step:"profile" step_category:"employee"`
//	}
//
//	value := FieldTagValue(User{}, "Name", "step_category") // -> "employee"
func FieldTagValue(structType interface{}, fieldName string, tag string, tagValSeparator string) string {
	field, ok := getStructField(structType, fieldName)
	if !ok {
		return ""
	}
	return parseTagValue(field, tag, tagValSeparator)
}

// FieldHasTags checks if the given field in a struct type has all specified tags.
//
// Example:
//
//	type User struct {
//	    Name string `step:"profile" step_category:"employee"`
//	}
//	hasTags := FieldHasTags(User{}, "Name", []string{"step", "step_category"}) // -> true
func FieldHasTags(structType interface{}, fieldName string, tags []string) bool {
	field, ok := getStructField(structType, fieldName)
	if !ok {
		return false
	}
	keys := make(map[string]struct{})
	for _, key := range parseTagKeys(field) {
		keys[key] = struct{}{}
	}
	for _, tag := range tags {
		if _, ok := keys[tag]; !ok {
			return false
		}
	}
	return true
}

// FieldTagKeys returns all the tags for a specific field.
//
// Example:
//
//	type User struct {
//	    Name string `step:"profile" step_category:"employee"`
//	}
//
//	values := FieldTagKeys(User{}, "Name") // -> ["step", "step_category"]
func FieldTagKeys(structType interface{}, fieldName string) []string {
	field, ok := getStructField(structType, fieldName)
	if !ok {
		return nil
	}
	return parseTagKeys(field)
}

// FieldTagKeyValue returns the tag key and value for a specific field in a struct.
//
// Example:
//
//	type User struct {
//	    Name string `step:"profile" step_category:"employee"`
//	}
//
//	key, value, ok := FieldTagKeyValue(User{}, "Name", "step")
//	// -> "step", "profile", true
func FieldTagKeyValue(structType interface{}, fieldName string, tagKey string, tagValSeparator string) (key string, value string, found bool) {
	field, ok := getStructField(structType, fieldName)
	if !ok {
		return "", "", false
	}

	value = parseTagValue(field, tagKey, tagValSeparator)
	if value == "" {
		return "", "", false
	}

	return tagKey, value, true
}

// Field returns the field value and a boolean indicating if the field was found and is settable.
//
// Example:
//
//	type User struct {
//	    Field1 string
//	    Field2 string
//	}
//
//	value := Field(&User{}, "Field1") // -> Field1
func Field(structType interface{}, fieldName string) reflect.StructField {
	t := derefType(structType)
	if t.Kind() != reflect.Struct {
		return reflect.StructField{}
	}

	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == fieldName {
			return t.Field(i)
		}
	}

	return reflect.StructField{}
}

// FieldSet sets a field on the given object using reflection.
//
// Example:
//
//	type User struct {
//	    Field1 string
//	    Field2 string
//	}
//
//	FieldSet(&User{}, "Field1", "value")
func FieldSet(structType interface{}, fieldName string, value interface{}) error {
	field, settable := FieldValue(structType, fieldName)
	if !settable {
		return fmt.Errorf("field %s is not valid in %T", fieldName, structType)
	}

	valueReflect := reflect.ValueOf(value)
	if valueReflect.Type().AssignableTo(field.Type()) {
		field.Set(valueReflect)
		return nil
	}

	if field.Kind() == reflect.String &&
		valueReflect.Kind() == reflect.String {
		field.SetString(valueReflect.String())
		return nil
	}

	return fmt.Errorf("cannot assign value of type %T to field %s of type %s", value, fieldName, field.Type())
}

// FieldValue returns the field value and a boolean indicating if the field was found and is settable.
//
// Example:
//
//	type User struct {
//	    Name string
//	}
//
//	value, settable := FieldValue(&User{}, "Name")
func FieldValue(intrfc interface{}, fieldName string) (reflect.Value, bool) {
	v := reflect.ValueOf(intrfc)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return reflect.Value{}, false
	}
	field := v.Elem().FieldByName(fieldName)
	return field, field.IsValid() && field.CanSet()
}

// Fields returns the fields of a struct.
//
// Example:
//
//	type User struct {
//	    Field1 string
//	    Field2 int
//	}
//
//	fields := Fields(User{}) // -> [Field1, Field2]
func Fields(structType interface{}) []reflect.StructField {
	t := derefType(structType)
	if t.Kind() != reflect.Struct {
		return nil
	}
	fields := make([]reflect.StructField, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i))
	}
	return fields
}

// FieldValues returns the values of all fields in a struct.
//
// Example:
//
//	type User struct {
//	    Field1 string
//	    Field2 int
//	}
//
//	values := FieldValues(User{}) // -> [Field1, Field2]
func FieldValues(structInstance interface{}) []reflect.Value {
	v := reflect.ValueOf(structInstance)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	values := make([]reflect.Value, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values = append(values, v.Field(i))
	}
	return values
}

// FieldNameByTagValue returns the names of struct fields where the specified tag has the given expected value.
//
// Example:
//
//	type User struct {
//	    Field1 string `step_category:"employee"`
//	    Field2 string `step_category:"customer"`
//	}
//	fields := FieldNameByTagValue(User{}, "step_category", "employee") // -> "Field1"
func FieldNameByTagValue(structType interface{}, tagKey, expectedValue string) string {
	var result string
	t := derefType(structType)
	if t.Kind() != reflect.Struct {
		return result
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get(tagKey) == expectedValue {
			result = field.Name
		}
	}
	return result
}

// FieldNamesByTagValue returns the names of struct fields where the specified tag has the given expected value.
//
// Example:
//
//	type User struct {
//	    Field1 string `step_category:"employee"`
//	    Field2 string `step_category:"customer"`
//	}
//	fields := FieldNamesByTagValue(User{}, "step_category", "employee") // -> ["Field1"]
func FieldNamesByTagValue(structType interface{}, tagKey, expectedValue string) []string {
	var result []string
	t := derefType(structType)
	if t.Kind() != reflect.Struct {
		return result
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get(tagKey) == expectedValue {
			result = append(result, field.Name)
		}
	}
	return result
}

// FieldTagValues returns all values for a specific struct tag key.
// If the tag contains multiple comma-separated values, they are split into a slice.
//
// Example:
//
//	type User struct {
//	    Name string `step_restrictions:"emp_only,has_code"`
//	}
//
//	values := FieldTagValues(User{}, "Name", "step_restrictions")
//	// -> ["emp_only", "has_code"]
func FieldTagValues(structType interface{}, fieldName string, tagKey string, tagValSeparator string) []string {
	field, ok := getStructField(structType, fieldName)
	if !ok {
		return nil
	}
	return parseTagValues(field, tagKey, tagValSeparator)
}

// FieldHasTagValue checks if the given field has a specific value in a multi-value tag.
//
// Example:
//
//	type User struct {
//	    Name string `step_restrictions:"emp_only,has_code"`
//	}
//
//	hasValue := FieldHasTagValue(User{}, "Name", "step_restrictions", "has_code") // -> true
func FieldHasTagValue(structType interface{}, fieldName string, tagKey string, expectedValue string, tagValSeparator string) bool {
	values := FieldTagValues(structType, fieldName, tagKey, tagValSeparator)
	for _, v := range values {
		if v == expectedValue {
			return true
		}
	}
	return false
}

// FieldsByTagContainsValue returns the names of struct fields where the specified tag contains the given value.
//
// Example:
//
//	type User struct {
//	    Field1 string `step_restrictions:"emp_only,has_code"`
//	    Field2 string `step_restrictions:"customer"`
//	    Field3 string `step_restrictions:"has_code"`
//	}
//
//	fields := FieldsByTagContainsValue(User{}, "step_restrictions", "has_code")
//	// -> ["Field1", "Field3"]
func FieldsByTagContainsValue(structType interface{}, tagKey, expectedValue, tagValSeparator string) []reflect.StructField {
	var result []reflect.StructField
	t := derefType(structType)
	if t.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		for _, tagValue := range parseTagValues(field, tagKey, tagValSeparator) {
			if tagValue == expectedValue {
				result = append(result, field)
				break
			}
		}
	}

	return result
}

// StructTypeName returns the fully-qualified name of the struct type.
//
// Example:
//
//	type User struct {}
//	name := StructTypeName(User{}) // -> "your/package/path.User"
//
// If the struct type is defined in the main package or has no package path,
// it returns just the type name.
func StructTypeName(structType interface{}) string {
	t := derefType(structType)
	if t == nil {
		return "<unknown>"
	}
	if t.PkgPath() != "" {
		return t.PkgPath() + "." + t.Name()
	}
	return t.Name()
}

// ------------------------------------- Private Helper functions -------------------------------------

// derefType returns the underlying type if a pointer is provided.
func derefType(structType interface{}) reflect.Type {
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// getStructField returns the StructField and false if not found.
func getStructField(structType interface{}, fieldName string) (reflect.StructField, bool) {
	t := derefType(structType)
	if t.Kind() != reflect.Struct {
		return reflect.StructField{}, false
	}
	return t.FieldByName(fieldName)
}

// parseTagKeys returns all tag keys for a given StructField.
func parseTagKeys(field reflect.StructField) []string {
	tagString := string(field.Tag)
	if tagString == "" {
		return nil
	}
	parts := strings.Split(tagString, " ")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) == 2 {
			keys = append(keys, kv[0])
		}
	}
	return keys
}

// parseTagValue returns the value of a specific tag key for a given StructField.
func parseTagValue(field reflect.StructField, tag string, tagValSeparator string) string {
	// Get the tag value
	tagValue := field.Tag.Get(tag)
	if tagValue == "" {
		return ""
	}

	// If no separator specified or it's a space, return the whole value
	if tagValSeparator == "" || tagValSeparator == " " {
		return tagValue
	}

	// If separator is specified and exists in the value, return the first part
	parts := strings.Split(tagValue, tagValSeparator)
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}

	return tagValue
}

// parseTagValues returns all values (separated by tagValSeparator) for a specific tag key in a StructField.
func parseTagValues(field reflect.StructField, tagKey string, tagValSeparator string) []string {
	// Get the tag value
	tagValue := field.Tag.Get(tagKey)
	if tagValue == "" {
		return nil
	}

	// If no separator specified or it's a space, return as single value
	if tagValSeparator == "" || tagValSeparator == " " {
		return []string{tagValue}
	}

	// Split by the separator
	parts := strings.Split(tagValue, tagValSeparator)
	values := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

// ------------------------------------- Type-Safe Wrapper (Optional) -------------------------------------

// Reflector provides optional type-safe reflection operations for a specific struct type.
// Usage:
//
//	// Create type-safe reflector (optional)
//	userReflector := util.ForType[object.User]()
//	value := userReflector.FieldTagValue("Name", "step_category")
//
//	// Or use existing functions (backward compatible)
//	value := util.FieldTagValue(user, "Name", "step_category")
type Reflector[T any] struct {
	// structType is cached for performance
	structType reflect.Type
}

// ForType creates a type-safe reflector for the given struct type.
// This is optional - existing functions remain available for backward compatibility.
//
// Example:
//
//	userReflector := util.ForType[object.User]()
//	tagValue := userReflector.FieldTagValue("CompletedStepProfile", "step_category")
//	// Compile-time error if field doesn't exist in object.User
func ForType[T any]() *Reflector[T] {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return &Reflector[T]{structType: t}
}

// FieldTagValue returns the value of a specific struct tag for the given field.
func (reflector *Reflector[T]) FieldTagValue(fieldName string, tag string, tagValSeparator string) string {
	var zero T
	return FieldTagValue(zero, fieldName, tag, tagValSeparator)
}

// FieldHasTags checks if the given field has all specified tags.
func (reflector *Reflector[T]) FieldHasTags(fieldName string, tags []string) bool {
	var zero T
	return FieldHasTags(zero, fieldName, tags)
}

// FieldTagKeys returns all the tags for a specific field.
func (reflector *Reflector[T]) FieldTagKeys(fieldName string) []string {
	var zero T
	return FieldTagKeys(zero, fieldName)
}

// FieldTagKeyValue returns the tag key and value for a specific field.
func (reflector *Reflector[T]) FieldTagKeyValue(fieldName string, tagKey string, tagValSeparator string) (key string, value string, found bool) {
	var zero T
	return FieldTagKeyValue(zero, fieldName, tagKey, tagValSeparator)
}

// Field returns the field struct information.
func (reflector *Reflector[T]) Field(fieldName string) reflect.StructField {
	var zero T
	return Field(zero, fieldName)
}

// FieldValue returns the field value from a struct instance.
func (reflector *Reflector[T]) FieldValue(instance *T, fieldName string) (reflect.Value, bool) {
	return FieldValue(instance, fieldName)
}

// Fields returns all fields of the struct type.
func (reflector *Reflector[T]) Fields() []reflect.StructField {
	var zero T
	return Fields(zero)
}

// FieldValues returns the values of all fields from a struct instance.
func (reflector *Reflector[T]) FieldValues(instance *T) []reflect.Value {
	return FieldValues(instance)
}

// FieldNameByTagValue returns the name of the first field where the specified tag has the given value.
func (reflector *Reflector[T]) FieldNameByTagValue(tagKey, expectedValue string) string {
	var zero T
	return FieldNameByTagValue(zero, tagKey, expectedValue)
}

// FieldNamesByTagValue returns the names of all fields where the specified tag has the given value.
func (reflector *Reflector[T]) FieldNamesByTagValue(tagKey, expectedValue string) []string {
	var zero T
	return FieldNamesByTagValue(zero, tagKey, expectedValue)
}

// FieldTagValues returns all values for a specific struct tag key on a field.
func (reflector *Reflector[T]) FieldTagValues(fieldName string, tagKey string, tagValSeparator string) []string {
	var zero T
	return FieldTagValues(zero, fieldName, tagKey, tagValSeparator)
}

// FieldHasTagValue checks if the given field has a specific value in a multi-value tag.
func (reflector *Reflector[T]) FieldHasTagValue(fieldName string, tagKey string, expectedValue string, tagValSeparator string) bool {
	var zero T
	return FieldHasTagValue(zero, fieldName, tagKey, expectedValue, tagValSeparator)
}

// FieldsByTagContainsValue returns all fields where the specified tag contains the given value.
func (reflector *Reflector[T]) FieldsByTagContainsValue(tagKey, expectedValue, tagValSeparator string) []reflect.StructField {
	var zero T
	return FieldsByTagContainsValue(zero, tagKey, expectedValue, tagValSeparator)
}

// StructTypeName returns the fully-qualified name of the struct type.
func (reflector *Reflector[T]) StructTypeName() string {
	var zero T
	return StructTypeName(zero)
}
