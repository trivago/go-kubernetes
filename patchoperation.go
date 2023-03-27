// Copied from
// https://github.com/douglasmakey/admissionkubernetes

package kubernetes

import (
	"strings"
)

// PatchOperation is an operation of a JSON patch https://tools.ietf.org/html/rfc6902.
// This is required to report changes back through an admissionreview response.
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	From  string      `json:"from,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// NewPatchOperationAdd returns an "add" JSON patch operation.
func NewPatchOperationAdd(path string, value interface{}) PatchOperation {
	return PatchOperation{
		Op:    "add",
		Path:  path,
		Value: value,
	}
}

// NewPatchOperationRemove returns a "remove" JSON patch operation.
func NewPatchOperationRemove(path string) PatchOperation {
	return PatchOperation{
		Op:   "remove",
		Path: path,
	}
}

// NewPatchOperationReplace returns a "replace" JSON patch operation.
func NewPatchOperationReplace(path string, value interface{}) PatchOperation {
	return PatchOperation{
		Op:    "replace",
		Path:  path,
		Value: value,
	}
}

// NewPatchOperationCopy returns a "copy" JSON patch operation.
func NewPatchOperationCopy(from, path string) PatchOperation {
	return PatchOperation{
		Op:   "copy",
		Path: path,
		From: from,
	}
}

// NewPatchOperationMove returns a "move" JSON patch operation.
func NewPatchOperationMove(from, path string) PatchOperation {
	return PatchOperation{
		Op:   "move",
		Path: path,
		From: from,
	}
}

// EscapeJSONPath converts an array of strings (path elments) to a valid
// JSONPatch path, escaping special characters if needed.
// See https://jsonpatch.com/#json-pointer
func EscapeJSONPath(path []string) string {
	capacity := 0
	escape := make([]bool, len(path))

	// Calculate build capacity and detect the need for escaping
	for i, part := range path {
		switch {
		case strings.ContainsRune(part, '/'):
			escape[i] = true
			capacity += len(part) + 2

		case part[len(part)-1] == ']':
			startIdx := strings.LastIndexByte(part, '[')
			numberLen := (len(part) - 1) - (startIdx + 1)
			if numberLen < 1 {
				numberLen = 1 // Empty arrays use "-"
			}
			capacity += numberLen + 2 // two "/" required

		default:
			capacity += len(part) + 1
		}
	}

	var b strings.Builder
	b.Grow(capacity)

	// Generate the escaped path in a memory friendly way
	for i, part := range path {
		b.WriteRune('/')
		switch {
		case escape[i]:
			b.WriteString(strings.ReplaceAll(part, "/", "~1"))

		case part[len(part)-1] == ']':
			startIdx := strings.LastIndexByte(part, '[')
			b.WriteString(part[:startIdx])
			b.WriteRune('/')
			if (len(part)-1)-(startIdx+1) == 0 {
				b.WriteRune('-')
			} else {
				b.WriteString(part[startIdx+1 : len(part)-1])
			}

		default:
			b.WriteString(part)
		}
	}

	return b.String()
}
