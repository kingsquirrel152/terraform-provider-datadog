package datadog

import (
	"fmt"
	"sort"
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
			// Seemingly there is not a "nice" way to accomadate a posix timestamp in normal TF workflow.
			// "date_happened": {
			// 	Type:        schema.TypeInt,
			// 	Required:    false,
			// 	Description: "POSIX timestamp of the event. Must be sent as an integer",
			// },
			"priority": {
				Type:         schema.TypeString,
				Required:     false,
				Description:  "The priority of the event: normal or low",
				ValidateFunc: validateDatadogEventPriorityType,
				ForceNew:     true,
			},
			"host": {
				Type:        schema.TypeString,
				Required:    false,
				Description: "Host name to associate with the event. Any tags associated with the host are also applied to this event",
				ForceNew:    true,
			},
			"tags": {
				Type:        schema.TypeSet,
				Required:    false,
				Description: "A list of tags to apply to the event",
				Elem:        &schema.Schema{Type: schema.TypeString},
				ForceNew:    true,
			},
			"alert_type": {
				Type:         schema.TypeString,
				Required:     false,
				Description:  "If it’s an alert event, set its type between: error, warning, info, and success.",
				ValidateFunc: validateDatadogEventAlertType,
				ForceNew:     true,
			},
			"aggregation_key": {
				Type:        schema.TypeString,
				Required:    false,
				Description: "An arbitrary string to use for aggregation. Limited to 100 characters. If you specify a key, all events using that key are grouped together in the Event Stream",
				ForceNew:    true,
			},
			"source_type_name": {
				Type:        schema.TypeString,
				Required:    false,
				Description: "The type of event being posted. Options: nagios, hudson, jenkins, my_apps, chef, puppet, git, bitbucket…",
				ForceNew:    true,
			},
			"device_name": {
				Type:        schema.TypeList,
				Required:    false,
				Description: "A list of device names to post the event with.",
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
	event.SetPriority(d.Get("priority").(string))
	event.SetHost(d.Get("host").(string))

	if attr, ok := d.GetOk("tags"); ok {
		tags := []string{}
		for _, s := range attr.(*schema.Set).List() {
			tags = append(tags, s.(string))
		}
		sort.Strings(tags)
		event.Tags = tags
	}

	event.SetAlertType(d.Get("alert_type").(string))
	event.SetAggregation(d.Get("aggregation_key").(string))
	event.SetSourceType(d.Get("source_type_name").(string))
	return &event, nil
}

func validateDatadogEventPriorityType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	switch value {
	case "normal", "low":
		break
	default:
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid recurrence type parameter %q. Valid parameters are normal, low", k, value))
	}
	return
}

func validateDatadogEventAlertType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	switch value {
	case "error", "warning", "info", "success":
		break
	default:
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid recurrence type parameter %q. Valid parameters are error, warning, info, or success", k, value))
	}
	return
}
