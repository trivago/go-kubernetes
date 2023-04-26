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
	obj := NamedObject(make(map[string]interface{}))
	err := jsoniter.UnmarshalFromString(testFieldCleanerJSON, &obj)
	assert.NoError(t, err)

	ManagedFields.Clean(obj)

	assert.False(t, obj.Has(NewPath(PathMetadata, "uid")))
	assert.False(t, obj.Has(NewPath(PathMetadata, "resourceVersion")))
	assert.False(t, obj.Has(NewPath(PathMetadata, "creationTimestamp")))
	assert.False(t, obj.Has(NewPath(PathSpec, "finalizers", "resourceVersion")))
	assert.False(t, obj.Has(Path{"status"}))
}
