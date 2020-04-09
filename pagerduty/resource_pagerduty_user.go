package pagerduty

import (
	"fmt"
	"log"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourcePagerDutyUser() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyUserCreate,
		Read:   resourcePagerDutyUserRead,
		Update: resourcePagerDutyUserUpdate,
		Delete: resourcePagerDutyUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"email": {
				Type:     schema.TypeString,
				Required: true,
			},

			"color": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"role": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "user",
			},

			"job_title": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"avatar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"teams": {
				Type:       schema.TypeSet,
				Deprecated: "Use the 'pagerduty_team_membership' resource instead.",
				Computed:   true,
				Optional:   true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"time_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"invitation_sent": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
		},
	}
}

func buildUserStruct(d *schema.ResourceData) pagerduty.User {
	user := pagerduty.User{
		Name:  d.Get("name").(string),
		Email: d.Get("email").(string),
		APIObject: pagerduty.APIObject{
			ID: d.Id(),
		},
	}

	if attr, ok := d.GetOk("color"); ok {
		user.Color = attr.(string)
	}

	if attr, ok := d.GetOk("time_zone"); ok {
		user.Timezone = attr.(string)
	}

	if attr, ok := d.GetOk("role"); ok {
		user.Role = attr.(string)
	}

	if attr, ok := d.GetOk("job_title"); ok {
		user.JobTitle = attr.(string)
	}

	if attr, ok := d.GetOk("description"); ok {
		user.Description = attr.(string)
	}

	return user
}

func resourcePagerDutyUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	req := buildUserStruct(d)

	log.Printf("[INFO] Creating PagerDuty user %s", req.Name)

	user, err := client.CreateUser(req)
	if err != nil {
		return err
	}

	d.SetId(user.ID)

	return resourcePagerDutyUserUpdate(d, meta)
}

func resourcePagerDutyUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty user %s", d.Id())

	user, err := client.GetUser(d.Id(), pagerduty.GetUserOptions{})
	if err != nil {
		return handleNotFoundError(err, d)
	}

	d.Set("name", user.Name)
	d.Set("email", user.Email)
	d.Set("time_zone", user.Timezone)
	d.Set("html_url", user.HTMLURL)
	d.Set("color", user.Color)
	d.Set("role", user.Role)
	d.Set("avatar_url", user.AvatarURL)
	d.Set("description", user.Description)
	d.Set("job_title", user.JobTitle)
	d.Set("invitation_sent", user.InvitationSent)

	if err := d.Set("teams", flattenTeams(user.Teams)); err != nil {
		return fmt.Errorf("error setting teams: %s", err)
	}

	return nil
}

func resourcePagerDutyUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	req := buildUserStruct(d)

	log.Printf("[INFO] Updating PagerDuty user %s", d.Id())

	if _, err := client.UpdateUser(req); err != nil {
		return err
	}

	if d.HasChange("teams") {
		o, n := d.GetChange("teams")

		if o == nil {
			o = new(schema.Set)
		}

		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		remove := expandStringList(os.Difference(ns).List())
		add := expandStringList(ns.Difference(os).List())

		for _, t := range remove {

			if _, err := client.GetTeam(t); err != nil {
				log.Printf("[INFO] PagerDuty team: %s not found, removing dangling team reference for user %s", t, d.Id())
				continue
			}

			log.Printf("[INFO] Removing PagerDuty user %s from team: %s", d.Id(), t)

			if err := client.RemoveUserFromTeam(t, d.Id()); err != nil {
				return err
			}
		}

		for _, t := range add {
			log.Printf("[INFO] Adding PagerDuty user %s to team: %s", d.Id(), t)

			if err := client.AddUserToTeam(t, d.Id()); err != nil {
				return err
			}
		}
	}

	return resourcePagerDutyUserRead(d, meta)
}

func resourcePagerDutyUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty user %s", d.Id())

	if err := client.DeleteUser(d.Id()); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func flattenTeams(teams []pagerduty.Team) []string {
	res := make([]string, len(teams))
	for i, t := range teams {
		res[i] = t.ID
	}

	return res
}
