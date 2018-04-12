package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccPagerDutyExtension_import(t *testing.T) {
	extensionName := fmt.Sprintf("tf-%s", acctest.RandString(5))
	name := fmt.Sprintf("tf-%s", acctest.RandString(5))
	url := "https://example.com/receive_a_pagerduty_webhook"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyExtensionConfig(name, extensionName, url),
			},

			{
				ResourceName:      "pagerduty_extension.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
