package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	jqPathTests = map[string]Path{
		"":           {},
		"a":          {"a"},
		"'a'":        {"a"},
		"a[]":        {"a", "-"},
		"'a[]'":      {"a[]"},
		"a[1]":       {"a", "1"},
		"'a[1]'":     {"a[1]"},
		"a.b":        {"a", "b"},
		"a.'b'":      {"a", "b"},
		"a.b[]":      {"a", "b", "-"},
		"a.'b[]'":    {"a", "b[]"},
		"a.b[1]":     {"a", "b", "1"},
		"a.'b[1]'":   {"a", "b[1]"},
		"a.b.c":      {"a", "b", "c"},
		"a.'b'.c":    {"a", "b", "c"},
		"a.b[].c":    {"a", "b", "-", "c"},
		"a.'b[]'.c":  {"a", "b[]", "c"},
		"a.b[1].c":   {"a", "b", "1", "c"},
		"a.'b[1]'.c": {"a", "b[1]", "c"},
		"a.'b.c'":    {"a", "b.c"},
		"a.'b.c'[]":  {"a", "b.c", "-"},
		"a.'b.c'[1]": {"a", "b.c", "1"},
		"a.'b.c[]'":  {"a", "b.c[]"},
		"a.'b.c[1]'": {"a", "b.c[1]"},
	}

	jsonPathTests = map[string]Path{
		"/":            {},
		"/a":           {"a"},
		"/a/-":         {"a", "-"},
		"/a/1":         {"a", "1"},
		"/a/b":         {"a", "b"},
		"/a/b/c":       {"a", "b", "c"},
		"/a/b/-/c":     {"a", "b", "-", "c"},
		"/a/b/1/c":     {"a", "b", "1", "c"},
		"/a/b~1c":      {"a", "b/c"},
		"/a/b~1c/-":    {"a", "b/c", "-"},
		"/a/b~1c/1":    {"a", "b/c", "1"},
		"/a/b~1c/d":    {"a", "b/c", "d"},
		"/a/b~0c":      {"a", "b~c"},
		"/a/b~0c/-":    {"a", "b~c", "-"},
		"/a/b~0c/1":    {"a", "b~c", "1"},
		"/a/b~0c/d":    {"a", "b~c", "d"},
		"/a/b~0c~1d":   {"a", "b~c/d"},
		"/a/b~0c~1d/-": {"a", "b~c/d", "-"},
		"/a/b~0c~1d/1": {"a", "b~c/d", "1"},
		"/a/b~0c~1d/e": {"a", "b~c/d", "e"},
	}
)

func TestPathFromJQ(t *testing.T) {
	for s, p := range jqPathTests {
		assert.Equalf(t, p, NewPathFromJQFormat(s), "%s", s)
	}
}

func TestPathFromJSONPath(t *testing.T) {
	for s, p := range jsonPathTests {
		assert.Equal(t, p, NewPathFromJSONPathFormat(s), "%s", s)
	}
}

func TestToJSONPath(t *testing.T) {
	for s, p := range jsonPathTests {
		assert.Equal(t, s, p.ToJSONPath())
	}
}

func TestSplitKey(t *testing.T) {
	var (
		p Path
		k string
	)

	p, k = Path{}.SplitKey()
	assert.Equal(t, "", k)
	assert.Equal(t, Path{}, p)

	p, k = Path{"a"}.SplitKey()
	assert.Equal(t, "a", k)
	assert.Equal(t, Path{}, p)

	p, k = Path{"a", "b"}.SplitKey()
	assert.Equal(t, "b", k)
	assert.Equal(t, Path{"a"}, p)

	p, k = Path{"a", "b", "-"}.SplitKey()
	assert.Equal(t, "b", k)
	assert.Equal(t, Path{"a"}, p)

	p, k = Path{"a", "b", "1"}.SplitKey()
	assert.Equal(t, "b", k)
	assert.Equal(t, Path{"a"}, p)

	p, k = Path{"a", "b", "-", "-"}.SplitKey()
	assert.Equal(t, "b", k)
	assert.Equal(t, Path{"a"}, p)

	p, k = Path{"a", "-", "b", "-"}.SplitKey()
	assert.Equal(t, "b", k)
	assert.Equal(t, Path{"a", "-"}, p)
}

func TestNewPath(t *testing.T) {
	original := Path{"a", "b"}
	new := NewPath(original, "c")

	assert.Equal(t, Path{"a", "b", "c"}, new)

	original[0] = "x"
	assert.Equal(t, Path{"a", "b", "c"}, new)
}

func TestConcatPath(t *testing.T) {
	original1 := Path{"a", "b"}
	original2 := Path{"c", "d"}
	new := ConcatPaths(original1, original2)

	assert.Equal(t, Path{"a", "b", "c", "d"}, new)

	original1[0] = "1"
	original2[0] = "2"
	assert.Equal(t, Path{"a", "b", "c", "d"}, new)
}

func TestIsArray(t *testing.T) {
	testCase := Path{"a", "b", "-", "c", "9"}

	var (
		isArray  bool
		notation ArrayNotation
	)

	isArray, notation = testCase.IsArray(0)
	assert.False(t, isArray)
	assert.Equal(t, ArrayNotationInvalid, notation)

	isArray, notation = testCase.IsArray(1)
	assert.True(t, isArray)
	assert.Equal(t, ArrayNotationTraversal, notation)

	isArray, notation = testCase.IsArray(2)
	assert.True(t, isArray)
	assert.Equal(t, ArrayNotationTraversal, notation)

	isArray, notation = testCase.IsArray(3)
	assert.True(t, isArray)
	assert.Equal(t, ArrayNotationIndex, notation)

	isArray, notation = testCase.IsArray(4)
	assert.True(t, isArray)
	assert.Equal(t, ArrayNotationIndex, notation)

	isArray, notation = testCase.IsArray(5)
	assert.False(t, isArray)
	assert.Equal(t, ArrayNotationInvalid, notation)
}
