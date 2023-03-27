package kubernetes

type FieldCleaner struct {
	fields []string
	nested map[string]FieldCleaner
}

var KubernetesManagedFields = FieldCleaner{
	nested: map[string]FieldCleaner{
		"metadata": {
			fields: []string{
				"managedFields",
				"creationTimestamp",
				"generation",
				"resourceVersion",
				"uid",
				"finalizers",
			},
			nested: map[string]FieldCleaner{
				"labels": {
					fields: []string{
						"app.kubernetes.io/managed-by",
					},
				},
				"annotations": {
					fields: []string{
						"deployment.kubernetes.io/revision",
						"kubectl.kubernetes.io/last-applied-configuration",
					},
				},
			},
		},
		"status": {},
	},
}

func (f FieldCleaner) isSingleKey() bool {
	return len(f.fields) == 0 && len(f.nested) == 0
}

// Remove fields from an existing object
func (f FieldCleaner) Clean(obj map[string]interface{}) map[string]interface{} {
	for _, key := range f.fields {
		delete(obj, key)
	}

	for key, cleaner := range f.nested {
		if cleaner.isSingleKey() {
			delete(obj, key)
			continue
		}
		if subTree, ok := obj[key]; ok {
			obj[key] = cleaner.Clean(subTree.(map[string]interface{}))
		}
	}

	return obj
}
