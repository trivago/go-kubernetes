# Changelog

## [2.0.0](https://github.com/trivago/go-kubernetes/compare/v1.0.0...v2.0.0) (2023-04-26)


### ⚠ BREAKING CHANGES

* It's not required to split key and path anymore, it's just path now
* A unified `walk` function for searching / modifying the namedObject structure
* NamedObject functions now return errors
* Hash function now processes nested fields
* namedObject v2
* Rename namespacedObject to namedObject
* Remove `client.GetNamespacedResourceHandle`
* Make `client.ListAllObjects` return a slice of `NamedObject` and a wrapped error instead of `Unstructured` objects
* Rename `KubernetesManagedFields` to `ManagedFields` to avoid stuttering
* Cleanup API ([#2](https://github.com/trivago/go-kubernetes/issues/2))
* rename kubernetesManagedField to ManageFields
* rename namespacedobject to namedobject

### Features

* `SetLabel` added ([77535e8](https://github.com/trivago/go-kubernetes/commit/77535e835c5ab763bc2193fcfae7512eadd8b2b7))
* A unified `walk` function for searching / modifying the namedObject structure ([77535e8](https://github.com/trivago/go-kubernetes/commit/77535e835c5ab763bc2193fcfae7512eadd8b2b7))
* Add predefined variables for commonly used GVRs ([6895476](https://github.com/trivago/go-kubernetes/commit/6895476bd3da5cc0f1af5e11d091712c6d1730cc))
* Cleanup API ([#2](https://github.com/trivago/go-kubernetes/issues/2)) ([6895476](https://github.com/trivago/go-kubernetes/commit/6895476bd3da5cc0f1af5e11d091712c6d1730cc))
* Distinguish between `client.GetObject` and `client.GetNamespacedObject` ([6895476](https://github.com/trivago/go-kubernetes/commit/6895476bd3da5cc0f1af5e11d091712c6d1730cc))
* It's not required to split key and path anymore, it's just path now ([77535e8](https://github.com/trivago/go-kubernetes/commit/77535e835c5ab763bc2193fcfae7512eadd8b2b7))
* NamedObject functions now return errors ([77535e8](https://github.com/trivago/go-kubernetes/commit/77535e835c5ab763bc2193fcfae7512eadd8b2b7))
* namedObject v2 ([77535e8](https://github.com/trivago/go-kubernetes/commit/77535e835c5ab763bc2193fcfae7512eadd8b2b7))
* New `Path` object for handling paths, replacing `EscapeJSONPath` and `StringToPath` ([77535e8](https://github.com/trivago/go-kubernetes/commit/77535e835c5ab763bc2193fcfae7512eadd8b2b7))
* Remove `client.GetNamespacedResourceHandle` ([6895476](https://github.com/trivago/go-kubernetes/commit/6895476bd3da5cc0f1af5e11d091712c6d1730cc))
* rename namespacedobject to namedobject ([cca7587](https://github.com/trivago/go-kubernetes/commit/cca758715219f1fab1ee53a6746d2679ee2d1822))
* Rename namespacedObject to namedObject ([6895476](https://github.com/trivago/go-kubernetes/commit/6895476bd3da5cc0f1af5e11d091712c6d1730cc))


### Bug Fixes

* Hash function now processes nested fields ([77535e8](https://github.com/trivago/go-kubernetes/commit/77535e835c5ab763bc2193fcfae7512eadd8b2b7))
* Make `client.ListAllObjects` return a slice of `NamedObject` and a wrapped error instead of `Unstructured` objects ([6895476](https://github.com/trivago/go-kubernetes/commit/6895476bd3da5cc0f1af5e11d091712c6d1730cc))
* Rename `KubernetesManagedFields` to `ManagedFields` to avoid stuttering ([6895476](https://github.com/trivago/go-kubernetes/commit/6895476bd3da5cc0f1af5e11d091712c6d1730cc))
* rename kubernetesManagedField to ManageFields ([81e8273](https://github.com/trivago/go-kubernetes/commit/81e827300e2d176ff3f3c8bef76d9b00e6caff9c))


### Miscellaneous

* add coverage and verbose output to tests ([6895476](https://github.com/trivago/go-kubernetes/commit/6895476bd3da5cc0f1af5e11d091712c6d1730cc))
* Bump google-github-actions/release-please-action from 3.7.1 to 3.7.5 ([#1](https://github.com/trivago/go-kubernetes/issues/1)) ([bdb9664](https://github.com/trivago/go-kubernetes/commit/bdb96641e8accf7d0197852115990cc8c25b6242))
* remove unused variables and constants ([6895476](https://github.com/trivago/go-kubernetes/commit/6895476bd3da5cc0f1af5e11d091712c6d1730cc))
* renamed files to match contents ([6895476](https://github.com/trivago/go-kubernetes/commit/6895476bd3da5cc0f1af5e11d091712c6d1730cc))
