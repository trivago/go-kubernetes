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
				"array": ["a", "b", "c"],
				"arrayInArray": [
					["a", "b", "c"],
					["c", "d"]
				]
			},
			"array": [
				{
					"value": "value",
					"emptyArray": [],
					"array": ["a", "b", "c"],
					"arrayInArray": [
						["a", "b", "c"],
						["c", "d"]
					], "obj" : {
						"value": "value",
						"emptyArray": [],
						"array": ["a", "b", "c"],
						"arrayInArray": [
							["a", "b", "c"],
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
						"array": ["a", "b", "c"],
						"arrayInArray": [
							["a", "b", "c"],
							["c", "d"]
						]
					}
				}
			]
		},
		"obj" : {
			"value": "value",
			"emptyArray": [],
			"array": ["a", "b", "c"],
			"arrayInArray": [
				["a", "b", "c"],
				["c", "d"]
			],
			"obj" : {
				"value": "value",
				"emptyArray": [],
				"array": ["a", "b", "c"],
				"arrayInArray": [
					["a", "b", "c"],
					["c", "d"]
				]
			}
		},
		"array": [
			{
				"value": "value",
				"emptyArray": [],
				"array": ["a", "b", "c"],
				"arrayInArray": [
					["a", "b", "c"],
					["c", "d"]
				],
				"obj" : {
					"value": "value",
					"emptyArray": [],
					"array": ["a", "b", "c"],
					"arrayInArray": [
						["a", "b", "c"],
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
					"array": ["a", "b", "c"],
					"arrayInArray": [
						["a", "b", "c"],
						["c", "d"]
					]
				}
			}
		],
		"arrayInArray": [
			["a", "b", "c"],
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

	err = obj.SetName("foo")
	assert.NoError(t, err)
	err = obj.SetNamespace("bar")
	assert.NoError(t, err)

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

	err = obj.SetAnnotation("foo/esc", "changed")
	assert.NoError(t, err)
	err = obj.SetAnnotation("new", "shiny")
	assert.NoError(t, err)

	b, err = obj.GetAnnotation("foo/esc")
	assert.NoError(t, err)
	assert.Equal(t, "changed", b)

	n, err := obj.GetAnnotation("new")
	assert.NoError(t, err)
	assert.Equal(t, "shiny", n)
}

func TestLabels(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)
	assert.True(t, obj.HasLabels())

	a, err := obj.GetLabel("foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", a)

	assert.True(t, obj.IsLabelSetTo("foo", "bar"))
	assert.True(t, obj.IsLabelSetTo("foo", "BaR"))
	assert.False(t, obj.IsLabelSetTo("foo", "foo"))
	assert.False(t, obj.IsLabelSetTo("bar", "-"))

	assert.False(t, obj.IsLabelNotSetTo("foo", "bar"))
	assert.False(t, obj.IsLabelNotSetTo("foo", "BaR"))
	assert.True(t, obj.IsLabelNotSetTo("foo", "foo"))
	assert.True(t, obj.IsLabelNotSetTo("bar", "-"))

	b, err := obj.GetLabel("foo/esc")
	assert.NoError(t, err)
	assert.Equal(t, "escaped", b)
}

func TestRemoveManagedFields(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(configMapJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)
	obj.RemoveManagedFields()

	assert.False(t, obj.Has(Path{"metadata", "resourceVersion"}))
	assert.False(t, obj.Has(Path{"metadata", "uid"}))
}

func TestSet(t *testing.T) {
	obj := NewNamedObject("test")

	var (
		value interface{}
		err   error
	)

	// Create top-level field
	err = obj.Set(NewPathFromJQFormat("testRoot"), "newValue")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("testRoot"))
	assert.NoError(t, err)
	assert.Equal(t, "newValue", value)

	// Change top-level field
	err = obj.Set(NewPathFromJQFormat("testRoot"), "changed")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("testRoot"))
	assert.NoError(t, err)
	assert.Equal(t, "changed", value)

	// Create top-level array
	err = obj.Set(NewPathFromJQFormat("testRootArray[]"), "newValue")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("testRootArray[]"))
	assert.NoError(t, err)
	assert.Equal(t, "newValue", value)

	// Change top-level array
	err = obj.Set(NewPathFromJQFormat("testRootArray[0]"), "changed")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("testRootArray[0]"))
	assert.NoError(t, err)
	assert.Equal(t, "changed", value)

	// Append to top-level array
	err = obj.Set(NewPathFromJQFormat("testRootArray[]"), "append")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("testRootArray[1]"))
	assert.NoError(t, err)
	assert.Equal(t, "append", value)

	// Create new section
	err = obj.Set(NewPathFromJQFormat("new1.test"), "newValue")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("new1.test"))
	assert.NoError(t, err)
	assert.Equal(t, "newValue", value)

	// Create new field in existing section
	err = obj.Set(NewPathFromJQFormat("new1.test2"), "newValue")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("new1.test2"))
	assert.NoError(t, err)
	assert.Equal(t, "newValue", value)

	// Change field in existing section
	err = obj.Set(NewPathFromJQFormat("new1.test2"), "changed")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("new1.test2"))
	assert.NoError(t, err)
	assert.Equal(t, "changed", value)

	// Create new array in existing section
	err = obj.Set(NewPathFromJQFormat("new1.test3[]"), "newValue")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("new1.test3[]"))
	assert.NoError(t, err)
	assert.Equal(t, "newValue", value)

	// Change array in existing section
	err = obj.Set(NewPathFromJQFormat("new1.test3[0]"), "changed")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("new1.test3[0]"))
	assert.NoError(t, err)
	assert.Equal(t, "changed", value)

	// Append to array in existing section
	err = obj.Set(NewPathFromJQFormat("new1.test3[]"), "append")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("new1.test3[1]"))
	assert.NoError(t, err)
	assert.Equal(t, "append", value)

	// Create new hierachy
	err = obj.Set(NewPathFromJQFormat("new2.test.test"), "newValue")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("new2.test.test"))
	assert.NoError(t, err)
	assert.Equal(t, "newValue", value)

	// Create new array hiearchy
	err = obj.Set(NewPathFromJQFormat("newArray[].newArray[].newArray[]"), "newValue")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("newArray[].newArray[].newArray[]"))
	assert.NoError(t, err)
	assert.Equal(t, "newValue", value)
	value, err = obj.Get(NewPathFromJQFormat("newArray[].newArray[]"))
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"newArray": []interface{}{"newValue"},
	}, value)

	// Create new multi-array
	err = obj.Set(NewPathFromJQFormat("new3[][]"), "newValue")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("new3[][]"))
	assert.NoError(t, err)
	assert.Equal(t, "newValue", value)

	// change new multi-array
	err = obj.Set(NewPathFromJQFormat("new3[0][0]"), "changed")
	assert.NoError(t, err)
	value, err = obj.GetString(NewPathFromJQFormat("new3[0][0]"))
	assert.NoError(t, err)
	assert.Equal(t, "changed", value)

	expectedJson := `{
		"metadata": {
			"name":"test"
		},
		"testRoot":"changed",
		"testRootArray":["changed","append"],
		"new1":{
			"test":"newValue",
			"test2":"changed",
			"test3":["changed","append"]
		},
		"new2":{
			"test":{
				"test":"newValue"
			}
		},
		"new3":[["changed"]],
		"newArray":[{
			"newArray":[{
				"newArray":["newValue"]
			}]
		}]
	}`

	json := runtime.RawExtension{
		Raw: []byte(expectedJson),
	}

	expectedObj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)
	assert.Equal(t, expectedObj, obj)
}

func TestDelete(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testCasesJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)

	// Field
	assert.True(t, obj.Has(NewPathFromJQFormat("metadata.name")))
	err = obj.Delete(NewPathFromJQFormat("metadata.name"))
	assert.NoError(t, err)
	assert.False(t, obj.Has(NewPathFromJQFormat("metadata.name")))
	assert.True(t, obj.Has(NewPathFromJQFormat("metadata")))

	// Array
	path := NewPathFromJQFormat("a.obj.array")
	assert.True(t, obj.Has(path))

	err = obj.Delete(NewPathFromJQFormat("a.obj.array[1]"))
	assert.NoError(t, err)
	assert.True(t, obj.Has(path))

	value, err := obj.Get(path)
	assert.NoError(t, err)
	assert.True(t, obj.Has(path))
	assert.Equal(t, 2, len(value.([]interface{})))

	err = obj.Delete(path)
	assert.NoError(t, err)
	assert.False(t, obj.Has(path))

	// Array-in-array tests
	path = NewPathFromJQFormat("a.obj.arrayInArray")
	path0 := NewPathFromJQFormat("a.obj.arrayInArray[0]")
	path1 := NewPathFromJQFormat("a.obj.arrayInArray[1]")
	assert.True(t, obj.Has(path))

	// Delete element in nested array
	err = obj.Delete(NewPathFromJQFormat("a.obj.arrayInArray[0][1]"))
	assert.NoError(t, err)
	assert.True(t, obj.Has(path))

	value, err = obj.Get(path0)
	assert.NoError(t, err)
	assert.True(t, obj.Has(path))
	assert.True(t, obj.Has(path0))
	assert.True(t, obj.Has(path1))
	assert.Equal(t, 2, len(value.([]interface{})))

	// Delete nested array
	err = obj.Delete(path0)
	assert.NoError(t, err)
	assert.True(t, obj.Has(path))
	assert.True(t, obj.Has(path0))
	assert.False(t, obj.Has(path1))

	// Object
	path = NewPathFromJQFormat("a.obj")
	assert.True(t, obj.Has(path))

	err = obj.Delete(path)
	assert.NoError(t, err)
	assert.False(t, obj.Has(path))
}

func TestGet(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(testCasesJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)

	value, err := obj.Get(NewPathFromJQFormat("a.obj.value"))
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	value, err = obj.Get(NewPathFromJQFormat("a.obj.array"))
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"a", "b", "c"}, value)

	value, err = obj.Get(NewPathFromJQFormat("a.obj.arrayInArray[0]"))
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"a", "b", "c"}, value)

	value, err = obj.Get(NewPathFromJQFormat("a.obj.arrayInArray[0][1]"))
	assert.NoError(t, err)
	assert.Equal(t, "b", value)

	value, err = obj.Get(NewPathFromJQFormat("a.array[0].value"))
	assert.NoError(t, err)
	assert.Equal(t, "value", value)
}

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

func TestHashChanges(t *testing.T) {
	obj := NamedObject(make(map[string]interface{}))

	hash1, err := obj.Hash()
	assert.NoError(t, err)

	err = obj.SetName("foo")
	assert.NoError(t, err)
	hash2, err := obj.Hash()
	assert.NoError(t, err)
	assert.NotEqual(t, hash1, hash2)

	err = obj.SetAnnotation("bar", "foo")
	assert.NoError(t, err)
	hash3, err := obj.Hash()
	assert.NoError(t, err)
	assert.NotEqual(t, hash2, hash3)

	err = obj.SetAnnotation("zaa", "moo")
	assert.NoError(t, err)
	hash4, err := obj.Hash()
	assert.NoError(t, err)
	assert.NotEqual(t, hash3, hash4)

	err = obj.SetAnnotation("foo", "bar")
	assert.NoError(t, err)
	hash5, err := obj.Hash()
	assert.NoError(t, err)
	assert.NotEqual(t, hash4, hash5)

	err = obj.Delete(Path{"metadata", "annotations", "foo"})
	assert.NoError(t, err)

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

	newOptionalNodeAffinityPath := Path{"spec", "affinity", "nodeAffinity", "preferredDuringSchedulingIgnoredDuringExecution", "-"}
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

	path, value, err := obj.GeneratePatch(newOptionalNodeAffinityPath, affinityPatch)
	assert.NoError(t, err)
	assert.Equal(t, Path{"spec", "affinity", "nodeAffinity", "preferredDuringSchedulingIgnoredDuringExecution"}, path)
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

	patchPath, _, err := obj.GeneratePatch(Path{"spec", "affinity", "nodeAffinity", "preferredDuringSchedulingIgnoredDuringExecution", "-"}, affinityPatch)
	assert.NoError(t, err)
	assert.Equal(t, Path{"spec", "affinity", "nodeAffinity", "preferredDuringSchedulingIgnoredDuringExecution"}, patchPath)
}

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

	_, err = obj.Walk(NewPathFromJQFormat("a.obj[].value"), WalkArgs{})
	assert.NotNil(t, err)

	_, err = obj.Walk(NewPathFromJQFormat("a.obj[0].value"), WalkArgs{})
	assert.NotNil(t, err)

	v, err = obj.Walk(NewPathFromJQFormat("a.obj.array"), WalkArgs{})
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"a", "b", "c"}, v)

	_, err = obj.Walk(NewPathFromJQFormat("a.obj.array.foo"), WalkArgs{})
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
	_, _, err = obj.GeneratePatch(NewPathFromJQFormat("a.obj[].value"), "value")
	assert.NotNil(t, err)

	// Array requested, map found
	// Should yield an error as array traversal indicator cannot be used
	_, _, err = obj.GeneratePatch(NewPathFromJQFormat("a.obj[]"), "value")
	assert.NotNil(t, err)

	// Map requested, array found
	// Should yield an error as array traversal indicator is missing
	_, _, err = obj.GeneratePatch(NewPathFromJQFormat("a.obj.array.key"), "value")
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
