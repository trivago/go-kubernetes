# Changelog

## [3.2.2](https://github.com/trivago/go-kubernetes/compare/v3.2.1...v3.2.2) (2025-02-10)


### Bug Fixes

* allow multiple audiences to be given ([80c7830](https://github.com/trivago/go-kubernetes/commit/80c7830eb7c115b464417733bb77b3af50985c24))

## [3.2.1](https://github.com/trivago/go-kubernetes/compare/v3.2.0...v3.2.1) (2025-01-10)


### Bug Fixes

* lowercase object kind for type check in GetServiceAccountToken ([5d6d747](https://github.com/trivago/go-kubernetes/commit/5d6d7471665ca85562ddfcd60a4d39056bbe3de6))

## [3.2.0](https://github.com/trivago/go-kubernetes/compare/v3.1.0...v3.2.0) (2025-01-09)


### Features

* add GetKind, GetVersion and GetUID to NamedObjects ([8be5cfb](https://github.com/trivago/go-kubernetes/commit/8be5cfb9a75c753182715e83105080e3bb9ff5be))
* allow service account tokens to be bound to a pod ([9c249d1](https://github.com/trivago/go-kubernetes/commit/9c249d16da7cd3f316a0d1af036e3bd690390cda))

## [3.1.0](https://github.com/trivago/go-kubernetes/compare/v3.0.0...v3.1.0) (2025-01-08)


### Features

* add GetServiceAccountToken ([2176842](https://github.com/trivago/go-kubernetes/commit/2176842086a8d854a9d6bbad428d849851c10738))

## [3.0.0](https://github.com/trivago/go-kubernetes/compare/v2.5.2...v3.0.0) (2025-01-07)


### ⚠ BREAKING CHANGES

* support field selectors in queries

### Features

* support field selectors in queries ([face352](https://github.com/trivago/go-kubernetes/commit/face352ccb4d61207e7540cebe9903e00e2ef09e))

## [2.5.2](https://github.com/trivago/go-kubernetes/compare/v2.5.1...v2.5.2) (2024-07-11)


### Bug Fixes

* remove output of apply/patch result ([57ce0d4](https://github.com/trivago/go-kubernetes/commit/57ce0d4ec3a76abd943c8ccf4645b092f2a44bb3))

## [2.5.1](https://github.com/trivago/go-kubernetes/compare/v2.5.0...v2.5.1) (2024-07-10)


### Bug Fixes

* add result of patch and apply command to debug out ([3c29f5c](https://github.com/trivago/go-kubernetes/commit/3c29f5c3102654d4f46a018a2c74a669674d5275))

## [2.5.0](https://github.com/trivago/go-kubernetes/compare/v2.4.0...v2.5.0) (2024-01-11)


### Features

* add patch function to client ([#30](https://github.com/trivago/go-kubernetes/issues/30)) ([3e9fc2e](https://github.com/trivago/go-kubernetes/commit/3e9fc2e2994e5f1cf363ec799d47484eacb3cc9b))

## [2.4.0](https://github.com/trivago/go-kubernetes/compare/v2.3.0...v2.4.0) (2024-01-11)


### Features

* add GetSection and GetList convenience functions ([fe7bf17](https://github.com/trivago/go-kubernetes/commit/fe7bf17e7df6771a42a41927cb7c0d7a1a9d9b95))
* Add support for LabelSelector based queries ([#29](https://github.com/trivago/go-kubernetes/issues/29)) ([fe7bf17](https://github.com/trivago/go-kubernetes/commit/fe7bf17e7df6771a42a41927cb7c0d7a1a9d9b95))
* add support for LabelSelector when listing objects ([fe7bf17](https://github.com/trivago/go-kubernetes/commit/fe7bf17e7df6771a42a41927cb7c0d7a1a9d9b95))
* allow list for namespaced resources ([dcbdbaa](https://github.com/trivago/go-kubernetes/commit/dcbdbaab4dcd712841b96482640d11d409561c44))

## [2.3.0](https://github.com/trivago/go-kubernetes/compare/v2.2.0...v2.3.0) (2023-11-17)


### Features

* Implement k8s unstructured object API ([3a596fa](https://github.com/trivago/go-kubernetes/commit/3a596fabb4b359d8f1761ba4ffd4141065a76591))

## [2.2.0](https://github.com/trivago/go-kubernetes/compare/v2.1.0...v2.2.0) (2023-11-07)


### Features

* Add a function to list available contexts ([e6f1372](https://github.com/trivago/go-kubernetes/commit/e6f13722664c861520162c642c48e63758c5880b))
* allow kubeconfig context selection ([ed89650](https://github.com/trivago/go-kubernetes/commit/ed8965000261210abe59638bec7e0e013ce7443a))
* Introduce NewClientUsingContext to make NewClient non-breaking. ([0682d96](https://github.com/trivago/go-kubernetes/commit/0682d96de9c191207025ad1d52b618cbf287023a))


### Miscellaneous

* revert to release auto-detection ([a6a4adb](https://github.com/trivago/go-kubernetes/commit/a6a4adbacabd5c5362e2f223ddf1414250d460a0))

## [2.1.0](https://github.com/trivago/go-kubernetes/compare/v2.0.0...v2.1.0) (2023-11-07)


### Features

* add client.Apply ([98e1442](https://github.com/trivago/go-kubernetes/commit/98e14423127f5e18a4676bef71387378b81f1929))
* add client.Delete ([98e1442](https://github.com/trivago/go-kubernetes/commit/98e14423127f5e18a4676bef71387378b81f1929))
* add kubernetes delete and apply ([#13](https://github.com/trivago/go-kubernetes/issues/13)) ([98e1442](https://github.com/trivago/go-kubernetes/commit/98e14423127f5e18a4676bef71387378b81f1929))

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
