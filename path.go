package kubernetes

import "strings"

// Path holds a list of path elements that can be used to traverse a
// namedObject. Arrays access is denoted with 2 elements in the list: the name
// of the array and the traversalNotation. The later is either "-" for "any" or
// a number denoting the index.
type Path []string

// ArrayNotation defines the type of an array index notation. Either Index, for
// explicit indexing or Traversal for "any" access.
type ArrayNotation int

const (
	// ArrayNotationInvalid is used when parsing did neither yield index nor
	// traversal notation
	ArrayNotationInvalid = ArrayNotation(-1)
	// ArrayNotationIndex is used when direct element access is requested
	ArrayNotationIndex = ArrayNotation(0)
	// ArrayNotationTraversal is used when any element access is requested
	ArrayNotationTraversal = ArrayNotation(1)
)

var (
	unescapeJSONPath = strings.NewReplacer("~1", "/", "~0", "~")
	escapeJSONPath   = strings.NewReplacer("/", "~1", "~", "~0")
)

var (
	// PathMetadata holds the common path to an object's metadata section
	PathMetadata = Path{"metadata"}

	// PathMetadataName holds the common path to an object's name
	PathMetadataName = Path{"metadata", "name"}

	// PathMetadataGenerateName holds the common path to an object's name prefix
	PathMetadataGenerateName = Path{"metadata", "generateName"}

	// PathMetadataNamespace holds the common path to an object's namespace
	PathMetadataNamespace = Path{"metadata", "namespace"}

	// PathLabels holds the common path to an object's label section
	PathLabels = Path{"metadata", "labels"}

	// PathAnnotations holds the common path to an object's annotation section
	PathAnnotations = Path{"metadata", "annotations"}

	// PathOwnerReference holds the common path to an object's owner section
	PathOwnerReference = Path{"metadata", "ownerReferences"}

	// PathOwnerReference holds the common path to an object-owner's kind
	PathOwnerReferenceKind = Path{"metadata", "ownerReferences", "kind"}

	// PathSpec holds the common path to an object's spec section
	PathSpec = Path{"spec"}
)

// NewPath creates a new path object by appending a key to the given path.
// Note that this function will always allocate new memory.
func NewPath(p Path, key ...string) Path {
	newPath := make(Path, len(p), len(p)+1)
	copy(newPath, p)
	return append(newPath, key...)
}

// ConcatPaths will create a new path object by concatenating both pathes.
// Note that this function will always allocate new memory.
func ConcatPaths(p1, p2 Path) Path {
	newPath := make(Path, len(p1), len(p1)+len(p2))
	copy(newPath, p1)
	return append(newPath, p2...)
}

// NewPathFromJQFormat accepts a JQ-style path and transforms it into a Path
// object. Field names can be quoted using single tick. Arrays need to use
// square-braces postfixes (Array[]). Empty braces translated to "all" (read)
// or "append" (write).
func NewPathFromJQFormat(jqPath string) Path {
	if len(jqPath) == 0 {
		return Path{}
	}

	path := Path{}
	lastSplitIdx := int(0)
	quoted := false

	// Function called every time an identifier has been parsed
	addElement := func(idx int, rn rune) {
		element := jqPath[lastSplitIdx:idx]
		if len(element) > 0 {
			path = append(path, element)
		} else if rn == ']' {
			path = append(path, "-")
		}

		lastSplitIdx = idx + 1
	}

	for idx, rn := range jqPath {
		switch rn {
		case '\'':
			if quoted {
				quoted = false
				addElement(idx, rn)
				continue
			}
			quoted = true
			lastSplitIdx = idx + 1

		case '.':
			if quoted {
				continue
			}
			addElement(idx, rn)

		case '[':
			if quoted {
				continue
			}
			addElement(idx, rn)

		case ']':
			if quoted {
				continue
			}
			addElement(idx, rn)
		}
	}

	addElement(len(jqPath), 0)
	return path
}

// NewPathFromJSONPathFormat accepts a JSON path and transforms it into a Path
// object.
// See https://jsonpatch.com/#json-pointer
func NewPathFromJSONPathFormat(jsonPath string) Path {
	if len(jsonPath) == 0 || jsonPath == "/" {
		return Path{}
	}

	path := strings.Split(jsonPath, "/")

	if len(path[0]) == 0 {
		path = path[1:]
	}

	for i, p := range path {
		if strings.ContainsRune(p, '~') {
			path[i] = unescapeJSONPath.Replace(p)
		}
	}

	return path
}

// ToJSONPath converts the path to a valid JSONPatch path, escaping special
// characters if needed.
// See https://jsonpatch.com/#json-pointer
func (p Path) ToJSONPath() string {
	var b strings.Builder

	if len(p) == 0 {
		return "/"
	}

	// Best effort assumption on size.
	// If a to-be-escaped character is used, the string build needs to allocate
	// additional memory. Otherwise this is a 1-alloc operation.
	capacity := 0
	for _, e := range p {
		capacity += len(e) + 1
	}
	b.Grow(capacity)

	for _, e := range p {
		b.WriteRune('/')
		b.WriteString(escapeJSONPath.Replace(e))
	}

	return b.String()
}

// SplitKey extracts the last element from the path and returns it as a separate
// key. If the last element denotes an array access, the access pattern (all or
// explicit index) is dropped and only the name is returned.
func (p Path) SplitKey() (Path, string) {
	if len(p) == 0 {
		return Path{}, ""
	}

	keyIdx := len(p) - 1
	key := p[keyIdx]
	for keyIdx > 0 && (key[0] == '-' || (key[0] >= '0' && key[0] <= '9')) {
		keyIdx--
		key = p[keyIdx]
	}

	return p[:keyIdx], key
}

// IsArray returns true if an element at a specific location is either a fully
// referenced array (name + index notation) or if it is just an index notation.
func (p Path) IsArray(idx int) (bool, ArrayNotation) {
	if len(p) == 0 {
		return false, ArrayNotationInvalid
	}
	if idx >= len(p) {
		return false, ArrayNotationInvalid
	}

	pathAtIdx := p[idx:]

	// Unnamed array
	if len(pathAtIdx) < 1 {
		return false, ArrayNotationInvalid
	}
	if notation := GetArrayNotation(pathAtIdx[0]); notation != ArrayNotationInvalid {
		return true, notation
	}

	// Named array
	if len(pathAtIdx) < 2 {
		return false, ArrayNotationInvalid
	}
	if notation := GetArrayNotation(pathAtIdx[1]); notation != ArrayNotationInvalid {
		return true, notation
	}

	// Not an array
	return false, ArrayNotationInvalid
}

// GetArrayNotation returns the notation type of an array index notation value
func GetArrayNotation(key string) ArrayNotation {
	switch {
	case len(key) == 0:
		return ArrayNotationInvalid
	case key[0] == '-':
		return ArrayNotationTraversal
	case key[0] >= '0' && key[0] <= '9':
		return ArrayNotationIndex
	}
	return ArrayNotationInvalid
}
