package pagerduty

import (
	"fmt"
	"log"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourcePagerDutyService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyServiceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourcePagerDutyServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty service")

	searchName := d.Get("name").(string)

	o := pagerduty.ListServiceOptions{
		Query: searchName,
	}

	resp, err := client.ListServices(o)
	if err != nil {
		return err
	}

	var found *pagerduty.Service

	for _, service := range resp.Services {
		if service.Name == searchName {
			found = &service
			break
		}
	}

	if found == nil {
		return fmt.Errorf("Unable to locate any service with the name: %s", searchName)
	}

	d.SetId(found.ID)
	d.Set("name", found.Name)

	return nil
}
