package datadog

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	datadog "github.com/zorkian/go-datadog-api"
)

func TestAccDatadogEvent_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogEventConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogEventsExists(),
					resource.TestCheckResourceAttr(
						"datadog_event.foo", "title", "foo"),
					resource.TestCheckResourceAttr(
						"datadog_event.foo", "text", "foo"),
				),
			},
		},
	})
}

func TestAccDatadogEvent_Full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogEventConfigFull,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogEventsExists(),
					resource.TestCheckResourceAttr(
						"datadog_event.foo", "title", "foo"),
					resource.TestCheckResourceAttr(
						"datadog_event.foo", "text", "foo"),
					resource.TestCheckResourceAttr(
						"datadog_event.foo", "priority", "error"),
					resource.TestCheckResourceAttr(
						"datadog_event.foo", "host", "some_host"),
					resource.TestCheckResourceAttr(
						"datadog_event.foo", "alert_type", "info"),
					// Tags are a TypeSet => use a weird way to access members by their hash
					// TF TypeSet is internally represented as a map that maps computed hashes
					// to actual values. Since the hashes are always the same for one value,
					// this is the way to get them.
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.2644851163", "baz"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1750285118", "foo:bar"),
				),
			},
		},
	})
}

func TestAccDatadogEvent_Trigger(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogEventConfigTriggerAlways,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogEventsExists(),
					resource.TestCheckResourceAttr(
						"datadog_event.foo", "title", "foo"),
					resource.TestCheckResourceAttr(
						"datadog_event.foo", "text", "foo"),
				),
			},
			{
				// With a trigger we expect the event to be recreated
				Config:             testAccCheckDatadogEventConfigTriggerAlways,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

const testAccCheckDatadogEventConfigBasic = `
resource "datadog_event" "foo" {
  title = "foo"
  text  = "foo"
}
`

const testAccCheckDatadogEventConfigTriggerAlways = `
resource "datadog_event" "foo" {
  title = "foo"
  text  = "foo"

  triggers = {
    always = "${timestamp()}"
  }
}
`

const testAccCheckDatadogEventConfigFull = `
resource "datadog_event" "foo" {
  title      = "foo"
  text       = "foo"
  priority   = "error"
  host       = "some_host"
  alert_type = "info"
  
  tags = ["foo:bar", "baz"]
}
`

// Check to see if events in terraform state exist in the datadog API
func testAccCheckDatadogEventsExists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		if err := datadogEventExistsHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

// NOTE: Datadog events seem to be eventually consistent, so we retry the query
// and timeout after a short period of time
func datadogEventExistsHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_event" {
			continue
		}
		id, _ := strconv.Atoi(r.Primary.ID)

		timeout := time.After(10 * time.Second)
		tick := time.Tick(1 * time.Second)
		// Keep trying until we're timed out or successfully retrieved the event
		for {
			select {
			case <-timeout:
				return fmt.Errorf("Timed out retrieving event")
			case <-tick:
				if _, err := client.GetEvent(id); err == nil {
					break // Event was found and we can move on to the next resource
				}
			}
		}
	}
	return nil
}
