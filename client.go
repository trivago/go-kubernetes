package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// Client allows communication with the kubernetes API.
type Client struct {
	apiClient           typedClients
	client              dynamic.Interface
	discoveryClient     *discovery.DiscoveryClient
	groupResourceMapper meta.RESTMapper

	schemaCache map[string]schema.GroupVersionKind
}

// typedClients holds kubernetes clients for different API groups.
type typedClients struct {
	corev1 *corev1client.CoreV1Client
}

// NewClusterClient creates a new kubernetes client for the current cluster.
func NewClusterClient() (*Client, error) {
	return NewClientUsingContext("", "")
}

// NewClient creates a new kubernetes client for a given path to a kubeconfig.
// The client will use the default context from the kubeconfig file.
func NewClient(path string) (*Client, error) {
	return NewClientUsingContext(path, "")
}

// NewClientUsingContext creates a new kubernetes client for a given path to a
// kubeconfig file. If no file is given, an in-cluster client will be created.
// The context parameter can be used to specify a specific context from the
// kubeconfig file. When left empty, the default context will be used.
func NewClientUsingContext(path, context string) (*Client, error) {
	var (
		err    error
		config *restclient.Config
	)

	k8sClient := Client{
		schemaCache: make(map[string]schema.GroupVersionKind),
	}

	if path == "" {
		// In cluster client if path is empty
		config, err = restclient.InClusterConfig()
		if err != nil {
			log.Error().Msg("failed to build in-cluster kubeconfig")
			return nil, err
		}
	} else {
		// Out of cluster client if path is given.
		rules := clientcmd.ClientConfigLoadingRules{
			ExplicitPath: path,
		}
		// Support context overrides
		overrides := clientcmd.ConfigOverrides{
			CurrentContext: context,
		}
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&rules, &overrides).ClientConfig()
		if err != nil {
			log.Error().Msgf("failed to load kubeconfig from %s", path)
			return nil, err
		}
	}

	k8sClient.client, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Error().Msg("failed to create in-cluster kubernetes client")
		return nil, err
	}

	k8sClient.apiClient.corev1, err = corev1client.NewForConfig(config)
	if err != nil {
		log.Error().Msg("failed to create in-cluster kubernetes core v1 client")
		return nil, err
	}

	k8sClient.discoveryClient, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Error().Msg("failed to create in-cluster kubernetes discovery client")
		return nil, err
	}

	groupResources, err := restmapper.GetAPIGroupResources(k8sClient.discoveryClient)
	if err != nil {
		log.Error().Msg("failed to create in-cluster kubernetes group resource mapper")
		return nil, err
	}
	k8sClient.groupResourceMapper = restmapper.NewDiscoveryRESTMapper(groupResources)

	return &k8sClient, nil
}

// GetContextsFromConfig reads a kubeconfig file and returns a list of contexts names.
func GetContextsFromConfig(path string) ([]string, error) {
	rules := clientcmd.ClientConfigLoadingRules{
		ExplicitPath: path,
	}
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&rules, &clientcmd.ConfigOverrides{})
	rawConfig, err := config.RawConfig()
	if err != nil {
		return []string{}, err
	}

	contextNames := make([]string, 0, len(rawConfig.Contexts))
	for name := range rawConfig.Contexts {
		contextNames = append(contextNames, name)
	}

	return contextNames, nil
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

// ListAllObjects returns a list of all objects for a given type that is assumed to be global.
func (k8s *Client) ListAllObjects(resource schema.GroupVersionResource, labelSelector, fieldSelector string) ([]NamedObject, error) {
	return k8s.list(resource, "", labelSelector, fieldSelector)
}

// ListAllObjectsInNamespace returns a list of all objects for a given type in a given namespace.
func (k8s *Client) ListAllObjectsInNamespace(resource schema.GroupVersionResource, namespace, labelSelector, fieldSelector string) ([]NamedObject, error) {
	return k8s.list(resource, namespace, labelSelector, fieldSelector)
}

// ListAllObjectsInNamespaceMatching returns a list of all objects matching a given selector struct.
// This struct is used in varios API objects like namespaceSelector or objectSelector.
// Use ParseLabelSelector to create this struct from an existing object.
func (k8s *Client) ListAllObjectsInNamespaceMatching(resource schema.GroupVersionResource, namespace string, labelMatchExpression metav1.LabelSelector, fieldSelector string) ([]NamedObject, error) {
	labelSelector := metav1.FormatLabelSelector(&labelMatchExpression)
	return k8s.list(resource, namespace, labelSelector, fieldSelector)
}

// ListAllObjectsMatching returns a list of all objects matching a given selector struct.
// This struct is used in varios API objects like namespaceSelector or objectSelector.
// Use ParseLabelSelector to create this struct from an existing object.
func (k8s *Client) ListAllObjectsMatching(resource schema.GroupVersionResource, labelMatchExpression metav1.LabelSelector, fieldSelector string) ([]NamedObject, error) {
	labelSelector := metav1.FormatLabelSelector(&labelMatchExpression)
	return k8s.list(resource, "", labelSelector, fieldSelector)
}

// list returns a list of objects for a given type.
// Namespace, labelSelector and fieldSelector are optional arguments. If namespace is left empty,
// a global resource is expected. If selector is left empty, all objects will
// be returned.
func (k8s *Client) list(resource schema.GroupVersionResource, namespace, labelSelector, fieldSelector string) ([]NamedObject, error) {
	start := time.Now()
	defer func() {
		log.Debug().Msgf("list operation took %s", time.Since(start).String())
	}()

	options := metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	}

	var resourceHandle dynamic.ResourceInterface

	if len(namespace) > 0 {
		resourceHandle = k8s.client.Resource(resource).Namespace(namespace)
	} else {
		resourceHandle = k8s.client.Resource(resource)
	}

	list, err := resourceHandle.List(context.Background(), options)
	if err != nil {
		return []NamedObject{}, err
	}

	resultList := make([]NamedObject, 0, len(list.Items))
	for _, rawObject := range list.Items {
		obj, parseErr := NamedObjectFromUnstructured(rawObject)
		if parseErr != nil {
			if err == nil {
				err = parseErr
			} else {
				err = errors.Wrapf(err, "failed to parse item: %v", parseErr)
			}
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

// Patch applies a set of patches on a given kubernetes object.
// The patches are applied as json patches.
func (k8s *Client) Patch(resource schema.GroupVersionResource, object NamedObject, patches []PatchOperation, options metav1.PatchOptions) {
	start := time.Now()
	defer func() {
		log.Debug().Msgf("patch operation took %s", time.Since(start).String())
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

	patchData, err := json.Marshal(patches)
	if err != nil {
		log.Error().Err(err).Interface("patches", patches).Msgf("failed to marshal patch data for %s", identifier)
		return
	}

	if _, err := resourceHandle.Patch(context.Background(), object.GetName(), types.JSONPatchType, patchData, metav1.PatchOptions{}); err != nil {
		log.Error().Err(err).Interface("patches", patches).Msgf("failed to apply patch for %s", identifier)
	} else {
		log.Debug().Msgf("applied %s", identifier)
	}
}

// GetServiceAccountToken returns a token for a given service account.
// This requires the calling service to have the necessary permissions for
// `authentication.k8s.io/tokenrequests`.
func (k8s *Client) GetServiceAccountToken(serviceAccountName, namespace string, expiration time.Duration, audiences []string, pod NamedObject, ctx context.Context) (string, error) {
	expirationSec := int64(expiration.Seconds())
	var boundPodRef authenticationv1.BoundObjectReference

	if len(pod) > 0 {
		boundPodRef = authenticationv1.BoundObjectReference{
			Kind:       pod.GetKind(),
			APIVersion: pod.GetVersion(),
			Name:       pod.GetName(),
			UID:        types.UID(pod.GetUID()),
		}

		if strings.ToLower(boundPodRef.Kind) != "pod" {
			return "", fmt.Errorf("bound object reference must be a pod or nil")
		}
	}

	request := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences:         audiences,
			ExpirationSeconds: &expirationSec,
			BoundObjectRef:    &boundPodRef,
		},
	}

	response, err := k8s.apiClient.corev1.ServiceAccounts(namespace).CreateToken(ctx, serviceAccountName, request, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	if len(response.Status.Token) == 0 {
		return "", fmt.Errorf("no token in server response")
	}

	return response.Status.Token, nil
}
