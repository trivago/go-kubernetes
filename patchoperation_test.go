package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeJSONPath(t *testing.T) {
	var (
		test1 = []string{"foo", "bar"}
		test2 = []string{"a/b"}
		test3 = []string{"a/b", "c"}
		test4 = []string{"a/b", "c/d"}
		test5 = []string{"a", "b[]", "c"}
		test6 = []string{"a", "b[1]", "c"}
	)

	assert.Equal(t, "/foo/bar", EscapeJSONPath(test1))
	assert.Equal(t, "/a~1b", EscapeJSONPath(test2))
	assert.Equal(t, "/a~1b/c", EscapeJSONPath(test3))
	assert.Equal(t, "/a~1b/c~1d", EscapeJSONPath(test4))
	assert.Equal(t, "/a/b/-/c", EscapeJSONPath(test5))
	assert.Equal(t, "/a/b/1/c", EscapeJSONPath(test6))
}
