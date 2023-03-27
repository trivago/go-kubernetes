package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testNamespacedObjectJSON = `{
    "apiVersion": "v1",
    "kind": "ConfigMap",
    "metadata": {
      "name": "test",
      "namespace": "default",
      "annotations": {
        "foo": "bar",
        "foo/esc": "escaped"
      },
      "labels": {
        "foo": "bar",
        "foo/esc": "escaped"
      },
      "resourceVersion": "2026177",
      "uid": "5f5ec878-3922-4f1f-8ca2-0a878066d09a"
    },
    "data": {
      "dashboard.json": "{}"
    },
    "array": [
      {
        "search": "0",
        "nested": [{
          "search": "1"
        }]
      },{
        "search": "2"
      }
    ]
  }`

	testPodJSON = `{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
      "labels": {
        "app": "aclaus-dummy-22270"
      },
      "name": "aclaus-dummy-22270",
      "namespace": "affinity-controller"
    },
    "spec": {
      "affinity": {
        "nodeAffinity": {
          "requiredDuringSchedulingIgnoredDuringExecution": {
            "nodeSelectorTerms": [
              {
                "matchExpressions": [
                  {
                    "key": "pool",
                    "operator": "In",
                    "values": ["priority"]
                  }
                ]
              }
            ]
          }
        }
      },
      "tolerations": [
        {
          "effect": "NoSchedule",
          "key": "cloud.google.com/gke-spot",
          "operator": "Equal"
        }
      ]
    }
  }`
)

func TestNamespacedObjectFromRaw(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testNamespacedObjectJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)
	assert.Equal(t, "test", obj.GetName())
	assert.Equal(t, "default", obj.GetNamespace())
}

func TestNamespacedObjectRename(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testNamespacedObjectJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)
	assert.Equal(t, "test", obj.GetName())
	assert.Equal(t, "default", obj.GetNamespace())

	obj.SetName("foo")
	obj.SetNamespace("bar")

	assert.Equal(t, "foo", obj.GetName())
	assert.Equal(t, "bar", obj.GetNamespace())
}

func TestAnnotations(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testNamespacedObjectJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)
	assert.True(t, obj.HasAnnotations())

	a, aOk := obj.GetAnnotation("foo")
	assert.True(t, aOk)
	assert.Equal(t, "bar", a)

	assert.True(t, obj.IsAnnotationSetTo("foo", "bar"))
	assert.True(t, obj.IsAnnotationSetTo("foo", "BaR"))
	assert.False(t, obj.IsAnnotationSetTo("foo", "foo"))
	assert.False(t, obj.IsAnnotationSetTo("bar", "-"))

	assert.False(t, obj.IsAnnotationNotSetTo("foo", "bar"))
	assert.False(t, obj.IsAnnotationNotSetTo("foo", "BaR"))
	assert.True(t, obj.IsAnnotationNotSetTo("foo", "foo"))
	assert.True(t, obj.IsAnnotationNotSetTo("bar", "-"))

	b, bOk := obj.GetAnnotation("foo/esc")
	assert.True(t, bOk)
	assert.Equal(t, "escaped", b)

	obj.SetAnnotation("foo/esc", "changed")
	obj.SetAnnotation("new", "shiny")

	b, bOk = obj.GetAnnotation("foo/esc")
	assert.True(t, bOk)
	assert.Equal(t, "changed", b)

	n, nOk := obj.GetAnnotation("new")
	assert.True(t, nOk)
	assert.Equal(t, "shiny", n)
}

func TestLabels(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testNamespacedObjectJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)
	assert.True(t, obj.HasLabels())

	a, aOk := obj.GetLabel("foo")
	assert.True(t, aOk)
	assert.Equal(t, "bar", a)

	assert.True(t, obj.IsLabelSetTo("foo", "bar"))
	assert.True(t, obj.IsLabelSetTo("foo", "BaR"))
	assert.False(t, obj.IsLabelSetTo("foo", "foo"))
	assert.False(t, obj.IsLabelSetTo("bar", "-"))

	assert.False(t, obj.IsLabelNotSetTo("foo", "bar"))
	assert.False(t, obj.IsLabelNotSetTo("foo", "BaR"))
	assert.True(t, obj.IsLabelNotSetTo("foo", "foo"))
	assert.True(t, obj.IsLabelNotSetTo("bar", "-"))

	b, bOk := obj.GetLabel("foo/esc")
	assert.True(t, bOk)
	assert.Equal(t, "escaped", b)
}

func TestRemoveManagedFields(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testNamespacedObjectJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)
	obj.RemoveManagedFields()

	assert.False(t, obj.Has([]string{"metadata"}, "resourceVersion"))
	assert.False(t, obj.Has([]string{"metadata"}, "uid"))
}

func TestSet(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testNamespacedObjectJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)
	obj.RemoveManagedFields()

	typeMismatchOk := obj.Set([]string{"data", "dashboard.json"}, "sub", "{}")
	assert.False(t, typeMismatchOk)

	// Create field
	setOk := obj.Set([]string{"spec", "data"}, "test.json", "{}")
	assert.True(t, setOk)

	field, fieldOk := obj.GetString([]string{"spec", "data"}, "test.json")
	assert.True(t, fieldOk)
	assert.Equal(t, "{}", field)

	// Change array
	setOk = obj.Set([]string{"array[1]"}, "search", "3")
	assert.True(t, setOk)

	field, fieldOk = obj.GetString([]string{"array[1]"}, "search")
	assert.True(t, fieldOk)
	assert.Equal(t, "3", field)

	// Bulk change array
	setOk = obj.Set([]string{"array[]"}, "search", "4")
	assert.True(t, setOk)

	field, fieldOk = obj.GetString([]string{"array[0]"}, "search")
	assert.True(t, fieldOk)
	assert.Equal(t, "4", field)

	field, fieldOk = obj.GetString([]string{"array[1]"}, "search")
	assert.True(t, fieldOk)
	assert.Equal(t, "4", field)

	// Create annotation
	obj.SetAnnotation("test", "test")

	a, aOk := obj.GetAnnotation("test")
	assert.True(t, obj.Has([]string{"metadata", "annotations"}, "test"))
	assert.True(t, aOk)
	assert.Equal(t, "test", a)
}

func TestGet(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testNamespacedObjectJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)

	value := obj.Get([]string{"array[]"}, "search")
	assert.Equal(t, "0", value)

	value = obj.Get([]string{"array[0]"}, "search")
	assert.Equal(t, "0", value)

	value = obj.Get([]string{"array[1]"}, "search")
	assert.Equal(t, "2", value)

	value = obj.Get([]string{"array[2]"}, "search")
	assert.Nil(t, value)

	value = obj.Get([]string{"array"}, "search")
	assert.Nil(t, value)

	value = obj.Get([]string{"array[]", "nested[]"}, "search")
	assert.Equal(t, "1", value)
}

func TestFind(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testNamespacedObjectJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)

	// Test "any" search
	value := obj.Find([]string{"array[]"}, "search", nil)
	assert.Equal(t, 2, len(value))
	assert.Equal(t, []string{"array[0]", "search"}, value[0])
	assert.Equal(t, []string{"array[1]", "search"}, value[1])

	// Test "any" first search
	singleValue := obj.FindFirst([]string{"array[]"}, "search", nil)
	assert.NotEmpty(t, value)
	assert.Equal(t, []string{"array[0]", "search"}, singleValue)

	// Test specific search
	value = obj.Find([]string{"array[]"}, "search", "0")
	assert.Equal(t, 1, len(value))

	// Test nested "any" search
	value = obj.Find([]string{"array[]", "nested[]"}, "search", nil)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, []string{"array[0]", "nested[0]", "search"}, value[0])

	// Test out of bound
	value = obj.Find([]string{"array[]", "nested[1]"}, "search", nil)
	assert.Empty(t, value)

	// Test mismatch
	value = obj.Find([]string{"array[]"}, "search", "10")
	assert.Empty(t, value)
}

func TestSplitPathKey(t *testing.T) {
	path0 := []string{}
	path1 := []string{"foo"}
	path2 := []string{"foo", "bar"}
	path3 := []string{"foo", "bar", "baz"}

	p, k := SplitPathKey(path0)
	assert.Empty(t, p)
	assert.Empty(t, k)

	p, k = SplitPathKey(path1)
	assert.Empty(t, p)
	assert.Equal(t, "foo", k)

	p, k = SplitPathKey(path2)
	assert.Equal(t, []string{"foo"}, p)
	assert.Equal(t, "bar", k)

	p, k = SplitPathKey(path3)
	assert.Equal(t, []string{"foo", "bar"}, p)
	assert.Equal(t, "baz", k)
}

func TestComplexHash(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testNamespacedObjectJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)

	hash, err := obj.Hash()
	assert.NoError(t, err)
	assert.NotEqual(t, uint64(0), hash)

	hashStr, err := obj.HashStr()
	assert.NoError(t, err)

	// The following asserts that hashing stays stable between runs.
	// If the testNamespacedObjectJSON object is changed, a new hash will be
	// generated and this test fails.
	assert.Equal(t, "UhcMof5X3kM=", hashStr)
}

func TestHashChanges(t *testing.T) {
	obj := NamespacedObject(make(map[string]interface{}))

	hash1, err := obj.Hash()
	assert.NoError(t, err)

	obj.SetName("foo")
	hash2, err := obj.Hash()
	assert.NoError(t, err)
	assert.NotEqual(t, hash1, hash2)

	obj.SetAnnotation("bar", "foo")
	hash3, err := obj.Hash()
	assert.NoError(t, err)
	assert.NotEqual(t, hash2, hash3)

	obj.SetAnnotation("zaa", "moo")
	hash4, err := obj.Hash()
	assert.NoError(t, err)
	assert.NotEqual(t, hash3, hash4)

	obj.SetAnnotation("foo", "bar")
	hash5, err := obj.Hash()
	assert.NoError(t, err)
	assert.NotEqual(t, hash4, hash5)

	obj.Delete([]string{"metadata", "annotations"}, "foo")

	hash6, err := obj.Hash()
	assert.NoError(t, err)
	assert.Equal(t, hash4, hash6)
}

func TestPatchFixPatchPath(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testNamespacedObjectJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)

	var (
		path  []string
		value interface{}
	)

	path, value = obj.FixPatchPath([]string{"array[]", "foo", "newKey"}, "newValue")
	assert.Equal(t, []string{"array[]", "foo"}, path)
	assert.Equal(t, map[string]interface{}{
		"newKey": "newValue",
	}, value)

	// key/value pair does not exist
	path, value = obj.FixPatchPath([]string{"metadata", "field"}, "newValue")
	assert.Equal(t, []string{"metadata", "field"}, path)
	assert.Equal(t, "newValue", value)

	// array does not exist
	path, value = obj.FixPatchPath([]string{"metadata", "list[]"}, "newValue")
	assert.Equal(t, []string{"metadata", "list"}, path)
	assert.Equal(t, []interface{}{"newValue"}, value)

	// array element does not exist
	path, value = obj.FixPatchPath([]string{"array[3]", "foo"}, "newValue")
	assert.Equal(t, []string{"array[]", "foo"}, path)
	assert.Equal(t, "newValue", value)

	// top level array element does not exist
	path, value = obj.FixPatchPath([]string{"array[3]"}, "newValue")
	assert.Equal(t, []string{"array[]"}, path)
	assert.Equal(t, "newValue", value)

	// top level element does not exist
	path, value = obj.FixPatchPath([]string{"spec"}, "newValue")
	assert.Equal(t, []string{"spec"}, path)
	assert.Equal(t, "newValue", value)

	// nested key/value pair does not exist
	path, value = obj.FixPatchPath([]string{"metadata", "annotations", "newKey"}, "newValue")
	assert.Equal(t, []string{"metadata", "annotations", "newKey"}, path)
	assert.Equal(t, "newValue", value)

	// multiple nested arrays
	// TODO: wraps value in map[string]interface{} twice
	// path, value = obj.FixPatchPath([]string{"array[]", "first[]", "second[]", "key"}, "value")
	// assert.Equal(t, []string{"array[]", "first"}, path)
	// assert.Equal(t, map[string]interface{}{
	// 	"second": []interface{}{
	// 		map[string]interface{}{
	// 			"key": "value",
	// 		},
	// 	}}, value)

	// multiple nested key/value pairs
	// second to last is map
	path, value = obj.FixPatchPath([]string{"array[]", "first", "second", "key"}, "value")
	assert.Equal(t, []string{"array[]", "first"}, path)
	assert.Equal(t, map[string]interface{}{
		"second": map[string]interface{}{
			"key": "value",
		}}, value)

	// multiple nested arrays and key/value pairs
	// second to last is array
	path, value = obj.FixPatchPath([]string{"array[]", "first", "second[]", "key"}, "value")
	assert.Equal(t, []string{"array[]", "first"}, path)
	assert.Equal(t, map[string]interface{}{
		"second": []interface{}{
			map[string]interface{}{
				"key": "value",
			},
		}}, value)

	// Last key is array
	path, value = obj.FixPatchPath([]string{"metadata", "affinity", "nodeAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]"}, map[string]interface{}{"key": "value"})
	assert.Equal(t, []string{"metadata", "affinity"}, path)
	assert.Equal(t, map[string]interface{}{
		"nodeAffinity": map[string]interface{}{
			"preferredDuringSchedulingIgnoredDuringExecution": []interface{}{
				map[string]interface{}{
					"key": "value",
				},
			},
		},
	}, value)
}

func TestPodFixPatchPath(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testPodJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)

	var (
		path  []string
		value interface{}
	)

	newOptionalNodeAffinityPath := []string{"spec", "affinity", "nodeAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]"}
	affinityPatch := map[string]interface{}{
		"weight": 100,
		"preference": map[string]interface{}{
			"matchExpressions": []map[string]interface{}{
				{
					"key":      "test",
					"operator": "In",
					"values": []string{
						"true",
					},
				},
			},
		},
	}

	path, value = obj.FixPatchPath(newOptionalNodeAffinityPath, affinityPatch)
	assert.Equal(t, []string{"spec", "affinity", "nodeAffinity", "preferredDuringSchedulingIgnoredDuringExecution"}, path)
	assert.Equal(t, []interface{}{affinityPatch}, value)
}

func TestPodCases(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testPodJSON),
	}

	obj, err := NamespacedObjectFromRaw(&json)
	assert.NoError(t, err)

	// Check error case
	// path := []string{"spec", "affinity", "nodeAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "nodeSelectorTerms[]", "matchExpressions[]"}
	// foundPath := obj.FindFirst(path, "key", "pool")
	// assert.Equal(t, []string{"spec", "affinity", "nodeAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "nodeSelectorTerms[0]", "matchExpressions[0]", "key"}, foundPath)

	// Check error case
	affinityPatch := map[string]interface{}{
		"weight": 100,
		"preference": map[string]interface{}{
			"matchExpressions": []map[string]interface{}{
				{
					"key":      "cloud.google.com/gke-spot",
					"operator": "In",
					"values": []string{
						"true",
					},
				},
			},
		},
	}

	patchPath, _ := obj.FixPatchPath([]string{"spec", "affinity", "nodeAffinity", "preferredDuringSchedulingIgnoredDuringExecution[]"}, affinityPatch)
	assert.Equal(t, []string{"spec", "affinity", "nodeAffinity", "preferredDuringSchedulingIgnoredDuringExecution"}, patchPath)
}
