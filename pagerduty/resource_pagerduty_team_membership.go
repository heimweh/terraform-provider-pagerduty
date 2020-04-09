package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourcePagerDutyTeamMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyTeamMembershipCreate,
		Read:   resourcePagerDutyTeamMembershipRead,
		Update: resourcePagerDutyTeamMembershipUpdate,
		Delete: resourcePagerDutyTeamMembershipDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "manager",
				ValidateFunc: validateValueFunc([]string{
					"observer",
					"responder",
					"manager",
				}),
			},
		},
	}
}
func resourcePagerDutyTeamMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	userID := d.Get("user_id").(string)
	teamID := d.Get("team_id").(string)
	// role := d.Get("role").(string)

	log.Printf("[DEBUG] Adding user: %s to team: %s", userID, teamID)

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		// XXX(heimweh): This is missing.
		// if _, err := client.AddUserWithRole(teamID, userID, role); err != nil {
		if err := client.AddUserToTeam(teamID, userID); err != nil {
			if isErrCode(err, 500) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, teamID))

	return resourcePagerDutyTeamMembershipRead(d, meta)
}

func resourcePagerDutyTeamMembershipRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	userID, teamID := resourcePagerDutyTeamMembershipParseID(d.Id())

	log.Printf("[DEBUG] Reading user: %s from team: %s", userID, teamID)

	res, err := client.ListMembers(teamID, pagerduty.ListMembersOptions{})
	if err != nil {
		return handleNotFoundError(err, d)
	}

	for _, member := range res.Members {
		if member.APIObject.ID == userID {
			d.Set("user_id", userID)
			d.Set("team_id", teamID)
			d.Set("role", member.Role)

			return nil
		}
	}

	log.Printf("[WARN] Removing %s since the user: %s is not a member of: %s", d.Id(), userID, teamID)
	d.SetId("")

	return nil
}

func resourcePagerDutyTeamMembershipUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	userID := d.Get("user_id").(string)
	teamID := d.Get("team_id").(string)
	role := d.Get("role").(string)

	log.Printf("[DEBUG] Updating user: %s to team: %s with role: %s", userID, teamID, role)

	// To update existing membership resource, We can use the same API as creating a new membership.
	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		/// XXX(heimweh): This is missing.
		// if _, err := client.AddUserWithRole(teamID, userID, role); err != nil {
		if err := client.AddUserToTeam(teamID, userID); err != nil {
			if isErrCode(err, 500) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, teamID))

	return nil
}

func resourcePagerDutyTeamMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	userID, teamID := resourcePagerDutyTeamMembershipParseID(d.Id())

	log.Printf("[DEBUG] Removing user: %s from team: %s", userID, teamID)

	if err := client.RemoveUserFromTeam(teamID, userID); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourcePagerDutyTeamMembershipParseID(id string) (string, string) {
	parts := strings.Split(id, ":")
	return parts[0], parts[1]
}

func isTeamMember(user *pagerduty.User, teamID string) bool {
	var found bool

	for _, team := range user.Teams {
		if teamID == team.ID {
			found = true
		}
	}

	return found
}
