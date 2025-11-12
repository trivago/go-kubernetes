package kubernetes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"maps"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/cespare/xxhash"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NamedObject represents a kubernetes object and provides common functionality
// such as patch generators or accessing common fields.
type NamedObject map[string]interface{}

// NewNamedObject create a new object with metdata.name set
func NewNamedObject(name string) NamedObject {
	return NamedObject{
		"metadata": map[string]interface{}{
			"name": name,
		},
	}
}

// NamedObjectFromUnstructured converts a raw runtime object intor a
// namespaced object. If the object does not have name or namespace set an
// error will be returned.
func NamedObjectFromRaw(data *runtime.RawExtension) (NamedObject, error) {
	if data.Raw == nil {
		if data.Object == nil {
			return NamedObject{}, ErrNoData{}
		}
		var err error
		if data.Raw, err = jsoniter.Marshal(data.Object); err != nil {
			return NamedObject{}, err
		}
	}

	parsed := unstructured.Unstructured{
		Object: make(map[string]interface{}),
	}

	// Read JSON data into a map and change object name and namespace
	if err := jsoniter.Unmarshal(data.Raw, &parsed.Object); err != nil {
		return NamedObject{}, err
	}

	return NamedObjectFromUnstructured(parsed)
}

// NamedObjectFromUnstructured converts an unstructured Kubernetes object
// into a namespaced object. If the object does not have name or namespace set
// an error will be returned.
func NamedObjectFromUnstructured(unstructuredObj unstructured.Unstructured) (NamedObject, error) {
	obj := NamedObject(unstructuredObj.Object)

	// "generateName" is used by pods before a, e.g., ReplicaSet controler
	// processed the pod.
	if !obj.Has(PathMetadataName) && !obj.Has(PathMetadataGenerateName) {
		return obj, ErrMissingName{}
	}

	return obj, nil
}

// Find looks for a path with the given value and returns all matching paths.
// If nil is passed as a value, all full matching paths will be returned.
func (obj NamedObject) FindAll(path Path, value interface{}) ([]Path, error) {
	matchedPaths := []Path{}

	matchValue := func(v interface{}, path Path) bool {
		if value == nil || reflect.DeepEqual(v, value) {
			matchedPaths = append(matchedPaths, path)
			return true
		}
		return false
	}

	_, err := obj.Walk(path, WalkArgs{
		MatchAll:  true,
		MatchFunc: matchValue,
	})

	return matchedPaths, err
}

// FindFirst looks for a path with the given value and returns the first,
// resolved, matching path. If nil is passed as a value just the path will be
// matched.
func (obj NamedObject) FindFirst(path Path, value interface{}) (Path, error) {
	matchedPath := Path{}

	matchValue := func(v interface{}, path Path) bool {
		if value == nil || reflect.DeepEqual(v, value) {
			matchedPath = path
			return true
		}
		return false
	}

	_, err := obj.Walk(path, WalkArgs{
		MatchAll:  false,
		MatchFunc: matchValue,
	})

	return matchedPath, err
}

// Get will return an object for a given path.
// If the object or any part of the path does not exist, nil is returned.
// If an unindexed array notation is used ("[]") the first matching path is
// returned.
func (obj NamedObject) Get(path Path) (interface{}, error) {
	return obj.Walk(path, WalkArgs{})
}

// Set will set a value for a given key on a given path.
// The path will be created if not existing through a call to GeneratePatch.
func (obj NamedObject) Set(path Path, value interface{}) error {
	p, v, _ := obj.GeneratePatch(path, value)
	setValue := func(interface{}) interface{} {
		return v
	}

	_, err := obj.Walk(p, WalkArgs{
		MutateFunc: setValue,
	})

	return err
}

// Delete will remove a given key on a given path.
// If an unindexed array notation is used ("[]") the first matching path will be
// used, which might lead to the key not being deleted.
// If the path is not valid because a key in the path does not exist, is no
// map or array, false will be returned. If the key is deleted or does not exist,
// true will be returned.
func (obj NamedObject) Delete(path Path) error {
	deleteKey := func(interface{}) interface{} {
		return nil
	}

	_, err := obj.Walk(path, WalkArgs{
		MutateFunc: deleteKey,
	})
	return err
}

// Has will return true if a key on a given path is set.
func (obj NamedObject) Has(path Path) bool {
	_, err := obj.Walk(path, WalkArgs{})
	return err == nil
}

// GetString will return a string value assigned to a given key on a given path.
// If the object is not a string or the path or key does not exist, an empty
// string and an error are returned.
func (obj NamedObject) GetString(path Path) (string, error) {
	value, err := obj.Get(path)
	if err != nil {
		return "", err
	}

	str, ok := value.(string)
	if !ok {
		return str, ErrIncorrectType(reflect.TypeOf(value).String())
	}
	return str, nil
}

// GetSection will return a map[string]interface{} (a sub-section) assigned to a
// given key on a given path.
// If the object is not a map or the path or key does not exist, nil and an
// error are returned.
func (obj NamedObject) GetSection(path Path) (map[string]interface{}, error) {
	value, err := obj.Get(path)
	if err != nil {
		return nil, err
	}

	obj, ok := value.(map[string]interface{})
	if !ok {
		return obj, ErrIncorrectType(reflect.TypeOf(value).String())
	}
	return obj, nil
}

// GetList will return a []interface{} assigned to a given key on a given path.
// If the object is not a list or the path or key does not exist, nil and an
// error are returned.
func (obj NamedObject) GetList(path Path) ([]interface{}, error) {
	value, err := obj.Get(path)
	if err != nil {
		return nil, err
	}

	slice, ok := value.([]interface{})
	if !ok {
		return nil, ErrIncorrectType(reflect.TypeOf(value).String())
	}
	return slice, nil
}

// GetName will return the name of the object.
// The name can be a prefix if a pod is processed before it has been processed
// by the corresponding, e.g., ReplicaSet controller.
// If the name is not set, an empty string is returned.
func (obj NamedObject) GetName() string {
	if name, err := obj.GetString(PathMetadataName); err == nil {
		return name
	}
	if namePrefix, err := obj.GetString(PathMetadataGenerateName); err == nil {
		return namePrefix
	}
	return ""
}

// GetName will return the namespace of the object.
// If the namespace is not set, an empty string is returned.
func (obj NamedObject) GetNamespace() string {
	if namespace, err := obj.GetString(PathMetadataNamespace); err == nil {
		return namespace
	}
	return ""
}

// GetKind returns the kind of the object.
func (obj NamedObject) GetKind() string {
	if kind, err := obj.GetString(Path{"kind"}); err == nil {
		return kind
	}
	return ""
}

// GetVersion returns the apiVersion of the object.
func (obj NamedObject) GetVersion() string {
	if version, err := obj.GetString(Path{"apiVersion"}); err == nil {
		return version
	}
	return ""
}

// GetUID returns the UID of the object from the metadata.
func (obj NamedObject) GetUID() string {
	if uid, err := obj.GetString(Path{"metadata", "uid"}); err == nil {
		return uid
	}
	return ""
}

// GetOwnerKind returns the resource kind of an owning resource, e.g.,
// ReplicaSet if the pod is managed by a ReplicaSet
func (obj NamedObject) GetOwnerKind() string {
	if owner, err := obj.GetString(PathOwnerReferenceKind); err == nil {
		return owner
	}
	return ""
}

// GetLabel will return the value of a given label.
// If the label is not set, an empty string and false is returned.
func (obj NamedObject) GetLabel(key string) (string, error) {
	return obj.GetString(NewPath(PathLabels, key))
}

// HasLabels returns true if a labels section exists
func (obj NamedObject) HasLabels() bool {
	return obj.Has(PathLabels)
}

// IsLabelSetTo checks if a specific label is set to a given value.
// The comparison is done in a case insensitive way.
func (obj NamedObject) IsLabelSetTo(key, value string) bool {
	label, err := obj.GetString(NewPath(PathLabels, key))
	if err != nil {
		return false
	}
	return strings.EqualFold(label, value)
}

// IsLabelNotSetTo checks if a specific label is not set to a given value.
// The comparison is done in a case insensitive way.
func (obj NamedObject) IsLabelNotSetTo(key, value string) bool {
	label, err := obj.GetString(NewPath(PathLabels, key))
	if err != nil {
		return true
	}
	return !strings.EqualFold(label, value)
}

// GetAnnotation will return the value of a given label.
// If the annotation is not set, an empty string and false is returned.
func (obj NamedObject) GetAnnotation(key string) (string, error) {
	return obj.GetString(NewPath(PathAnnotations, key))
}

// HasAnnotations returns true if an annotation section exists
func (obj NamedObject) HasAnnotations() bool {
	return obj.Has(PathAnnotations)
}

// IsAnnotationSetTo checks if a specific annotation is set to a given value.
// The comparison is done in a case insensitive way.
func (obj NamedObject) IsAnnotationSetTo(key, value string) bool {
	annotation, err := obj.GetString(NewPath(PathAnnotations, key))
	if err != nil {
		return false
	}
	return strings.EqualFold(annotation, value)
}

// IsAnnotationNotSetTo checks if a specific annotation is not set to a given value.
// The comparison is done in a case insensitive way.
func (obj NamedObject) IsAnnotationNotSetTo(key, value string) bool {
	annotation, err := obj.GetString(NewPath(PathAnnotations, key))
	if err != nil {
		return true
	}
	return !strings.EqualFold(annotation, value)
}

// SetName will set the name of the object.
func (obj NamedObject) SetName(value string) error {
	return obj.Set(PathMetadataName, value)
}

// SetName will set the namespace of the object.
func (obj NamedObject) SetNamespace(value string) error {
	return obj.Set(PathMetadataNamespace, value)
}

// SetAnnotation will set an annotation on the object.
// It will create the annotations section if it does not exist.
func (obj NamedObject) SetAnnotation(key, value string) error {
	return obj.Set(NewPath(PathAnnotations, key), value)
}

// SetAnnotation will set a label on the object.
// It will create the labels section if it does not exist.
func (obj NamedObject) SetLabel(key, value string) error {
	return obj.Set(NewPath(PathLabels, key), value)
}

// IsOfKind returns true if the object is of the given kind and/or apiVersion.
// Both kind and apiVersion can be an empty string, which translates to "any"
func (obj NamedObject) IsOfKind(kind, apiVersion string) bool {
	if kind != "" {
		value, err := obj.GetString(Path{"kind"})
		if err != nil || !strings.EqualFold(value, kind) {
			return false
		}
	}

	if apiVersion != "" {
		value, err := obj.GetString(Path{"apiVersion"})
		if err != nil || !strings.EqualFold(value, apiVersion) {
			return false
		}
	}

	return true
}

// CreateAddPatch generates an add patch based.
func (obj NamedObject) CreateAddPatch(path Path, value interface{}) PatchOperation {
	return NewPatchOperationAdd(path.ToJSONPath(), value)
}

// PatchField generates a replace patch.
func (obj NamedObject) CreateReplacePatch(path Path, value interface{}) PatchOperation {
	return NewPatchOperationReplace(path.ToJSONPath(), value)
}

// RemoveField generates a remove patch.
func (obj NamedObject) CreateRemovePatch(path Path) PatchOperation {
	return NewPatchOperationRemove(path.ToJSONPath())
}

// RemoveManagedFields removes managed fields from an object.
// See KubernetesManagedFields and FieldCleaner.
func (obj NamedObject) RemoveManagedFields() {
	ManagedFields.Clean(obj)
}

// ToJSON generates a JSON string out of this object
func (obj NamedObject) ToJSON() (string, error) {
	data, err := json.Marshal(obj)
	return string(data), err
}

// Hash calculates an ordered hash of the object.
func (obj NamedObject) Hash() (uint64, error) {
	hasher := xxhash.New()
	err := obj.getOrderedHash(hasher)
	return hasher.Sum64(), err
}

// Hash calculates an ordered hash of the object an returns a base64 encoded
// string.
func (obj NamedObject) HashStr() (string, error) {
	hasher := xxhash.New()
	err := obj.getOrderedHash(hasher)

	return base64.StdEncoding.EncodeToString(hasher.Sum([]byte{})), err
}

// getOrderedHash orders the keys in a NamedObject before creating an
// incremental hash on each key/value pair
func (obj NamedObject) getOrderedHash(hasher hash.Hash64) error {
	// Go maps are not ordered.
	// In order to get reproducible hashes, we need to sort each level.
	// We also cannot marshal to JSON and take a hash of this, as the resulting
	// JSON also has no ordering guarantees.

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()

	for _, k := range keys {
		hasher.Write([]byte(k))
		iv := obj[k]

		if err := doHash(hasher, k, iv); err != nil {
			return err
		}
	}

	return nil
}

// doHash caclulates the has for a key/value pair of a specfic type.
// Separated out of getOrderedHash so we can called it recursively during array
// iteration.
func doHash(hasher hash.Hash64, k string, iv interface{}) error {
	switch v := iv.(type) {
	case []byte:
		_, err := hasher.Write(v)
		return err

	case string:
		_, err := hasher.Write([]byte(v))
		return err

	case []string:
		var err error
		for _, str := range v {
			if _, err2 := hasher.Write([]byte(str)); err2 != nil {
				if err == nil {
					err = err2
				} else {
					err = errors.Wrapf(err, "failed to hash string in array for field %s: %v", k, err2)
				}
			}
		}
		return err

	case float32, float64:
		str := fmt.Sprintf("%f", v)
		_, err := hasher.Write([]byte(str))
		return err

	case int, int16, int32, int64:
		str := fmt.Sprintf("%d", v)
		_, err := hasher.Write([]byte(str))
		return err

	case uint, uint16, uint32, uint64:
		str := fmt.Sprintf("%u", v)
		_, err := hasher.Write([]byte(str))
		return err

	case bool:
		if v {
			_, err := hasher.Write([]byte("true"))
			return err
		}
		_, err := hasher.Write([]byte("false"))
		return err

	case NamedObject:
		return v.getOrderedHash(hasher)

	case []NamedObject:
		var err error
		for _, o := range v {
			if err2 := o.getOrderedHash(hasher); err2 != nil {
				if err == nil {
					err = err2
				} else {
					err = errors.Wrapf(err2, "failed to hash NamedObject in array for field %s: %v", k, err2)
				}
			}
		}
		return err

	case map[string]interface{}:
		o := NamedObject(v)
		return o.getOrderedHash(hasher)

	case []map[string]interface{}:
		var err error
		for _, msi := range v {
			o := NamedObject(msi)
			if err2 := o.getOrderedHash(hasher); err2 != nil {
				if err == nil {
					err = err2
				} else {
					err = errors.Wrapf(err2, "failed to hash map[string]interface{} for field %s: %v", k, err2)
				}
			}
		}
		return err

	case []interface{}:
		var err error
		for _, element := range v {
			if err2 := doHash(hasher, k, element); err2 != nil {
				if err == nil {
					err = err2
				} else {
					err = errors.Wrapf(err2, "failed to hash element in array for field %s: %v", k, err2)
				}
			}
		}
		return err

	default:
		return ErrUnsupportedHashType(fmt.Sprintf("cannot create hash for field %s of type %T", k, v))
	}
}

// Walk will iterate the path up until key is found or path cannot be matched.
// If key is found, the value of key and true is returned. Otherwise nil and
// false will be returned.
func (obj *NamedObject) Walk(path Path, args WalkArgs) (interface{}, error) {
	root := map[string]interface{}(*obj)
	return walk(root, path, args)
}

// GeneratePatch will reduce the given path so that only exisiting elements are
// included. The given value will be extended so that missing elements from the
// path will be created. Please note that path creation will fail if non-
// existing arrays are addressed using index notation.
func (obj NamedObject) GeneratePatch(path Path, value interface{}) (Path, interface{}, error) {
	if len(path) == 0 {
		return path, value, nil
	}

	validPath := Path{}
	_, err := obj.Walk(path, WalkArgs{
		MatchFunc: func(v interface{}, p Path) bool {
			validPath = p
			if GetArrayNotation(path[len(path)-1]) == ArrayNotationTraversal {
				// Traversal notation will be converted to index notation on match.
				// We need to keep the notation here in case we have an "append"
				// requested.
				validPath[len(validPath)-1] = "-"
			}
			return true
		},
		NotFoundFunc: func(p Path) {
			validPath = p
		},
	})

	// Full match or everything-but-last-key match
	if err == nil {
		return validPath, value, nil
	}

	// "Late" full match (last key does not exist)
	if len(validPath) == len(path) {
		return validPath, value, nil
	}

	// We should get ErrIsNotFound. Otherwise return the error
	if _, isNotFound := err.(ErrNotFound); !isNotFound {
		return validPath, value, err
	}

	firstIdx := len(validPath)

	// Generate the first node to attach the remaining hierarchy to
	var parentNode interface{}
	_, rootArrayNotation := path.IsArray(len(validPath) - 1)
	switch rootArrayNotation {
	case ArrayNotationInvalid:
		parentNode = map[string]interface{}{}

	case ArrayNotationTraversal:
		// If the array field does not exist, the traversal notation is missing and
		// we need to create an array is first node.
		if path[firstIdx] == "-" {
			parentNode = make([]interface{}, 1)
			firstIdx++
		}

	case ArrayNotationIndex:
		return validPath, value, ErrIndexNotation{}
	}

	extendedValue := parentNode

	// Helper function to add the current node to the parent node
	addToParent := func(key string, node interface{}) {
		switch parent := parentNode.(type) {
		case []interface{}:
			if key == "-" {
				parent[0] = node
			} else {
				parent[0] = map[string]interface{}{key: node}
			}

		case map[string]interface{}:
			parent[key] = node

		case nil:
			// Case: root is an existing array
			if key == "-" {
				extendedValue = []interface{}{node} // TODO: testcase root array-in-array
			} else {
				extendedValue = map[string]interface{}{key: node}
			}
		}
	}

	// Iterate but skip last key. This key will hold the value.
	for idx := firstIdx; idx < len(path); idx++ {
		key := path[idx]
		_, arrayNotation := path.IsArray(idx)

		switch arrayNotation {
		case ArrayNotationInvalid:
			// For the last element, skip map creation, as we will add "value" using
			// "key" after this loop
			if idx < len(path)-1 {
				newNode := make(map[string]interface{})
				addToParent(key, newNode)
				parentNode = newNode
			}

		case ArrayNotationTraversal:
			newNode := make([]interface{}, 1)
			addToParent(key, newNode)
			parentNode = newNode
			if key != "-" {
				idx++ // skip array notation
			}

		case ArrayNotationIndex:
			return validPath, value, ErrIndexNotation{}
		}
	}

	key := path[len(path)-1]
	addToParent(key, value)

	return validPath, extendedValue, nil
}

// walk is the internal implementation of the walk function, accepting different
// type of node objects.
func walk(node interface{}, path Path, args WalkArgs) (interface{}, error) {
	// Internal helper function to react on "not found"
	errNotFound := func(key string) (interface{}, error) {
		if args.NotFoundFunc != nil {
			walked := args.walkedPath
			if len(key) > 0 {
				walked = append(walked, key)
			}
			args.NotFoundFunc(walked)
		}
		return nil, ErrNotFound(key)
	}

	// If the path is empty we found the value.
	if len(path) == 0 {
		if args.MatchFunc != nil {
			if !args.MatchFunc(node, args.walkedPath) {
				return errNotFound("")
			}
		}

		value, err := args.onMutate(node)
		if err != nil {
			return nil, err
		}

		return value, nil
	}

	// Don't travers nil nodes
	if node == nil {
		return nil, ErrNotTraversable(args.getKey() + " is nil")
	}

	// We're still traversing through the path.
	// There's at least one more traversal step, i.e. len(path) >= 1.

	switch reflect.ValueOf(node).Kind() {
	// Node is a map
	case reflect.Map:
		object, ok := node.(map[string]interface{})
		if !ok {
			return nil, ErrNotTraversable(args.getKey() + " is not a map")
		}

		key := path[0]
		if GetArrayNotation(key) != ArrayNotationInvalid {
			return nil, ErrNotAnArray(args.getKey())
		}
		nextNode, exists := object[key]
		if !exists {
			if len(path) == 1 && args.MutateFunc != nil {
				// Make sure non-existing keys are created
				pseudoArgs := args.push(key, node)
				return pseudoArgs.onMutate(nil)
			}
			return errNotFound(key)
		}

		return walk(nextNode, path[1:], args.push(key, node))

	// Node is an array
	case reflect.Array, reflect.Slice:
		array, ok := node.([]interface{})
		if !ok {
			return nil, ErrNotTraversable(args.getKey() + " is not a slice")
		}

		arrayIdx := path[0]
		switch GetArrayNotation(arrayIdx) {
		// Explicit index
		case ArrayNotationIndex:
			idx, err := strconv.ParseInt(arrayIdx, 10, 32)
			if err != nil {
				return nil, err
			}
			if idx >= int64(len(array)) {
				return errNotFound(arrayIdx)
			}
			return walk(array[idx], path[1:], args.push(arrayIdx, node))

		// Traverse array
		case ArrayNotationTraversal:
			// Look for the first match only
			if !args.MatchAll {
				for idx, child := range array {
					idxStr := strconv.Itoa(idx)
					v, err := walk(child, path[1:], args.pushTraversal(idxStr, node))
					if err == nil {
						return v, nil
					}
				}
				return errNotFound("-")
			}

			// Try all indexes and collect matches in a list
			values := []interface{}{}
			for idx, child := range array {
				idxStr := strconv.Itoa(idx)
				v, err := walk(child, path[1:], args.pushTraversal(idxStr, node))
				if err == nil {
					values = append(values, v)
				}
				// Ignore errors in sub-paths
			}
			if len(values) == 0 {
				return errNotFound("-")
			}
			if len(values) == 1 {
				return values[0], nil
			}
			return values, nil

		// Array is missing traversal indicator
		default:
			return nil, ErrMissingArrayTraversal(args.getKey())
		}
	}

	return nil, ErrNotTraversable(args.getKey() + " is " + reflect.ValueOf(node).Kind().String())
}

// GetObjectKind implements the runtime.Object interface.
// Returns schema.EmptyObjectKind
func (obj NamedObject) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

// DeepCopyObject implements the runtime.Object interface.
func (obj NamedObject) DeepCopyObject() runtime.Object {
	new := make(NamedObject)
	maps.Copy(new, obj)
	return new
}

// NewEmptyInstance implements the runtime.Unstructured interface.
// Returns an empty NamedObject.
func (obj NamedObject) NewEmptyInstance() runtime.Unstructured {
	return NamedObject{}
}

// UnstructuredContent implements the runtime.Unstructured interface.
// Returns the object itself
func (obj NamedObject) UnstructuredContent() map[string]interface{} {
	return obj
}

// SetUnstructuredContent implements the runtime.Unstructured interface.
func (obj NamedObject) SetUnstructuredContent(new map[string]interface{}) {
	maps.Copy(obj, new)
}

// IsList implements the runtime.Unstructured interface.
// This function returns false
func (obj NamedObject) IsList() bool {
	return false
}

// EachListItem implements the runtime.Unstructured interface.
// This function does nothing.
func (obj NamedObject) EachListItem(func(runtime.Object) error) error {
	return nil
}

// EachListItemWithAlloc implements the runtime.Unstructured interface.
// This function does nothing.
func (obj NamedObject) EachListItemWithAlloc(func(runtime.Object) error) error {
	return nil
}
