package pagerduty

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/schema"
)

func timeToUTC(v string) (string, error) {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return "", err
	}

	return t.UTC().String(), nil
}

// validateRFC3339 validates that a date string has the correct RFC3339 layout
func validateRFC3339(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		errors = append(errors, fmt.Errorf("%s is not a valid format for argument: %s. Expected format: %s (RFC3339)", value, k, time.RFC3339))
	}

	return
}

func suppressRFC3339Diff(k, oldTime, newTime string, d *schema.ResourceData) bool {
	oldT, err := time.Parse(time.RFC3339, oldTime)
	if err != nil {
		log.Printf("[ERROR] Failed to parse %q (old %q). Expected format: %s (RFC3339)", oldTime, k, time.RFC3339)
		return false
	}
	newT, err := time.Parse(time.RFC3339, newTime)
	if err != nil {
		log.Printf("[ERROR] Failed to parse %q (new %q). Expected format: %s (RFC3339)", newTime, k, time.RFC3339)
		return false
	}
	return oldT.Equal(newT)
}

// Validate a value against a set of possible values
func validateValueFunc(values []string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (we []string, errors []error) {
		value := v.(string)
		valid := false
		for _, val := range values {
			if value == val {
				valid = true
				break
			}
		}

		if !valid {
			errors = append(errors, fmt.Errorf("%#v is an invalid value for argument %s. Must be one of %#v", value, k, values))
		}
		return
	}
}

// Takes the result of flatmap.Expand for an array of strings
// and returns a []*string
func expandStringList(configured []interface{}) []*string {
	vs := make([]*string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, aws.String(v.(string)))
		}
	}
	return vs
}

// Takes the result of schema.Set of strings and returns a []*string
func expandStringSet(configured *schema.Set) []*string {
	return expandStringList(configured.List())
}

// Takes list of pointers to strings. Expand to an array
// of raw strings and returns a []interface{}
// to keep compatibility w/ schema.NewSetschema.NewSet
func flattenStringList(list []*string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, *v)
	}
	return vs
}

func flattenStringSet(list []*string) *schema.Set {
	return schema.NewSet(schema.HashString, flattenStringList(list))
}
