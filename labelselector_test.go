package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	serviceJSON = `{
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
      "labels": {
        "app.kubernetes.io/instance": "test",
        "app.kubernetes.io/name": "test"
      },
      "name": "test",
      "namespace": "test"
    },
    "spec": {
      "clusterIP": "10.245.142.171",
      "clusterIPs": [
        "10.245.142.171"
      ],
      "internalTrafficPolicy": "Cluster",
      "ipFamilies": [
        "IPv4"
      ],
      "ipFamilyPolicy": "SingleStack",
      "ports": [
        {
            "name": "https-endpoint",
            "port": 443,
            "protocol": "TCP",
            "targetPort": "https-main"
        }
      ],
      "selector": {
        "app.kubernetes.io/instance": "test",
        "app.kubernetes.io/name": "test"
      },
      "sessionAffinity": "None",
      "type": "ClusterIP"
    }
	}`

	webhookJSON = `{
    "apiVersion": "admissionregistration.k8s.io/v1",
    "kind": "MutatingWebhookConfiguration",
    "metadata": {
      "name": "test"
		},
    "webhooks": [
      {
        "admissionReviewVersions": [
          "v1"
        ],
        "clientConfig": {
          "caBundle": "",
          "service": {
            "name": "test",
            "namespace": "test",
            "path": "/",
            "port": 443
          }
        },
        "failurePolicy": "Ignore",
        "matchPolicy": "Equivalent",
        "name": "test.test.svc.cluster.local",
        "namespaceSelector": {
          "matchExpressions": [
            {
              "key": "kubernetes.io/metadata.name",
              "operator": "NotIn",
              "values": [
                "kube-system",
                "kube-node-lease",
                "kube-public",
                "istio-system",
                "monitoring"
              ]
            }
          ]
        },
        "objectSelector": {
          "matchExpressions": [
            {
              "key": "trivago.com/test",
              "operator": "In",
              "values": [
               	"true"
              ]
            }
          ]
        },
        "reinvocationPolicy": "Never",
        "rules": [
          {
            "apiGroups": [
              ""
            ],
            "apiVersions": [
              "v1"
            ],
            "operations": [
              "CREATE",
              "UPDATE",
              "DELETE"
            ],
            "resources": [
              "configmaps"
            ],
            "scope": "Namespaced"
          }
        ],
        "sideEffects": "NoneOnDryRun",
        "timeoutSeconds": 10
      }
    ]
	}`
)

func TestParseServiceSelector(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(serviceJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)

	selectorValue, err := obj.Get(Path{"spec", "selector"})
	assert.NoError(t, err)

	selectorMap, ok := selectorValue.(map[string]interface{})
	assert.True(t, ok)

	selector, err := ParseLabelSelector(selectorMap)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(selector.MatchLabels))
	assert.Equal(t, 0, len(selector.MatchExpressions))

	assert.Equal(t, "test", selector.MatchLabels["app.kubernetes.io/instance"])
	assert.Equal(t, "test", selector.MatchLabels["app.kubernetes.io/name"])
}

func TestParseNamespaceSelector(t *testing.T) {
	json := runtime.RawExtension{
		Raw: []byte(webhookJSON),
	}

	obj, err := NamedObjectFromRaw(&json)
	assert.NoError(t, err)

	selectorMap, err := obj.GetSection(Path{"webhooks", "0", "namespaceSelector"})
	assert.NoError(t, err)

	selector, err := ParseLabelSelector(selectorMap)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(selector.MatchLabels))
	assert.Equal(t, 1, len(selector.MatchExpressions))

	expression := selector.MatchExpressions[0]

	assert.Equal(t, 5, len(expression.Values))
	assert.Equal(t, "kubernetes.io/metadata.name", expression.Key)
	assert.Equal(t, metav1.LabelSelectorOpNotIn, expression.Operator)
}
