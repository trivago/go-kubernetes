package kubernetes

import (
	"encoding/base64"
	"fmt"
	"hash"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/cespare/xxhash"
	jsoniter "github.com/json-iterator/go"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// NamedObject represents a kubernetes object and provides common functionality
// such as patch generators or accessing common fields.
type NamedObject map[string]interface{}

// WalkArgs is the parameter set passed to the walk function.
type WalkArgs struct {
	// MatchAll will iterate over all matches when set to true.
	MatchAll bool

	// MatchFunc is called whenever a path is found to be matching.
	// The path is the resolved path, i.e. array search notation is transformed
	// into array index notation.
	MatchFunc func(value interface{}, p Path) bool

	// MutateFunc allows a value to be modified or deleted after match.
	MutateFunc func(value interface{}) interface{}

	// NotFoundFunc is alled whenever walk needs to abort path walking.
	// The path contains the traversed path up to (and excluding) the key that was
	// not found,
	NotFoundFunc func(p Path)

	walkedPath         Path
	parent             interface{}
	onParentChangeFunc func(interface{})
}

// NamedObjectFromUnstructured converts a raw runtime object intor a
// namespaced object. If the object does not have name or namespace set an
// error will be returned.
func NamedObjectFromRaw(data *runtime.RawExtension) (NamedObject, error) {
	if data.Raw == nil {
		if data.Object == nil {
			return NamedObject{}, fmt.Errorf("no data found in raw object")
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
		return obj, fmt.Errorf("object does not have a name set")
	}

	return obj, nil
}

// Find looks for a path with the given value and returns all matching paths.
// If nil is passed as a value, all full matching paths will be returned.
func (obj NamedObject) FindAll(path Path, value interface{}) []Path {
	matchedPaths := []Path{}

	matchValue := func(v interface{}, path Path) bool {
		if value == nil || reflect.DeepEqual(v, value) {
			matchedPaths = append(matchedPaths, path)
			return true
		}
		return false
	}

	obj.Walk(path, WalkArgs{
		MatchAll:  true,
		MatchFunc: matchValue,
	})

	return matchedPaths
}

// FindFirst looks for a path with the given value and returns the first,
// resolved, matching path. If nil is passed as a value just the path will be
// matched.
func (obj NamedObject) FindFirst(path Path, value interface{}) Path {
	matchedPath := Path{}

	matchValue := func(v interface{}, path Path) bool {
		if value == nil || reflect.DeepEqual(v, value) {
			matchedPath = path
			return true
		}
		return false
	}

	obj.Walk(path, WalkArgs{
		MatchAll:  false,
		MatchFunc: matchValue,
	})

	return matchedPath
}

// Get will return an object for a given path.
// If the object or any part of the path does not exist, nil is returned.
// If an unindexed array notation is used ("[]") the first matching path is
// returned.
func (obj NamedObject) Get(path Path) (interface{}, error) {
	return obj.Walk(path, WalkArgs{})
}

// Set will set a value for a given key on a given path.
// The path will be created if not existing. Missing arrays in the path will be
// created but existing arrays will never be extended.
// If any part of the path is not a map[string]interface{} or a slice of the
// former, or the value cannot be set for any other reason, the function will
// return false.
func (obj NamedObject) Set(path Path, value interface{}) error {
	setValue := func(interface{}) interface{} {
		return value
	}

	_, err := obj.Walk(path, WalkArgs{
		MatchAll:   true,
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
		MatchAll:   true,
		MutateFunc: deleteKey,
	})
	return err
}

// Has will return true if a key on a given path is set.
func (obj NamedObject) Has(path Path) bool {
	_, err := obj.Walk(path, WalkArgs{})
	return err != nil
}

// GetString will return a string value assigned to a given key on a given path.
// If the object is not a string or the path or key does not exist, false is
// and an empty string returned.
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
func (obj NamedObject) SetName(value string) {
	// p, k := obj.GeneratePatch(PathMetadataNamespace, value)
	obj.Set(PathMetadataName, value)
}

// SetName will set the namespace of the object.
func (obj NamedObject) SetNamespace(value string) {
	// p, k := obj.GeneratePatch(PathMetadataNamespace, value)
	obj.Set(PathMetadataNamespace, value)
}

// SetAnnotation will set an annotation on the object.
// It will create the annotations section if it does not exist.
func (obj NamedObject) SetAnnotation(key, value string) {
	// p, k := obj.GeneratePatch(PathMetadataNamespace, value)
	obj.Set(NewPath(PathAnnotations, key), value)
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
		hasher.Write(v)
	case string:
		hasher.Write([]byte(v))
	case []string:
		for _, str := range v {
			hasher.Write([]byte(str))
		}

	case float32, float64:
		str := fmt.Sprintf("%f", v)
		hasher.Write([]byte(str))
	case int, int16, int32, int64:
		str := fmt.Sprintf("%d", v)
		hasher.Write([]byte(str))
	case uint, uint16, uint32, uint64:
		str := fmt.Sprintf("%u", v)
		hasher.Write([]byte(str))

	case bool:
		if v {
			hasher.Write([]byte("true"))
		} else {
			hasher.Write([]byte("false"))
		}

	case NamedObject:
		v.getOrderedHash(hasher)
	case []NamedObject:
		for _, o := range v {
			o.getOrderedHash(hasher)
		}

	case map[string]interface{}:
		o := NamedObject(v)
		o.getOrderedHash(hasher)
	case []map[string]interface{}:
		for _, msi := range v {
			o := NamedObject(msi)
			o.getOrderedHash(hasher)
		}
	case []interface{}:
		for _, element := range v {
			if err := doHash(hasher, k, element); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("Cannot create hash for field %s of type %T", k, v)
	}
	return nil
}

/*
func (obj NamedObject) FixPatchPath(path []string, value interface{}) ([]string, interface{}) {
	if len(path) == 0 {
		return path, value
	}

	var validPath []string
	lastPathIdx := len(path) - 1
	key := path[lastPathIdx]

	_, fullMatch := obj.Walk(path[:lastPathIdx], key, WalkArgs{
		NotFoundFunc: func(p []string) {
			validPath = p
		},
	})

	// Even if "key" does not exist, we have a valid path, as key will be generated
	if fullMatch {
		return path, value
	}

	var (
		nextNode      interface{}
		parent        interface{}
		extendedValue interface{}
	)

	// Add the next key to it, as a non-existing key needs to be the last element
	// of the path.
	nextKey := path[len(validPath)]

	if nextKey[len(nextKey)-1] == ']' {
		// Array keys need special handling as validPath does not distinguish
		// between "next key not found" and "subpath not part of next key".
		nextKey = nextKey[:strings.LastIndexByte(nextKey, '[')]

		if obj.Has(validPath, nextKey) {
			nextKey += "[]"
		} else {
			// value needs to be wrapped in an array
			if len(validPath) == len(path)-1 {
				value = []interface{}{value}
			} else {
				extendedValue = []interface{}{} // TODO: test case missing
			}
		}
	}
	validPath = append(validPath, nextKey)

	// "late full match"
	if len(validPath) == len(path) {
		return validPath, value
	}

	// At minimum, "key" is left

	// Walk remaining path, excluding key.
	// All elements are either an array or a map
	for idx := len(validPath); idx < len(path)-1; idx++ {
		element := path[idx]
		if element[len(element)-1] == ']' {
			element = element[:strings.LastIndexByte(element, '[')]
			nextNode = make([]interface{}, 1)
		} else {
			nextNode = make(map[string]interface{})
		}

		switch p := parent.(type) {
		case []interface{}: // TODO: test case missing
			p[0] = map[string]interface{}{
				element: nextNode,
			}
		case map[string]interface{}: // TODO: test case missing
			p[element] = nextNode
		case nil:
			extendedValue = map[string]interface{}{
				element: nextNode,
			}
		}
		parent = nextNode
	}

	// Add last key
	// Normalize it first
	if key[len(key)-1] == ']' {
		key = key[:strings.LastIndexByte(key, '[')]
		value = []interface{}{value}
	}

	switch p := parent.(type) {
	case []interface{}:
		p[0] = map[string]interface{}{
			key: value,
		}
	case map[string]interface{}:
		p[key] = value
	case nil:
		if key[len(key)-1] == ']' { // TODO: test case missing
			arrayKey := key[:strings.LastIndexByte(key, '[')] + "[]"
			extendedValue = map[string]interface{}{
				arrayKey: []interface{}{value},
			}
		} else {
			extendedValue = map[string]interface{}{
				key: value,
			}
		}
	}

	return validPath, extendedValue
}
*/

// Walk will iterate the path up until key is found or path cannot be matched.
// If key is found, the value of key and true is returned. Otherwise nil and
// false will be returned.
func (obj *NamedObject) Walk(path Path, args WalkArgs) (interface{}, error) {
	root := map[string]interface{}(*obj)
	return walk(root, path, args)
}

func walk(node interface{}, path Path, args WalkArgs) (interface{}, error) {
	// Internal helper function to react on "not found"
	errNotFound := func() (interface{}, error) {
		if args.NotFoundFunc != nil {
			args.NotFoundFunc(args.walkedPath)
		}
		return nil, ErrNotFound(path[0])
	}

	// If the path is empty we found the value.
	if len(path) == 0 {
		value := node
		if args.MutateFunc != nil {
			value = args.MutateFunc(value)

			// Modify the hierarchy. This requires modification of the parent node
			// This is either an array or a map. As we are working on a copy of the
			// "header data", we need to propagate the changes back.
			switch reflect.ValueOf(args.parent).Kind() {
			case reflect.Map:
				parent, ok := args.parent.(map[string]interface{})
				if !ok {
					return nil, ErrNotTraversable("parent is not a map")
				}
				if value == nil {
					delete(parent, args.getNodeKey())
				} else {
					parent[args.getNodeKey()] = value
				}
				args.onParentChange(parent)

			case reflect.Array, reflect.Slice:
				parent, ok := args.parent.([]interface{})
				if !ok {
					return nil, ErrNotTraversable("parent is not a slice")
				}
				idx, _ := strconv.ParseInt(args.getNodeKey(), 10, 32)

				if value == nil {
					parent = append(parent[:idx], parent[idx+1:]...)
				} else {
					parent[idx] = value
				}
				args.onParentChange(parent)

			default:
				// nil value, no change required
			}
		}

		if args.MatchFunc != nil {
			args.MatchFunc(value, args.walkedPath)
		}

		return value, nil
	}

	// Don't travers nil nodes
	if node == nil {
		return nil, ErrNotTraversable(args.getNodeKey() + " is nil")
	}

	// We're still traversing through the path.
	// There's at least one more traversal step, i.e. len(path) >= 1.

	switch reflect.ValueOf(node).Kind() {
	case reflect.Map:
		object, ok := node.(map[string]interface{})
		if !ok {
			return nil, ErrNotTraversable(args.getNodeKey() + " is not a map")
		}
		nextNode, exists := object[path[0]]
		if !exists {
			return errNotFound()
		}
		return walk(nextNode, path[1:], args.push(node, path[0], func(p interface{}) {
			object[path[0]] = p
		}))

	case reflect.Array, reflect.Slice:
		array, ok := node.([]interface{})
		if !ok {
			return nil, ErrNotTraversable(args.getNodeKey() + " is not a slice")
		}

		switch GetArrayNotation(path[0]) {
		case ArrayNotationIndex:
			// Explicit index traversal
			arrayIdx := path[0]
			idx, err := strconv.ParseInt(arrayIdx, 10, 32)
			if err != nil {
				return nil, err
			}
			if idx >= int64(len(array)) {
				return errNotFound()
			}
			return walk(array[idx], path[1:], args.push(node, arrayIdx, func(p interface{}) {
				args.onParentChange(p)
			}))

		case ArrayNotationTraversal:
			if !args.MatchAll {
				// Look for the first match only
				for idx, child := range array {
					idxStr := strconv.Itoa(idx)
					v, err := walk(child, path[1:], args.push(node, idxStr, func(p interface{}) {
						args.onParentChange(p)
					}))
					if err == nil {
						return v, nil
					}
				}
				return errNotFound()
			}

			// Try all indexes and collect matches in a list
			values := []interface{}{}
			for idx, child := range array {
				idxStr := strconv.Itoa(idx)
				v, err := walk(child, path[1:], args.push(node, idxStr, func(p interface{}) {
					args.onParentChange(p)
					node = p // make sure we pass the modified array to the next element
				}))
				if err == nil {
					values = append(values, v)
				}
				// Ignore errors in sub-paths
			}
			if len(values) == 0 {
				return errNotFound()
			}
			if len(values) == 1 {
				return values[0], nil
			}
			return values, nil

		default:
			return nil, ErrMissingArrayTraversal(args.getNodeKey())
		}
	}

	return nil, ErrNotTraversable(args.getNodeKey() + " is " + reflect.ValueOf(node).Kind().String())
}

// push creates a new args argument for the next recursion level.
// currentNode expects the node currently processed
// currentKey expects the key of the currently processed node
// onParentChange is a function called if the contents of currentNode changed
func (src WalkArgs) push(currentNode interface{}, currentKey string, onParentChange func(interface{})) WalkArgs {
	args := src
	args.walkedPath = NewPath(src.walkedPath, currentKey)
	args.parent = currentNode
	args.onParentChangeFunc = onParentChange
	return args
}

// onParentChange is a wrapper around onParentChangeFunc to avoid nil calls
func (args WalkArgs) onParentChange(p interface{}) {
	if args.onParentChangeFunc != nil {
		args.onParentChangeFunc(p)
	}
}

// getNodeKey returns the name of the current key
func (args WalkArgs) getNodeKey() string {
	if len(args.walkedPath) == 0 {
		return ""
	}
	return args.walkedPath[len(args.walkedPath)-1]
}
