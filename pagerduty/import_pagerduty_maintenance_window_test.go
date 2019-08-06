package pagerduty

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccPagerDutyMaintenanceWindow_import(t *testing.T) {
	window := fmt.Sprintf("tf-%s", acctest.RandString(5))
	windowStartTime := timeNowInAccLoc().Add(24 * time.Hour).Format(time.RFC3339)
	windowEndTime := timeNowInAccLoc().Add(48 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyTeamMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyMaintenanceWindowConfig(window, windowStartTime, windowEndTime),
			},

			{
				ResourceName:      "pagerduty_maintenance_window.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
