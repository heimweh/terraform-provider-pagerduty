package pagerduty

import (
	"encoding/json"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func flattenAPIObject(in ...pagerduty.APIObject) []string {
	var out []string

	for _, ref := range in {
		out = append(out, ref.ID)
	}

	return out
}

func expandAPIObject(objectType string, in ...interface{}) []pagerduty.APIObject {
	var out []pagerduty.APIObject

	for _, ref := range in {
		out = append(out, pagerduty.APIObject{
			ID:   ref.(string),
			Type: objectType,
		})
	}

	return out
}

func flattenAPIRef(in ...pagerduty.APIReference) []string {
	var out []string

	for _, ref := range in {
		out = append(out, ref.ID)
	}

	return out
}

func flattenAPIObjectSet(in ...pagerduty.APIReference) *schema.Set {
	var out []interface{}

	for _, ref := range in {
		out = append(out, ref.ID)
	}

	return schema.NewSet(schema.HashString, out)
}

func expandAPIObjectSet(refType string, in *schema.Set) []pagerduty.APIObject {
	var out []pagerduty.APIObject

	for _, ref := range in.List() {
		out = append(out, pagerduty.APIObject{
			ID:   ref.(string),
			Type: refType,
		})
	}

	return out
}

func expandAPIRef(refType string, in ...interface{}) []pagerduty.APIReference {
	var out []pagerduty.APIReference

	for _, ref := range in {
		out = append(out, pagerduty.APIReference{
			ID:   ref.(string),
			Type: refType,
		})
	}

	return out
}

func expandObject(in interface{}) interface{} {
	var out interface{}

	if err := json.Unmarshal([]byte(in.(string)), &out); err != nil {
		return nil
	}

	return out
}

func flattenObject(in interface{}) interface{} {
	out, err := json.Marshal(in)
	if err != nil {
		return nil
	}

	return string(out)
}
