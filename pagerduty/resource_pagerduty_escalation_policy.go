package pagerduty

import (
	"fmt"
	"log"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourcePagerDutyEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyEscalationPolicyCreate,
		Read:   resourcePagerDutyEscalationPolicyRead,
		Update: resourcePagerDutyEscalationPolicyUpdate,
		Delete: resourcePagerDutyEscalationPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"num_loops": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"teams": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"escalation_delay_in_minutes": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"target": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "user_reference",
									},
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildEscalationPolicyStruct(d *schema.ResourceData) *pagerduty.EscalationPolicy {
	escalationPolicy := &pagerduty.EscalationPolicy{
		Name:            d.Get("name").(string),
		EscalationRules: expandEscalationRules(d.Get("rule").([]interface{})),
	}

	if attr, ok := d.GetOk("description"); ok {
		escalationPolicy.Description = attr.(string)
	}

	if attr, ok := d.GetOk("num_loops"); ok {
		escalationPolicy.NumLoops = uint(attr.(int))
	}

	if attr, ok := d.GetOk("teams"); ok {
		escalationPolicy.Teams = expandAPIReference(attr, "team_reference")
	}

	return escalationPolicy
}

func resourcePagerDutyEscalationPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	escalationPolicy := buildEscalationPolicyStruct(d)

	log.Printf("[INFO] Creating PagerDuty escalation policy: %s", escalationPolicy.Name)

	escalationPolicy, err := client.CreateEscalationPolicy(*escalationPolicy)
	if err != nil {
		return err
	}

	d.SetId(escalationPolicy.ID)

	return resourcePagerDutyEscalationPolicyRead(d, meta)
}

func resourcePagerDutyEscalationPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty escalation policy: %s", d.Id())

	o := &pagerduty.GetEscalationPolicyOptions{}

	escalationPolicy, err := client.GetEscalationPolicy(d.Id(), o)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	d.Set("name", escalationPolicy.Name)
	d.Set("description", escalationPolicy.Description)
	d.Set("num_loops", escalationPolicy.NumLoops)

	if err := d.Set("teams", flattenAPIReference(escalationPolicy.Teams...)); err != nil {
		return fmt.Errorf("error setting teams: %s", err)
	}

	if err := d.Set("rule", flattenEscalationRules(escalationPolicy.EscalationRules)); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutyEscalationPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	escalationPolicy := buildEscalationPolicyStruct(d)

	log.Printf("[INFO] Updating PagerDuty escalation policy: %s", d.Id())

	if _, err := client.UpdateEscalationPolicy(d.Id(), escalationPolicy); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutyEscalationPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty escalation policy: %s", d.Id())

	if err := client.DeleteEscalationPolicy(d.Id()); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func expandEscalationRules(v interface{}) []pagerduty.EscalationRule {
	var escalationRules []pagerduty.EscalationRule

	for _, er := range v.([]interface{}) {
		rer := er.(map[string]interface{})
		escalationRule := pagerduty.EscalationRule{
			Delay: uint(rer["escalation_delay_in_minutes"].(int)),
		}

		for _, ert := range rer["target"].([]interface{}) {
			rert := ert.(map[string]interface{})
			escalationRuleTarget := pagerduty.APIObject{
				ID:   rert["id"].(string),
				Type: rert["type"].(string),
			}

			escalationRule.Targets = append(escalationRule.Targets, escalationRuleTarget)
		}

		escalationRules = append(escalationRules, escalationRule)
	}

	return escalationRules
}

func flattenEscalationRules(v []pagerduty.EscalationRule) []map[string]interface{} {
	var escalationRules []map[string]interface{}

	for _, er := range v {
		escalationRule := map[string]interface{}{
			"id":                          er.ID,
			"escalation_delay_in_minutes": er.Delay,
		}

		var targets []map[string]interface{}

		for _, ert := range er.Targets {
			escalationRuleTarget := map[string]interface{}{"id": ert.ID, "type": ert.Type}
			targets = append(targets, escalationRuleTarget)
		}

		escalationRule["target"] = targets

		escalationRules = append(escalationRules, escalationRule)
	}

	return escalationRules
}

func expandTeams(v interface{}) []pagerduty.APIReference {
	var teams []pagerduty.APIReference

	for _, t := range v.([]interface{}) {
		team := pagerduty.APIReference{
			ID:   t.(string),
			Type: "team_reference",
		}
		teams = append(teams, team)
	}

	return teams
}

func flattenTeamReferences(teams []pagerduty.APIReference) []string {
	res := make([]string, len(teams))
	for i, t := range teams {
		res[i] = t.ID
	}

	return res
}
