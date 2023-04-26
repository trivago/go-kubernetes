package kubernetes

import (
	"reflect"
	"strconv"
)

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

	appendOnMutate bool
	walkedPath     Path
	parent         interface{}
	previousArgs   *WalkArgs
}

// push creates a new args argument for the next recursion level.
// currentNode expects the node currently processed
// currentKey expects the key of the currently processed node
// onResetSelf is a function called if the contents of currentNode changed
func (src *WalkArgs) push(currentKey string, currentNode interface{}) WalkArgs {
	args := *src
	args.walkedPath = NewPath(src.walkedPath, currentKey)
	args.parent = currentNode
	args.previousArgs = src
	args.appendOnMutate = false
	return args
}

// pushTraversal calls push, but sets a flag so that any change will be appended
// and not overwrite. This function is used on arrays with traversal notation,
// as the key will always be a numeric index, hence discarding the "traversl"
// information.
func (src *WalkArgs) pushTraversal(currentKey string, currentNode interface{}) WalkArgs {
	args := src.push(currentKey, currentNode)
	args.appendOnMutate = true
	return args
}

// getKey returns the name of the current key
func (args WalkArgs) getKey() string {
	if len(args.walkedPath) == 0 {
		return ""
	}
	return args.walkedPath[len(args.walkedPath)-1]
}

// changeParentTo update the hierarchy, so that changes in the parent are
// properly reflected.
func (args WalkArgs) changeParentTo(newParent interface{}) {
	if args.previousArgs == nil {
		return
	}

	parentParentNode := args.previousArgs.parent
	nodeKey := args.previousArgs.getKey()

	switch reflect.ValueOf(parentParentNode).Kind() {
	case reflect.Map:
		parentParent := parentParentNode.(map[string]interface{})
		parentParent[nodeKey] = newParent

	case reflect.Array, reflect.Slice:
		parentParent := parentParentNode.([]interface{})
		nodeIdx, _ := strconv.ParseInt(nodeKey, 10, 32)
		parentParent[nodeIdx] = newParent
	}
}

// onMutate is a wrapper around MutateFunc, reacting correctly on changed values
// and adding them to the hierarchy
func (args WalkArgs) onMutate(value interface{}) (interface{}, error) {
	if args.MutateFunc == nil {
		return value, nil
	}

	key := args.getKey()
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
			delete(parent, key)
		} else {
			parent[key] = value
		}
		args.changeParentTo(parent)

	case reflect.Array, reflect.Slice:
		parent, ok := args.parent.([]interface{})
		if !ok {
			return nil, ErrNotTraversable("parent is not a slice")
		}
		idx, _ := strconv.ParseInt(key, 10, 32)

		if value == nil {
			parent = append(parent[:idx], parent[idx+1:]...)
		} else if args.appendOnMutate {
			parent = append(parent, value)
		} else {
			parent[idx] = value
		}
		args.changeParentTo(parent)

	default:
		// nil value, no change required
	}

	return value, nil
}
