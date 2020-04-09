package pagerduty

import (
	"fmt"
	"log"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourcePagerDutyEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyEscalationPolicyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourcePagerDutyEscalationPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty escalation policy")

	searchName := d.Get("name").(string)

	opts := pagerduty.ListEscalationPoliciesOptions{
		Query: searchName,
	}

	res, err := client.ListEscalationPolicies(opts)
	if err != nil {
		return err
	}

	var found *pagerduty.EscalationPolicy

	for _, policy := range res.EscalationPolicies {
		if policy.Name == searchName {
			found = &policy
			break
		}
	}

	if found == nil {
		return fmt.Errorf("Unable to locate any escalation policy with the name: %s", searchName)
	}

	d.SetId(found.ID)
	d.Set("name", found.Name)

	return nil
}
