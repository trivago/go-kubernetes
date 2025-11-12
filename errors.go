package kubernetes

import "fmt"

// ErrNotFound is returned when a requested path key or array index does not exist
// in a NamedObject. This error is used during path traversal operations when:
//   - A map key is not present in the object
//   - An array index is out of bounds
//   - A traversal operation ("-") finds no matching elements
//   - A MatchFunc returns false, indicating no match was found
//
// The error string contains the key or index that was not found.
type ErrNotFound string

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("Not found: %s", string(e))
}

// ErrNotTraversable is returned when attempting to traverse through a path element
// that cannot be navigated. This occurs when:
//   - A nil value is encountered in the traversal path
//   - A node is expected to be a map but is not of type map[string]interface{}
//   - A node is expected to be a slice but is not of type []interface{}
//   - A parent node during mutation is not the expected map or slice type
//   - A node is of an unsupported type for traversal (e.g., primitives, structs)
//
// The error string describes what was encountered and why it cannot be traversed.
type ErrNotTraversable string

func (e ErrNotTraversable) Error() string {
	return fmt.Sprintf("Not a traversable type: %s", string(e))
}

// ErrMissingArrayTraversal is returned when a path reaches an array but the next
// path segment does not include proper array notation. This occurs when:
//   - An array is encountered but the next path element is neither an index (e.g., "0", "1")
//     nor a traversal indicator ("-")
//   - The path syntax is invalid for array access
//
// Valid array notations are numeric indices or "-" for traversal. The error string
// contains the problematic path key.
type ErrMissingArrayTraversal string

func (e ErrMissingArrayTraversal) Error() string {
	return fmt.Sprintf("Array traversal indicator missing: %s", string(e))
}

// ErrNotAnArray is returned when array notation (index or traversal) is used on a
// path element that is not an array. This occurs when:
//   - A map/object is encountered but the path uses array syntax (e.g., "key/-" or "key/0")
//   - Array notation is applied to a non-slice type
//
// The error string contains the path key where the invalid array notation was used.
type ErrNotAnArray string

func (e ErrNotAnArray) Error() string {
	return fmt.Sprintf("Not an array: %s", string(e))
}

// ErrNotKeyValue is returned when a path operation expects a key-value structure
// (map[string]interface{}) but encounters a different type. This error type is defined
// but currently not used in the codebase. It may be reserved for future validation
// of path items that must be key-value objects.
//
// The error string would contain the problematic path item identifier.
type ErrNotKeyValue string

func (e ErrNotKeyValue) Error() string {
	return fmt.Sprintf("Path item is not a key/value object: %s", string(e))
}

// ErrIncorrectType is returned when a value retrieved from a path does not match
// the expected type for the operation. This occurs when:
//   - GetString is called but the value is not a string
//   - GetSection is called but the value is not a map[string]interface{}
//   - GetList is called but the value is not a []interface{}
//
// The error string contains the actual type that was encountered (e.g., "int", "bool").
type ErrIncorrectType string

func (e ErrIncorrectType) Error() string {
	return fmt.Sprintf("Incorrect type: %s", string(e))
}

// ErrIndexNotation is returned when attempting to use explicit array index notation
// during path extension operations that require dynamic array growth. This occurs when:
//   - Trying to create or add elements to an array using index notation (e.g., "0", "1")
//     instead of the append notation ("-")
//   - The operation would require inserting at a specific index during array creation
//
// Array modification operations must use "-" for appending; explicit indices are not
// supported during path extension.
type ErrIndexNotation string

func (e ErrIndexNotation) Error() string {
	return "Cannot append to array using index notation"
}
