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
type ErrIndexNotation struct{}

func (e ErrIndexNotation) Error() string {
	return "Cannot append to array using index notation"
}

// ErrNoData is returned when a RawExtension object does not contain any data.
// This occurs when both the Raw and Object fields are nil during conversion from
// a runtime.RawExtension to a NamedObject.
type ErrNoData struct{}

func (e ErrNoData) Error() string {
	return "No data found in raw object"
}

// ErrMissingName is returned when a Kubernetes object does not have a name or
// generateName field set in its metadata. This occurs during object validation
// in NamedObjectFromUnstructured when neither metadata.name nor metadata.generateName
// is present. All Kubernetes objects must have at least one of these fields defined.
type ErrMissingName struct{}

func (e ErrMissingName) Error() string {
	return "Object does not have a name set"
}

// ErrUnsupportedHashType is returned when attempting to hash a field with a type
// that is not supported by the hashing algorithm. This occurs when:
//   - A field type cannot be converted to a hashable representation
//   - An unknown or complex type is encountered during object hashing
//
// The error string contains the field name and its type information.
type ErrUnsupportedHashType string

func (e ErrUnsupportedHashType) Error() string {
	return string(e)
}

// ErrUnknownOperation is returned when an admission webhook receives a request
// with an operation type that is not recognized. Valid operations are Create,
// Update, and Delete. This error occurs in AdmissionRequestHook.Call when the
// admission request contains an unsupported operation.
//
// The error string contains the unknown operation name.
type ErrUnknownOperation string

func (e ErrUnknownOperation) Error() string {
	return fmt.Sprintf("Unknown admission operation: %s", string(e))
}

// ErrNoCallback is returned when an admission webhook receives a request for an
// operation that does not have a validation callback registered. This occurs in
// AdmissionRequestHook.Call when the operation (Create, Update, or Delete) handler
// is nil. The request is still marked as validated to avoid blocking operations.
//
// The error string contains the operation name that lacks a callback.
type ErrNoCallback string

func (e ErrNoCallback) Error() string {
	return fmt.Sprintf("Operation %s has no callback set", string(e))
}

// ErrInvalidBoundObjectRef is returned when attempting to create a service account
// token with an invalid bound object reference. This occurs when:
//   - A bound object reference is provided but is not a Pod
//   - The Kind field is set to something other than "pod" (case-insensitive)
//
// Service account tokens can only be bound to Pod objects or have no binding.
type ErrInvalidBoundObjectRef struct{}

func (e ErrInvalidBoundObjectRef) Error() string {
	return "Bound object reference must be a pod or nil"
}

// ErrNoToken is returned when a service account token request succeeds but the
// response does not contain a token. This occurs in GetServiceAccountToken when
// the Kubernetes API returns a successful response with an empty token field,
// indicating an unexpected API behavior.
type ErrNoToken struct{}

func (e ErrNoToken) Error() string {
	return "No token in server response"
}

// ErrParseError is returned when parsing label selector components fails due to
// type mismatches. This occurs in ParseLabelSelector when:
//   - A selector value is not a string
//   - matchLabels is not a map[string]string or map[string]interface{}
//   - matchExpressions is not the expected slice type
//   - A matchExpressions element is not a map[string]interface{}
//   - Required fields (key, operator, values) are not of the expected type
//
// The error string contains details about what failed to parse and the actual value.
type ErrParseError string

func (e ErrParseError) Error() string {
	return string(e)
}
