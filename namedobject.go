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
	// The path contains the traversed path up to (including) the key that was
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
	p, v, _ := obj.GeneratePatch(PathMetadataName, value)
	obj.Set(p, v)
}

// SetName will set the namespace of the object.
func (obj NamedObject) SetNamespace(value string) {
	p, v, _ := obj.GeneratePatch(PathMetadataNamespace, value)
	obj.Set(p, v)
}

// SetAnnotation will set an annotation on the object.
// It will create the annotations section if it does not exist.
func (obj NamedObject) SetAnnotation(key, value string) {
	p, v, _ := obj.GeneratePatch(NewPath(PathAnnotations, key), value)
	obj.Set(p, v)
}

// SetAnnotation will set a label on the object.
// It will create the labels section if it does not exist.
func (obj NamedObject) SetLabel(key, value string) {
	p, v, _ := obj.GeneratePatch(NewPath(PathLabels, key), value)
	obj.Set(p, v)
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
			fmt.Println("match:", validPath)
			return true
		},
		NotFoundFunc: func(p Path) {
			validPath = p
			fmt.Println("not found:", p)
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
		return validPath, value, ErrIndexNotation("")
	}

	fmt.Println("Existing", validPath)
	fmt.Println("Processing", path[firstIdx:])

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
		fmt.Println(key)

		switch arrayNotation {
		case ArrayNotationInvalid:
			// For the last element, skip map creation, as we will add "value" using
			// "key" after this loop
			if idx < len(path)-1 {
				fmt.Println("is a map")
				newNode := make(map[string]interface{})
				addToParent(key, newNode)
				parentNode = newNode
			}

		case ArrayNotationTraversal:
			fmt.Println("is an array")
			newNode := make([]interface{}, 1)
			addToParent(key, newNode)
			parentNode = newNode
			if key != "-" {
				idx++ // skip array notation
			}

		case ArrayNotationIndex:
			return validPath, value, ErrIndexNotation("")
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
		value := node
		if args.MutateFunc != nil {
			value = args.MutateFunc(value)

			// Modify the hierarchy. This requires modification of the parent node
			// This is either an array or a map. As we are working on a copy of the
			// "header data", we need to propagate the changes back.
			switch reflect.ValueOf(args.parent).Kind() {
			case reflect.Map:
				// TODO: This fails if a NamedObject is passed into this function
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
			if !args.MatchFunc(value, args.walkedPath) {
				return errNotFound("")
			}
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
		if GetArrayNotation(path[0]) != ArrayNotationInvalid {
			return nil, ErrNotAnArray(args.getNodeKey())
		}
		nextNode, exists := object[path[0]]
		if !exists {
			return errNotFound(path[0])
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
				return errNotFound(arrayIdx)
			}
			return walk(array[idx], path[1:], args.push(node, arrayIdx, func(p interface{}) { args.onParentChange(p) }))

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
				return errNotFound("-")
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
				return errNotFound("-")
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
