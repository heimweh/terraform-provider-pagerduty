package pagerduty

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourcePagerDutyVendor() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyVendorRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:     schema.TypeString,
				Optional: true,
				Removed:  "Use `name` instead. This attribute will be removed in a future version",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePagerDutyVendorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty vendor")

	searchName := d.Get("name").(string)

	opts := pagerduty.ListVendorOptions{
		Query: searchName,
	}

	res, err := client.ListVendors(opts)
	if err != nil {
		return err
	}

	var found *pagerduty.Vendor

	for _, vendor := range res.Vendors {
		if strings.EqualFold(vendor.Name, searchName) {
			found = &vendor
			break
		}
	}

	// We didn't find an exact match, so let's fallback to partial matching.
	if found == nil {
		pr := regexp.MustCompile("(?i)" + searchName)
		for _, vendor := range res.Vendors {
			if pr.MatchString(vendor.Name) {
				found = &vendor
				break
			}
		}
	}

	if found == nil {
		return fmt.Errorf("Unable to locate any vendor with the name: %s", searchName)
	}

	d.SetId(found.ID)
	d.Set("name", found.Name)
	d.Set("type", found.GenericServiceType)

	return nil
}
