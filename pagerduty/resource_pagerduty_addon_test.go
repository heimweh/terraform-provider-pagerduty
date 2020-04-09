package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("pagerduty_addon", &resource.Sweeper{
		Name: "pagerduty_addon",
		F:    testSweepAddon,
	})
}

func testSweepAddon(region string) error {
	config, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client, err := config.Client()
	if err != nil {
		return err
	}

	res, err := client.ListAddons(pagerduty.ListAddonOptions{})
	if err != nil {
		return err
	}

	for _, addon := range res.Addons {
		if strings.HasPrefix(addon.Name, "test") || strings.HasPrefix(addon.Name, "tf-") {
			log.Printf("Destroying add-on %s (%s)", addon.Name, addon.ID)
			if err := client.DeleteAddon(addon.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccPagerDutyAddon_Basic(t *testing.T) {
	addon := fmt.Sprintf("tf-%s", acctest.RandString(5))
	addonUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyAddonConfig(addon),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyAddonExists("pagerduty_addon.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_addon.foo", "name", addon),
					resource.TestCheckResourceAttr(
						"pagerduty_addon.foo", "src", "https://intranet.foo.com/status"),
				),
			},
			{
				Config: testAccCheckPagerDutyAddonConfigUpdated(addonUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyAddonExists("pagerduty_addon.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_addon.foo", "name", addonUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_addon.foo", "src", "https://intranet.bar.com/status"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyAddonDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_addon" {
			continue
		}

		if _, err := client.GetAddon(r.Primary.ID); err == nil {
			return fmt.Errorf("Add-on still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutyAddonExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No add-on ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, err := client.GetAddon(rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Add-on not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyAddonConfig(addon string) string {
	return fmt.Sprintf(`
resource "pagerduty_addon" "foo" {
  name = "%s"
  src  = "https://intranet.foo.com/status"
}
`, addon)
}

func testAccCheckPagerDutyAddonConfigUpdated(addon string) string {
	return fmt.Sprintf(`
resource "pagerduty_addon" "foo" {
  name = "%s"
  src  = "https://intranet.bar.com/status"
}
`, addon)
}
