package pagerduty

import (
	"log"

	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourcePagerDutyExtension() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyExtensionCreate,
		Read:   resourcePagerDutyExtensionRead,
		Update: resourcePagerDutyExtensionUpdate,
		Delete: resourcePagerDutyExtensionDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePagerDutyExtensionImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"endpoint_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"extension_objects": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"extension_schema": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"config": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.ValidateJsonString,
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
		},
	}
}

func buildExtensionStruct(d *schema.ResourceData) *pagerduty.Extension {
	Extension := &pagerduty.Extension{
		Name:        d.Get("name").(string),
		EndpointURL: d.Get("endpoint_url").(string),
		ExtensionSchema: pagerduty.APIObject{
			Type: "extension_schema_reference",
			ID:   d.Get("extension_schema").(string),
		},
		ExtensionObjects: expandAPIObjectSet(
			"service_reference",
			d.Get("extension_objects").(*schema.Set),
		),
		APIObject: pagerduty.APIObject{
			Type: "extension",
		},
	}

	if v, ok := d.GetOk("config"); ok {
		Extension.Config = expandObject(v)
	}

	return Extension
}

func resourcePagerDutyExtensionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	extension := buildExtensionStruct(d)

	log.Printf("[INFO] Creating PagerDuty extension %s", extension.Name)

	extension, err := client.CreateExtension(extension)
	if err != nil {
		return err
	}

	d.SetId(extension.ID)

	return resourcePagerDutyExtensionRead(d, meta)
}

func resourcePagerDutyExtensionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty extension %s", d.Id())

	extension, err := client.GetExtension(d.Id())
	if err != nil {
		return handleNotFoundError(err, d)
	}

	d.Set("summary", extension.Summary)
	d.Set("name", extension.Name)
	d.Set("endpoint_url", extension.EndpointURL)
	d.Set("html_url", extension.HTMLURL)

	if err := d.Set("extension_objects", flattenAPIObject(extension.ExtensionObjects...)); err != nil {
		log.Printf("[WARN] error setting extension_objects: %s", err)
	}

	d.Set("extension_schema", extension.ExtensionSchema)

	if err := d.Set("config", flattenObject(extension.Config)); err != nil {
		log.Printf("[WARN] error setting extension config: %s", err)
	}

	return nil
}

func resourcePagerDutyExtensionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	req := buildExtensionStruct(d)

	log.Printf("[INFO] Updating PagerDuty extension %s", d.Id())

	if _, err := client.UpdateExtension(d.Id(), req); err != nil {
		return err
	}

	return resourcePagerDutyExtensionRead(d, meta)
}

func resourcePagerDutyExtensionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty extension %s", d.Id())

	if err := client.DeleteExtension(d.Id()); err != nil {
		return handleNotFoundError(err, d)
	}

	d.SetId("")

	return nil
}

func resourcePagerDutyExtensionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*pagerduty.Client)

	extension, err := client.GetExtension(d.Id())

	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("error importing pagerduty_extension. Expecting an importation ID for extension")
	}

	d.Set("endpoint_url", extension.EndpointURL)
	d.Set("extension_objects", []string{extension.ExtensionObjects[0].ID})
	d.Set("extension_schema", extension.ExtensionSchema.ID)

	return []*schema.ResourceData{d}, err
}
