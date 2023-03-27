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

type NamespacedObject map[string]interface{}

type walkArgs struct {
	createPath   bool
	matchAll     bool
	path         []string
	matchFunc    func(value interface{}, path []string) bool
	mutateFunc   func(value interface{}) interface{}
	notFoundFunc func(path []string)
}

var (
	pathMetadata       = []string{"metadata"}
	pathLabels         = []string{"metadata", "labels"}
	pathAnnotations    = []string{"metadata", "annotations"}
	pathOwnerReference = []string{"metadata", "ownerReferences"}
)

// NamespacedObjectFromUnstructured converts a raw runtime object intor a
// namespaced object. If the object does not have name or namespace set an
// error will be returned.
func NamespacedObjectFromRaw(data *runtime.RawExtension) (NamespacedObject, error) {
	if data.Raw == nil {
		if data.Object == nil {
			return NamespacedObject{}, fmt.Errorf("no data found in raw object")
		}
		var err error
		if data.Raw, err = jsoniter.Marshal(data.Object); err != nil {
			return NamespacedObject{}, err
		}
	}

	parsed := unstructured.Unstructured{
		Object: make(map[string]interface{}),
	}

	// Read JSON data into a map and change object name and namespace
	if err := jsoniter.Unmarshal(data.Raw, &parsed.Object); err != nil {
		return NamespacedObject{}, err
	}

	return NamespacedObjectFromUnstructured(parsed)
}

// NamespacedObjectFromUnstructured converts an unstructured Kubernetes object
// into a namespaced object. If the object does not have name or namespace set
// an error will be returned.
func NamespacedObjectFromUnstructured(unstructuredObj unstructured.Unstructured) (NamespacedObject, error) {
	obj := NamespacedObject(unstructuredObj.Object)

	// "generateName" is used by pods before a, e.g., ReplicaSet controler
	// processed the pod.
	if !obj.Has(pathMetadata, "name") && !obj.Has(pathMetadata, "generateName") {
		return obj, fmt.Errorf("object does not have a name set")
	}

	return obj, nil
}

// StringToPath generates a path array from a json path.
func StringToPath(path string) []string {
	return strings.Split(path, ".")
}

// SplitPathKey splits a path array so that the last elemnt is returned as a
// separate string. The path object itself will not be copied.
func SplitPathKey(path []string) ([]string, string) {
	lastIdx := len(path) - 1
	if lastIdx < 0 {
		return path, ""
	}
	return path[0:lastIdx], path[lastIdx]
}

// Find looks for a key inside path with the given value and returns all
// matching paths. If nil is passed as a value, all pathes containing the key
// will be returned.
func (obj NamespacedObject) Find(path []string, key string, value interface{}) [][]string {
	paths := [][]string{}

	matchValue := func(v interface{}, path []string) bool {
		if value == nil || reflect.DeepEqual(v, value) {
			paths = append(paths, path)
			return true
		}
		return false
	}

	obj.walk(path, key, walkArgs{
		matchAll:  true,
		matchFunc: matchValue,
	})

	return paths
}

// FindFirst looks for a key inside path with the given value and returns the
// first matching path. If nil is passed as a value, the first path with the key
// set will be returned.
func (obj NamespacedObject) FindFirst(path []string, key string, value interface{}) []string {
	foundPath := []string{}

	matchValue := func(v interface{}, path []string) bool {
		if value == nil || reflect.DeepEqual(v, value) {
			foundPath = path
			return true
		}
		return false
	}

	obj.walk(path, key, walkArgs{
		matchAll:  false,
		matchFunc: matchValue,
	})

	return foundPath
}

// Get will return an object for a given path.
// If the object or any part of the path does not exist, nil is returned.
// If an unindexed array notation is used ("[]") the first matching path is
// returned.
func (obj NamespacedObject) Get(path []string, key string) interface{} {
	result, _ := obj.walk(path, key, walkArgs{})
	return result
}

// Set will set a value for a given key on a given path.
// The path will be created if not existing. Missing arrays in the path will be
// created but existing arrays will never be extended.
// If any part of the path is not a map[string]interface{} or a slice of the
// former, or the value cannot be set for any other reason, the function will
// return false.
func (obj NamespacedObject) Set(path []string, key string, value interface{}) bool {
	setValue := func(interface{}) interface{} {
		// TODO: Support arrays
		return value
	}

	_, ok := obj.walk(path, key, walkArgs{
		matchAll:   true,
		createPath: true,
		mutateFunc: setValue,
	})
	return ok
}

// Delete will remove a given key on a given path.
// If an unindexed array notation is used ("[]") the first matching path will be
// used, which might lead to the key not being deleted.
// If the path is not valid because a key in the path does not exist, is no
// map or array, false will be returned. If the key is deleted or does not exist,
// true will be returned.
func (obj NamespacedObject) Delete(path []string, key string) bool {
	deleteKey := func(interface{}) interface{} {
		return nil
	}

	_, found := obj.walk(path, key, walkArgs{
		matchAll:   true,
		mutateFunc: deleteKey,
	})
	return found
}

// Has will return true if a key on a given path is set.
func (obj NamespacedObject) Has(path []string, key string) bool {
	_, found := obj.walk(path, key, walkArgs{})
	return found
}

// GetString will return a string value assigned to a given key on a given path.
// If the object is not a string or the path or key does not exist, false is
// and an empty string returned.
func (obj NamespacedObject) GetString(path []string, key string) (string, bool) {
	if value := obj.Get(path, key); value == nil {
		return "", false
	} else {
		str, ok := value.(string)
		return str, ok
	}
}

// GetName will return the name of the object.
// The name can be a prefix if a pod is processed before it has been processed
// by the corresponding, e.g., ReplicaSet controller.
// If the name is not set, an empty string is returned.
func (obj NamespacedObject) GetName() string {
	if name, ok := obj.GetString(pathMetadata, "name"); ok {
		return name
	}
	if namePrefix, ok := obj.GetString(pathMetadata, "generateName"); ok {
		return namePrefix
	}

	return ""
}

// GetName will return the namespace of the object.
// If the namespace is not set, an empty string is returned.
func (obj NamespacedObject) GetNamespace() string {
	if namespace, ok := obj.GetString(pathMetadata, "namespace"); ok {
		return namespace
	}

	return ""
}

// GetOwnerKind returns the resource kind of an owning resource, e.g.,
// ReplicaSet if the pod is managed by a ReplicaSet
func (obj NamespacedObject) GetOwnerKind() string {
	if owner, ok := obj.GetString(pathOwnerReference, "kind"); ok {
		return owner
	}
	return ""
}

// GetLabel will return the value of a given label.
// If the label is not set, an empty string and false is returned.
func (obj NamespacedObject) GetLabel(key string) (string, bool) {
	return obj.GetString(pathLabels, key)
}

// HasLabels returns true if a labels section exists
func (obj NamespacedObject) HasLabels() bool {
	return obj.Has(SplitPathKey(pathLabels))
}

// IsLabelSetTo checks if a specific label is set to a given value.
// The comparison is done in a case insensitive way.
func (obj NamespacedObject) IsLabelSetTo(key, value string) bool {
	label, ok := obj.GetString(pathLabels, key)
	if !ok {
		return false
	}
	return strings.EqualFold(label, value)
}

// IsLabelNotSetTo checks if a specific label is not set to a given value.
// The comparison is done in a case insensitive way.
func (obj NamespacedObject) IsLabelNotSetTo(key, value string) bool {
	label, ok := obj.GetString(pathLabels, key)
	if !ok {
		return true
	}
	return !strings.EqualFold(label, value)
}

// GetAnnotation will return the value of a given label.
// If the annotation is not set, an empty string and false is returned.
func (obj NamespacedObject) GetAnnotation(key string) (string, bool) {
	return obj.GetString(pathAnnotations, key)
}

// HasAnnotations returns true if an annotation section exists
func (obj NamespacedObject) HasAnnotations() bool {
	return obj.Has(SplitPathKey(pathAnnotations))
}

// IsAnnotationSetTo checks if a specific annotation is set to a given value.
// The comparison is done in a case insensitive way.
func (obj NamespacedObject) IsAnnotationSetTo(key, value string) bool {
	annotation, ok := obj.GetString(pathAnnotations, key)
	if !ok {
		return false
	}
	return strings.EqualFold(annotation, value)
}

// IsAnnotationNotSetTo checks if a specific annotation is not set to a given value.
// The comparison is done in a case insensitive way.
func (obj NamespacedObject) IsAnnotationNotSetTo(key, value string) bool {
	annotation, ok := obj.GetString(pathAnnotations, key)
	if !ok {
		return true
	}
	return !strings.EqualFold(annotation, value)
}

// SetName will set the name of the object.
func (obj NamespacedObject) SetName(value string) {
	obj.Set(pathMetadata, "name", value)
}

// SetName will set the namespace of the object.
func (obj NamespacedObject) SetNamespace(value string) {
	obj.Set(pathMetadata, "namespace", value)
}

// SetAnnotation will set an annotation on the object.
// It will create the annotations section if it does not exist.
func (obj NamespacedObject) SetAnnotation(key, value string) {
	obj.Set(pathAnnotations, key, value)
}

// IsOfKind returns true if the object is of the given kind and/or apiVersion.
// Both kind and apiVersion can be an empty string, which translates to "any"
func (obj NamespacedObject) IsOfKind(kind, apiVersion string) bool {
	if kind != "" {
		value, isString := obj.Get([]string{}, "kind").(string)
		if !isString || !strings.EqualFold(value, kind) {
			return false
		}
	}

	if apiVersion != "" {
		value, isString := obj.Get([]string{}, "apiVersion").(string)
		if !isString || !strings.EqualFold(value, apiVersion) {
			return false
		}
	}

	return true
}

func (obj NamespacedObject) FixPatchPath(path []string, value interface{}) ([]string, interface{}) {
	if len(path) == 0 {
		return path, value
	}

	var validPath []string
	lastPathIdx := len(path) - 1
	key := path[lastPathIdx]

	_, fullMatch := obj.walk(path[:lastPathIdx], key, walkArgs{
		notFoundFunc: func(p []string) {
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

// CreateAddPatch generates an add patch based.
func (obj NamespacedObject) CreateAddPatch(path []string, value interface{}) PatchOperation {
	jsonPath := EscapeJSONPath(path)
	return NewPatchOperationAdd(jsonPath, value)
}

// PatchField generates a replace patch.
func (obj NamespacedObject) CreateReplacePatch(path []string, value interface{}) PatchOperation {
	jsonPath := EscapeJSONPath(path)
	return NewPatchOperationReplace(jsonPath, value)
}

// RemoveField generates a remove patch.
func (obj NamespacedObject) CreateRemovePatch(path []string) PatchOperation {
	jsonPath := EscapeJSONPath(path)
	return NewPatchOperationRemove(jsonPath)
}

// RemoveManagedFields removes managed fields from an object.
// See KubernetesManagedFields and FieldCleaner.
func (obj NamespacedObject) RemoveManagedFields() {
	KubernetesManagedFields.Clean(obj)
}

// Hash calculates an ordered hash of the object.
func (obj NamespacedObject) Hash() (uint64, error) {
	hasher := xxhash.New()
	err := obj.getOrderedHash(hasher)
	return hasher.Sum64(), err
}

// Hash calculates an ordered hash of the object an returns a base64 encoded
// string.
func (obj NamespacedObject) HashStr() (string, error) {
	hasher := xxhash.New()
	err := obj.getOrderedHash(hasher)

	return base64.StdEncoding.EncodeToString(hasher.Sum([]byte{})), err
}

// getOrderedHash orders the keys in a NamespacedObject before creating an
// incremental hash on each key/value pair
func (obj NamespacedObject) getOrderedHash(hasher hash.Hash64) error {
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

	case NamespacedObject:
		v.getOrderedHash(hasher)
	case []NamespacedObject:
		for _, o := range v {
			o.getOrderedHash(hasher)
		}

	case map[string]interface{}:
		o := NamespacedObject(v)
		o.getOrderedHash(hasher)
	case []map[string]interface{}:
		for _, msi := range v {
			o := NamespacedObject(msi)
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

// Has will return true if a key on a given path is set.
func (obj NamespacedObject) walk(searchPath []string, key string, args walkArgs) (interface{}, bool) {
	node := obj
	path := make([]string, 0, len(args.path)+len(searchPath)+1)
	path = append(path, args.path...)

	// Path contains either maps or arrays, as key, evtually holding other value
	// types is handled at the end. I.e. this loop only processes structural
	// elements.

	for searchPathIdx, searchPathElement := range searchPath {
		searchPathKey := searchPathElement
		arrayNotation := searchPathElement[len(searchPathElement)-1] == ']'

		if arrayNotation {
			searchPathKey = searchPathElement[:strings.LastIndexByte(searchPathElement, '[')]
		}

		// Does the key exist?
		// TODO: This test has to change if we accept node being an array, too.
		child, ok := node[searchPathKey]
		if !ok {
			if !args.createPath {
				args.onNotFound(path)
				return nil, false // not found
			}

			// Create the node and continue processing it
			if arrayNotation {
				child = []map[string]interface{}{
					make(map[string]interface{}),
				}
			} else {
				child = make(map[string]interface{})
			}
			node[searchPathKey] = child
		}

		// Walk an array / access a specific element
		if arrayNotation {
			childArray, ok := child.([]interface{})
			if !ok {
				args.onNotFound(path) // as-if not found
				return nil, false     // error: not an array
			}

			// If we do not find the element, return the array element in "existing"
			// notation.
			wildcardPath := append(path, searchPathKey+"[]")

			if len(childArray) == 0 {
				args.onNotFound(wildcardPath) // child not found
				return nil, false             // not found, empty
			}

			// Read array index from string
			arrayIdxStr := searchPathElement[strings.LastIndexByte(searchPathElement, '[')+1 : len(searchPathElement)-1]

			// Direct element access
			if len(arrayIdxStr) > 0 {
				arrayIdx, err := strconv.Atoi(arrayIdxStr)
				if err != nil {
					return nil, false // error: invalid array syntax
				}
				if arrayIdx >= len(childArray) {
					args.onNotFound(wildcardPath) // child not found
					return nil, false             // not found, out of bounds
				}

				path = append(path, searchPathElement) // Notation is already ok

				// TODO: Array-in-array does not work as we expect node to always be
				//       a map[string]interface{}. We can fix this by making it an
				//       interface{} but that requires more casting.
				mapElement, ok := childArray[arrayIdx].(map[string]interface{})
				if !ok {
					args.onNotFound(path) // as-if not found
					return nil, false     // error: not a map
				}

				node = mapElement
				continue
			}

			// Search for element(s)
			// If multi-match is requested, collect all matches
			matches := []interface{}{}

			for arrayIdx, element := range childArray {
				// TODO: Array-in-array does not work (see above)
				mapElement, ok := element.(map[string]interface{})
				if !ok {
					continue
				}

				// Process element via sub-tree call
				// Done recursively as we need to "pop" paths that don't yield a result
				// Pass the path up to this point to the nested call.
				// Note: copy because assign creates an implicit reference, conflicting
				//       with wildcardPath, which is also a reference.
				args.path = make([]string, 0, len(path)+1)
				args.path = append(args.path, path...)
				args.path = append(args.path, fmt.Sprintf("%s[%d]", searchPathKey, arrayIdx))

				if result, found := NamespacedObject(mapElement).walk(searchPath[searchPathIdx+1:], key, args); found {
					if !args.matchAll {
						return result, true // found
					}
					matches = append(matches, result)
				}
			}

			if len(matches) == 0 {
				args.onNotFound(wildcardPath) // child not found
				return nil, false             // not found
			}

			return matches, true
		}

		// Walk the tree
		if node, ok = child.(map[string]interface{}); !ok {
			args.onNotFound(path) // as-if not found
			return nil, false     // error: not a map
		}

		path = append(path, searchPathKey)
	}

	// TODO: key cannot be an array

	value, ok := node[key]

	if ok && args.matchFunc != nil {
		path = append(path, key)
		if !args.matchFunc(value, path) {
			args.onNotFound(path) // override existing
			return nil, false
		}
	}

	if args.mutateFunc != nil {
		newValue := args.mutateFunc(value)
		if newValue != nil {
			node[key] = newValue
			return newValue, true
		}
		delete(node, key)
		return nil, true
	}

	if !ok {
		args.onNotFound(path) // key not found
	}

	return value, ok
}

func (args walkArgs) onNotFound(path []string) {
	if args.notFoundFunc != nil {
		args.notFoundFunc(path)
	}
}
