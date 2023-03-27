package kubernetes

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	// ResourceConfigMap is the most commonly used GVR for ConfigMaps
	ResourceConfigMap = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	}

	// ResourceNamespace is the most commonly used GVR for Namespaces
	ResourceNamespace = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}

	// ResourceNode is the most commonly used GVR for Nodes
	ResourceNode = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "nodes",
	}

	// ResourcePod is the most commonly used GVR for Pods
	ResourcePod = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	// ResourceSecret is the most commonly used GVR for Secrets
	ResourceSecret = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}

	// ResourceService is the most commonly used GVR for Services
	ResourceService = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}

	// ResourceServiceAccount is the most commonly used GVR for ServiceAccounts
	ResourceServiceAccount = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "serviceaccounts",
	}

	// ResourceDaemonSet is the most commonly used GVR for DaemonSets
	ResourceDaemonSet = schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "daemonsets",
	}

	// ResourceDeployment is the most commonly used GVR for Deployments
	ResourceDeployment = schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	// ResourceStatefulSet is the most commonly used GVR for StatefulSets
	ResourceStatefulSet = schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "statefulsets",
	}
)
