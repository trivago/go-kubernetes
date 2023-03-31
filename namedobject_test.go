package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	configMapJSON = `{
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
    }
  }`

	podJSON = `{
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

	testCasesJSON = `{
		"metadata": {
			"name": "testcase"
		},
		"a" : {
			"obj" : {
				"value": "value",
				"emptyArray": [],
				"array": ["a", "b"],
				"arrayInArray": [
					["a", "b"],
					["c", "d"]
				]
			},
			"array": [
				{
					"value": "value",
					"emptyArray": [],
					"array": ["a", "b"],
					"arrayInArray": [
						["a", "b"],
						["c", "d"]
					], "obj" : {
						"value": "value",
						"emptyArray": [],
						"array": ["a", "b"],
						"arrayInArray": [
							["a", "b"],
							["c", "d"]
						]
					}
				},
				{
					"value": "value2",
					"emptyArray": [],
					"array": ["a2", "b2"],
					"arrayInArray": [
						["a2", "b2"],
						["c2", "d2"]
					],
					"obj" : {
						"value": "value",
						"emptyArray": [],
						"array": ["a", "b"],
						"arrayInArray": [
							["a", "b"],
							["c", "d"]
						]
					}
				}
			]
		},
		"obj" : {
			"value": "value",
			"emptyArray": [],
			"array": ["a", "b"],
			"arrayInArray": [
				["a", "b"],
				["c", "d"]
			],
			"obj" : {
				"value": "value",
				"emptyArray": [],
				"array": ["a", "b"],
				"arrayInArray": [
					["a", "b"],
					["c", "d"]
				]
			}
		},
		"array": [
			{
				"value": "value",
				"emptyArray": [],
				"array": ["a", "b"],
				"arrayInArray": [
					["a", "b"],
					["c", "d"]
				],
				"obj" : {
					"value": "value",
					"emptyArray": [],
					"array": ["a", "b"],
					"arrayInArray": [
						["a", "b"],
						["c", "d"]
					]
				}
			},
			{
				"value": "value2",
				"emptyArray": [],
				"array": ["a2", "b2"],
				"arrayInArray": [
					["a2", "b2"],
					["c2", "d2"]
				],
				"obj" : {
					"value": "value",
					"emptyArray": [],
					"array": ["a", "b"],
					"arrayInArray": [
						["a", "b"],
						["c", "d"]
					]
				}
			}
		],
		"arrayInArray": [
			["a", "b"],
			["c", "d"]
		],
		"emptyArray": []
	}`
)

func TestNamedObjectFromRaw(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)
	assert.Equal(t, "test", obj.GetName())
	assert.Equal(t, "default", obj.GetNamespace())
}

func TestNamedObjectRename(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
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
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)
	assert.True(t, obj.HasAnnotations())

	a, err := obj.GetAnnotation("foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", a)

	assert.True(t, obj.IsAnnotationSetTo("foo", "bar"))
	assert.True(t, obj.IsAnnotationSetTo("foo", "BaR"))
	assert.False(t, obj.IsAnnotationSetTo("foo", "foo"))
	assert.False(t, obj.IsAnnotationSetTo("bar", "-"))

	assert.False(t, obj.IsAnnotationNotSetTo("foo", "bar"))
	assert.False(t, obj.IsAnnotationNotSetTo("foo", "BaR"))
	assert.True(t, obj.IsAnnotationNotSetTo("foo", "foo"))
	assert.True(t, obj.IsAnnotationNotSetTo("bar", "-"))

	b, err := obj.GetAnnotation("foo/esc")
	assert.NoError(t, err)
	assert.Equal(t, "escaped", b)

	obj.SetAnnotation("foo/esc", "changed")
	obj.SetAnnotation("new", "shiny")

	b, err = obj.GetAnnotation("foo/esc")
	assert.NoError(t, err)
	assert.Equal(t, "changed", b)

	n, err := obj.GetAnnotation("new")
	assert.NoError(t, err)
	assert.Equal(t, "shiny", n)
}

/*
func TestLabels(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
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
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)
	obj.RemoveManagedFields()

	assert.False(t, obj.Has([]string{"metadata"}, "resourceVersion"))
	assert.False(t, obj.Has([]string{"metadata"}, "uid"))
}

func TestSet(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
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
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
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
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
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
}*/

func TestComplexHash(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)

	hash, err := obj.Hash()
	assert.NoError(t, err)
	assert.NotEqual(t, uint64(0), hash)

	hashStr, err := obj.HashStr()
	assert.NoError(t, err)

	// The following asserts that hashing stays stable between runs.
	// If the testNamedObjectJSON object is changed, a new hash will be
	// generated and this test fails.
	assert.Equal(t, "iuFW+tRydu8=", hashStr)
}

/*
func TestHashChanges(t *testing.T) {
	obj := NamedObject(make(map[string]interface{}))

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

	obj.Delete(Path{"metadata", "annotations", "foo"})

	hash6, err := obj.Hash()
	assert.NoError(t, err)
	assert.Equal(t, hash4, hash6)
}


func TestPodFixPatchPath(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(podJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
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
		Raw: []byte(podJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
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
*/

func TestWalk(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testCasesJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)

	var v interface{}

	v, err = obj.Walk(NewPathFromJQFormat("a.obj.value"), WalkArgs{})
	assert.NoError(t, err)
	assert.Equal(t, "value", v)

	v, err = obj.Walk(NewPathFromJQFormat("a.obj[].value"), WalkArgs{})
	assert.NotNil(t, err)

	v, err = obj.Walk(NewPathFromJQFormat("a.obj[0].value"), WalkArgs{})
	assert.NotNil(t, err)

	v, err = obj.Walk(NewPathFromJQFormat("a.obj.array"), WalkArgs{})
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"a", "b"}, v)

	v, err = obj.Walk(NewPathFromJQFormat("a.obj.array.foo"), WalkArgs{})
	assert.NotNil(t, err)

	v, err = obj.Walk(NewPathFromJQFormat("a.obj.array[]"), WalkArgs{})
	assert.NoError(t, err)
	assert.Equal(t, "a", v)

	v, err = obj.Walk(NewPathFromJQFormat("a.obj.array[1]"), WalkArgs{})
	assert.NoError(t, err)
	assert.Equal(t, "b", v)

	v, err = obj.Walk(NewPathFromJQFormat("a.array[].obj.value"), WalkArgs{})
	assert.NoError(t, err)
	assert.Equal(t, "value", v)

	v, err = obj.Walk(NewPathFromJQFormat("a.array[].obj.value"), WalkArgs{MatchAll: true})
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"value", "value"}, v)

	v, err = obj.Walk(NewPathFromJQFormat("a.array[0].obj.value"), WalkArgs{})
	assert.NoError(t, err)
	assert.Equal(t, "value", v)

	v, err = obj.Walk(NewPathFromJQFormat("a.array[0].value"), WalkArgs{})
	assert.NoError(t, err)
	assert.Equal(t, "value", v)
}

func TestGeneratePatch(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testCasesJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)

	var (
		path  Path
		value interface{}
	)

	// Complete path does not exist
	path, value, err = obj.GeneratePatch(Path{"spec", "field"}, "newValue")
	assert.NoError(t, err)
	assert.Equal(t, Path{"spec"}, path)
	assert.Equal(t, map[string]interface{}{
		"field": "newValue",
	}, value)

	// Root level element does not exist
	path, value, err = obj.GeneratePatch(Path{"kind"}, "test")
	assert.NoError(t, err)
	assert.Equal(t, Path{"kind"}, path)
	assert.Equal(t, "test", value)

	// Last key does not exist
	path, value, err = obj.GeneratePatch(Path{"obj", "test"}, "value")
	assert.NoError(t, err)
	assert.Equal(t, Path{"obj", "test"}, path)
	assert.Equal(t, "value", value)

	// Last key array does not exist
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("a.test[]"), "value")
	assert.NoError(t, err)
	assert.Equal(t, Path{"a", "test"}, path)
	assert.Equal(t, []interface{}{"value"}, value)

	// Append to array
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("a.obj.array[]"), "c")
	assert.NoError(t, err)
	assert.Equal(t, Path{"a", "obj", "array", "-"}, path)
	assert.Equal(t, "c", value)

	// Array requested, map found
	// Should yield an error as array traversal indicator cannot be used
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("a.obj[].value"), "value")
	assert.NotNil(t, err)

	// Array requested, map found
	// Should yield an error as array traversal indicator cannot be used
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("a.obj[]"), "value")
	assert.NotNil(t, err)

	// Map requested, array found
	// Should yield an error as array traversal indicator is missing
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("a.obj.array.key"), "value")
	assert.NotNil(t, err)

	// Array overwrite
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("a.obj.array"), "value")
	assert.NoError(t, err)
	assert.Equal(t, Path{"a", "obj", "array"}, path)
	assert.Equal(t, "value", value)

	// Key does not exist in existing, nested object
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("array[].obj.test"), "newValue")
	assert.NoError(t, err)
	assert.Equal(t, Path{"array", "-"}, path)
	assert.Equal(t, map[string]interface{}{
		"obj": map[string]interface{}{
			"test": "newValue",
		},
	}, value)

	// Array - object - array
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("a.array[].array[0]"), "newValue")
	assert.NoError(t, err)
	assert.Equal(t, Path{"a", "array", "0", "array", "0"}, path)
	assert.Equal(t, "newValue", value)

	// Array - array index
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("arrayInArray[][0]"), "newValue")
	assert.NoError(t, err)
	assert.Equal(t, Path{"arrayInArray", "0", "0"}, path)
	assert.Equal(t, "newValue", value)

	// Array - array append
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("emptyArray[][]"), "newValue")
	assert.NoError(t, err)
	assert.Equal(t, Path{"emptyArray", "-"}, path)
	assert.Equal(t, []interface{}{"newValue"}, value)

	// New Array - array
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("newArray[][]"), "newValue")
	assert.NoError(t, err)
	assert.Equal(t, Path{"newArray"}, path)
	assert.Equal(t, []interface{}{[]interface{}{"newValue"}}, value)

	// Key exists in exsiting, nested object
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("array[].obj.value"), "newValue")
	assert.NoError(t, err)
	assert.Equal(t, Path{"array", "0", "obj", "value"}, path)
	assert.Equal(t, "newValue", value)

	// Create new array in new object
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("a.newObj.newArray[].key"), "newValue")
	assert.NoError(t, err)
	assert.Equal(t, Path{"a", "newObj"}, path)
	assert.Equal(t, map[string]interface{}{
		"newArray": []interface{}{
			map[string]interface{}{
				"key": "newValue",
			},
		},
	}, value)

	// Create new array
	path, value, err = obj.GeneratePatch(NewPathFromJQFormat("a.newArray[].key"), "newValue")
	assert.NoError(t, err)
	assert.Equal(t, Path{"a", "newArray"}, path)
	assert.Equal(t, []interface{}{
		map[string]interface{}{
			"key": "newValue",
		},
	}, value)
}
