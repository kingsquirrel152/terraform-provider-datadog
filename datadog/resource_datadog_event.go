package datadog

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	datadog "github.com/zorkian/go-datadog-api"
)

func resourceDatadogEvent() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogEventCreate,
		Read:   resourceDatadogEventRead,
		Delete: resourceDatadogEventDelete,
		Schema: map[string]*schema.Schema{
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Title of the event",
				ForceNew:    true,
			},
			"text": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The body of the event",
				ForceNew:    true,
			},
			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDatadogEventCreate(d *schema.ResourceData, meta interface{}) error {
	event, err := buildDatadogEvent(d)
	if err != nil {
		return fmt.Errorf("Error building the event %s", err.Error())
	}
	event, err = meta.(*datadog.Client).PostEvent(event)
	if err != nil {
		return fmt.Errorf("Failed to post event using Datadog API: %s", err.Error())
	}
	id := event.GetId()
	d.SetId(strconv.Itoa(id))

	return nil
}

func resourceDatadogEventRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogEventDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

func buildDatadogEvent(d *schema.ResourceData) (*datadog.Event, error) {
	var event datadog.Event
	event.SetTitle(d.Get("title").(string))
	event.SetText(d.Get("text").(string))
	return &event, nil
}
