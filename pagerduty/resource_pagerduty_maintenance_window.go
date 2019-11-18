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
		APIObject: pagerduty.APIObject{
			ID: d.Id(),
		},
		StartTime: d.Get("start_time").(string),
		EndTime:   d.Get("end_time").(string),
		Services:  expandServices(d.Get("services").(*schema.Set)),
	}

	if v, ok := d.GetOk("description"); ok {
		window.Description = v.(string)
	}

	return window
}

func resourcePagerDutyMaintenanceWindowCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	window := buildMaintenanceWindowStruct(d)

	log.Printf("[INFO] Creating PagerDuty maintenance window")

	maintenanceWindow, err := client.CreateMaintenanceWindows(window)
	if err != nil {
		return err
	}

	d.SetId(maintenanceWindow.ID)

	return nil
}

func resourcePagerDutyMaintenanceWindowRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty maintenance window %s", d.Id())

	maintenanceWindow, err := client.GetMaintenanceWindow(d.Id(), pagerduty.GetMaintenanceWindowOptions{})
	if err != nil {
		return handleNotFoundError(err, d)
	}

	d.Set("description", maintenanceWindow.Description)
	d.Set("start_time", maintenanceWindow.StartTime)
	d.Set("end_time", maintenanceWindow.EndTime)

	if err := d.Set("services", flattenServices(maintenanceWindow.Services)); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutyMaintenanceWindowUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	window := buildMaintenanceWindowStruct(d)

	log.Printf("[INFO] Updating PagerDuty maintenance window %s", d.Id())

	if _, err := client.UpdateMaintenanceWindow(window); err != nil {
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

func expandServices(v *schema.Set) []pagerduty.APIObject {
	var services []pagerduty.APIObject

	for _, srv := range v.List() {
		service := pagerduty.APIObject{
			Type: "service_reference",
			ID:   srv.(string),
		}
		services = append(services, service)
	}

	return services
}

func flattenServices(v []pagerduty.APIObject) *schema.Set {
	var services []interface{}

	for _, srv := range v {
		services = append(services, srv.ID)
	}

	return schema.NewSet(schema.HashString, services)
}
