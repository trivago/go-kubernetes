package kubernetes

import (
	"context"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	client              dynamic.Interface
	discoveryClient     *discovery.DiscoveryClient
	groupResourceMapper meta.RESTMapper

	schemaCache map[string]schema.GroupVersionKind
}

// NewClient creates a new kubernetes client for a given path to a
// kubeconfig file. If no file is given, an in-cluster client will be created.
func NewClient(kubeconfig string) (*Client, error) {
	k8sClient := Client{
		schemaCache: make(map[string]schema.GroupVersionKind),
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Error().Msg("failed to build in-cluster kubeconfig")
		return nil, err
	}

	k8sClient.client, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Error().Msg("failed to create in-cluster kubernetes client")
		return nil, err
	}

	k8sClient.discoveryClient, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Error().Msg("failed to create in-cluster kubernetes discovery client")
		return nil, err
	}

	groupResources, err := restmapper.GetAPIGroupResources(k8sClient.discoveryClient)
	k8sClient.groupResourceMapper = restmapper.NewDiscoveryRESTMapper(groupResources)

	return &k8sClient, nil
}

// GetNamespacedResource creates an object to interact with a namespaced resource
func (k8s *Client) GetNamespacedResourceHandle(resource schema.GroupVersionResource, namespace string) dynamic.ResourceInterface {
	return k8s.client.Resource(resource).Namespace(namespace)
}

// GetObject returns a specific kubernetes object
func (k8s *Client) GetObject(resource schema.GroupVersionResource, name, namespace string) (NamedObject, error) {
	resourceHandle := k8s.GetNamespacedResourceHandle(resource, namespace)
	rawObject, err := resourceHandle.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return NamedObjectFromUnstructured(*rawObject)
}

// ListAllObjects returns a list of objects for a given type
func (k8s *Client) ListAllObjects(resource schema.GroupVersionResource, selector string) ([]unstructured.Unstructured, error) {
	resourceHandle := k8s.client.Resource(resource)
	options := metav1.ListOptions{
		LabelSelector: selector,
	}

	list, err := resourceHandle.List(context.Background(), options)
	if err != nil {
		return []unstructured.Unstructured{}, err
	}

	return list.Items, nil
}
