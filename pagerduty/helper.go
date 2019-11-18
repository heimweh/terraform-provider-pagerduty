package pagerduty

import "github.com/PagerDuty/go-pagerduty"

func flattenAPIReference(v ...pagerduty.APIReference) []string {
	var res []string

	for _, ref := range v {
		res = append(res, ref.ID)
	}

	return res
}

func expandAPIReference(v interface{}, referenceType string) []pagerduty.APIReference {
	var res []pagerduty.APIReference

	for _, id := range v.([]interface{}) {
		ref := pagerduty.APIReference{
			ID:   id.(string),
			Type: referenceType,
		}

		res = append(res, ref)
	}

	return res
}
