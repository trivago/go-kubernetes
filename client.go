package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
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
	start := time.Now()
	defer func() {
		log.Debug().Msgf("list operation took %s", time.Since(start).String())
	}()

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
				err = fmt.Errorf("error parsing item(s) from list")
			}
			errors.Wrap(err, err.Error())
		}
		resultList = append(resultList, obj)
	}

	return resultList, err
}

// Apply creates or updates a given kubernetes object.
// If a namespace is set, the object will be created in that namespace.
func (k8s *Client) Apply(resource schema.GroupVersionResource, object NamedObject, options metav1.ApplyOptions) {
	start := time.Now()
	defer func() {
		log.Debug().Msgf("apply operation took %s", time.Since(start).String())
	}()

	var (
		resourceHandle dynamic.ResourceInterface
		identifier     string
	)

	if object.GetNamespace() != "" {
		resourceHandle = k8s.client.Resource(resource).Namespace(object.GetNamespace())
		identifier = fmt.Sprintf("%s/%s", object.GetNamespace(), object.GetName())
	} else {
		resourceHandle = k8s.client.Resource(resource)
		identifier = object.GetName()
	}

	unstructuredObject := &unstructured.Unstructured{
		Object: object,
	}

	if _, err := resourceHandle.Apply(context.Background(), object.GetName(), unstructuredObject, options); err != nil {
		log.Error().Err(err).Interface(object.GetName(), object).Msgf("failed to trigger apply for %s", identifier)
	} else {
		log.Debug().Msgf("applied %s", identifier)
	}
}

// DeleteNamespaced removes a specific kubernetes object from a specific namespace.
// If an empty namespace is given, the object will be treated as a cluster-wide resource.
func (k8s *Client) DeleteNamespaced(resource schema.GroupVersionResource, name, namespace string) {
	start := time.Now()
	defer func() {
		log.Debug().Msgf("delete operation took %s", time.Since(start).String())
	}()

	var (
		resourceHandle dynamic.ResourceInterface
		identifier     string
	)

	if namespace != "" {
		resourceHandle = k8s.client.Resource(resource).Namespace(namespace)
		identifier = fmt.Sprintf("%s/%s", namespace, name)
	} else {
		resourceHandle = k8s.client.Resource(resource)
		identifier = name
	}

	if err := resourceHandle.Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		log.Error().Err(err).Msgf("failed to trigger delete for %s", identifier)
	} else {
		log.Info().Msgf("deleted %s", identifier)
	}
}
