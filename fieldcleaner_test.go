package kubernetes

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

const (
	testFieldCleanerJSON = `{
    "apiVersion": "v1",
    "kind": "Namespace",
    "metadata": {
        "creationTimestamp": "2022-09-26T09:54:21Z",
        "labels": {
						"app.kubernetes.io/managed-by": "Helm",
            "kubernetes.io/metadata.name": "test",
            "name": "test"
        },
        "name": "test",
        "resourceVersion": "29601",
        "uid": "333e29cc-24d0-496b-a376-d801676c86c5"
    },
    "spec": {
        "finalizers": [
            "kubernetes"
        ]
    },
    "status": {
        "phase": "Active"
    }
}`
)

func TestFieldCleaner(t *testing.T) {
	obj := NamespacedObject(make(map[string]interface{}))
	err := jsoniter.UnmarshalFromString(testFieldCleanerJSON, &obj)
	assert.NoError(t, err)

	KubernetesManagedFields.Clean(obj)

	assert.False(t, obj.Has([]string{"metadata"}, "uid"))
	assert.False(t, obj.Has([]string{"metadata"}, "resourceVersion"))
	assert.False(t, obj.Has([]string{"metadata"}, "creationTimestamp"))
	assert.False(t, obj.Has([]string{"spec", "finalizers"}, "resourceVersion"))
	assert.False(t, obj.Has([]string{}, "status"))
}
