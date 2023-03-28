package kubernetes

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// Client allows communication with the kubernetes API.
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

// GetNamedObject returns a specific kubernetes object
func (k8s *Client) GetNamedObject(resource schema.GroupVersionResource, name string) (NamedObject, error) {
	resourceHandle := k8s.client.Resource(resource)
	rawObject, err := resourceHandle.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return NamedObjectFromUnstructured(*rawObject)
}

// GetNamespacedObject returns a specific kubernetes object from a specific namespace
func (k8s *Client) GetNamespacedObject(resource schema.GroupVersionResource, name, namespace string) (NamedObject, error) {
	resourceHandle := k8s.client.Resource(resource).Namespace(namespace)
	rawObject, err := resourceHandle.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return NamedObjectFromUnstructured(*rawObject)
}

// ListAllObjects returns a list of objects for a given type
func (k8s *Client) ListAllObjects(resource schema.GroupVersionResource, selector string) ([]NamedObject, error) {
	resourceHandle := k8s.client.Resource(resource)
	options := metav1.ListOptions{
		LabelSelector: selector,
	}

	list, err := resourceHandle.List(context.Background(), options)
	if err != nil {
		return []NamedObject{}, err
	}

	resultList := make([]NamedObject, 0, len(list.Items))

	for _, rawObject := range list.Items {
		obj, parseErr := NamedObjectFromUnstructured(rawObject)
		if parseErr != nil {
			if err != nil {
				err = fmt.Errorf("Error parsing item(s) from list")
			}
			errors.Wrap(err, err.Error())
		}
		resultList = append(resultList, obj)
	}

	return resultList, err
}
