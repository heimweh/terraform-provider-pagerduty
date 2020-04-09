package pagerduty

import (
	"log"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourcePagerDutyMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyMaintenanceWindowCreate,
		Read:   resourcePagerDutyMaintenanceWindowRead,
		Update: resourcePagerDutyMaintenanceWindowUpdate,
		Delete: resourcePagerDutyMaintenanceWindowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"start_time": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateRFC3339,
				DiffSuppressFunc: suppressRFC3339Diff,
			},
			"end_time": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateRFC3339,
				DiffSuppressFunc: suppressRFC3339Diff,
			},

			"services": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
		},
	}
}

func buildMaintenanceWindowStruct(d *schema.ResourceData) pagerduty.MaintenanceWindow {
	window := pagerduty.MaintenanceWindow{
		StartTime: d.Get("start_time").(string),
		EndTime:   d.Get("end_time").(string),
		Services:  expandAPIObjectSet("service_reference", d.Get("services").(*schema.Set)),
		APIObject: pagerduty.APIObject{
			ID: d.Id(),
		},
	}

	if v, ok := d.GetOk("description"); ok {
		window.Description = v.(string)
	}

	return window
}

func resourcePagerDutyMaintenanceWindowCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	req := buildMaintenanceWindowStruct(d)

	log.Printf("[INFO] Creating PagerDuty maintenance window")

	// XXX(heimweh): What do we put in From?
	window, err := client.CreateMaintenanceWindow("", req)
	if err != nil {
		return err
	}

	d.SetId(window.ID)

	return nil
}

func resourcePagerDutyMaintenanceWindowRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty maintenance window %s", d.Id())

	opts := pagerduty.GetMaintenanceWindowOptions{}

	window, err := client.GetMaintenanceWindow(d.Id(), opts)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	d.Set("description", window.Description)
	d.Set("start_time", window.StartTime)
	d.Set("end_time", window.EndTime)

	if err := d.Set("services", flattenAPIObject(window.Services...)); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutyMaintenanceWindowUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	req := buildMaintenanceWindowStruct(d)

	log.Printf("[INFO] Updating PagerDuty maintenance window %s", d.Id())

	if _, err := client.UpdateMaintenanceWindow(req); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutyMaintenanceWindowDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty maintenance window %s", d.Id())

	if err := client.DeleteMaintenanceWindow(d.Id()); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
