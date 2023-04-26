// Copied from
// https://github.com/douglasmakey/admissionkubernetes

package kubernetes

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
