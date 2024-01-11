package kubernetes

import (
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ParseLabelSelector parses a label selector from a map[string]interface{}.
// If any of the required keys is of the wrong type, an error is returned as well
// as all keys that were parsed successfully up to that point.
// A valid label selector looks like this in YAML:
//
//	matchLabels:
//	  app.kubernetes.io/instance: test
//	  app.kubernetes.io/name: test
//	matchExpressions:
//	  - key: app.kubernetes.io/instance
//	    operator: In
//	    values:
//	      - test
//
// If neither matchLabels nor matchExpressions are present, the selector is expected
// to be a map[string]string, containing the matchLabels section directly.
func ParseLabelSelector(obj map[string]interface{}) (metav1.LabelSelector, error) {
	var (
		selector metav1.LabelSelector
		ok       bool
	)

	matchLabels, hasMatchLabels := obj["matchLabels"]
	matchExpressions, hasMatchExpressions := obj["matchExpressions"]

	// If there is neither matchLabels nor matchExpressions, we expect the a service-style
	// selector, which is basically just matchLabels
	if !hasMatchLabels && !hasMatchExpressions {
		selector.MatchLabels = make(map[string]string)
		for k, v := range obj {
			stringValue, ok := v.(string)
			if !ok {
				return selector, fmt.Errorf("failed to parse selector[%s] as string : %v", k, v)
			}
			selector.MatchLabels[k] = stringValue
		}
		return selector, nil
	}

	if hasMatchLabels {
		selector.MatchLabels, ok = matchLabels.(map[string]string)
		if !ok {
			// Support any-type key/value maps (named objects)
			untyped, ok := matchLabels.(map[string]interface{})
			if !ok {
				return selector, fmt.Errorf("failed to parse matchLabels as map[string]string or map[string]interface{} : %v", matchLabels)
			}

			selector.MatchLabels = make(map[string]string)
			for k, v := range untyped {
				stringValue, ok := v.(string)
				if !ok {
					return selector, fmt.Errorf("failed to parse matchLabels[%s] as string : %v", k, v)
				}
				selector.MatchLabels[k] = stringValue
			}
		}
	}

	if hasMatchExpressions {
		selector.MatchExpressions, ok = matchExpressions.([]metav1.LabelSelectorRequirement)
		if !ok {
			// Support any-type key/value maps (named objects)
			untypedList, ok := matchExpressions.([]interface{})
			if !ok {
				return selector, fmt.Errorf("failed to parse matchExpressions as []metav1.LabelSelectorRequirement or []interface{} : %v", matchExpressions)
			}

			selector.MatchExpressions = make([]metav1.LabelSelectorRequirement, 0, len(untypedList))
			for i, v := range untypedList {
				untypedMap, ok := v.(map[string]interface{})
				if !ok {
					return selector, fmt.Errorf("failed to parse matchExpressions[%d] as map[string]interface{} : %v", i, v)
				}

				parsed, err := parseLabelSelectorRequirement(untypedMap)
				if err != nil {
					return selector, errors.Wrap(err, "failed to parse matchExpressions")
				}
				selector.MatchExpressions = append(selector.MatchExpressions, parsed)
			}
		}
	}

	return selector, nil
}

// parseLabelSelectorRequirement parses a signle value from a matchExpressions list.
// This function is called internally by ParseLabelSelector.
func parseLabelSelectorRequirement(obj map[string]interface{}) (metav1.LabelSelectorRequirement, error) {
	var (
		req metav1.LabelSelectorRequirement
		ok  bool
	)

	req.Key, ok = obj["key"].(string)
	if !ok {
		return req, fmt.Errorf("failed to parse key as string : %v", obj["key"])
	}

	operatorStr, ok := obj["operator"].(string)
	if !ok {
		return req, fmt.Errorf("failed to parse operator as metav1.LabelSelectorOperator (string) : %v", obj["operator"])
	}
	req.Operator = metav1.LabelSelectorOperator(operatorStr)

	req.Values, ok = obj["values"].([]string)
	if !ok {
		untypedList, ok := obj["values"].([]interface{})
		if !ok {
			return req, fmt.Errorf("failed to parse values as []string or []interface{} : %v", obj["values"])
		}

		req.Values = make([]string, 0, len(untypedList))
		for i, v := range untypedList {
			stringValue, ok := v.(string)
			if !ok {
				return req, fmt.Errorf("failed to parse values[%d] as string : %v", i, v)
			}

			req.Values = append(req.Values, stringValue)
		}
	}

	return req, nil
}
