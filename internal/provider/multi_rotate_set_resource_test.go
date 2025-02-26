package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func timeCheckWithVariance(t time.Time, variance string) func(string) error {
	return func(s string) error {
		timeCheck, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return err
		}
		v, err := time.ParseDuration(variance)
		if err != nil {
			return err
		}
		if timeCheck.Before(t.Add(-v)) || timeCheck.After(t.Add(v)) {
			return fmt.Errorf("expected time to be within %s of %s, got %s", variance, t.Format(time.RFC3339), s)
		}
		return nil
	}
}

func TestAccMultirotateSet(t *testing.T) {
	n := time.Now().Round(time.Second)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMultirotateSetResourceConfig(n),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("multirotate_set.test", "rotation_set.0.expiration", timeCheckWithVariance(n.Add(time.Hour), "5s")),
					resource.TestCheckResourceAttrWith("multirotate_set.test", "rotation_set.1.expiration", timeCheckWithVariance(n.Add(time.Hour*2), "5s")),
					resource.TestCheckResourceAttr("multirotate_set.test", "current_rotation", "1"),
				),
			},
			// Update and Read testing
			{
				Config: testAccMultirotateSetResourceConfig(n.Add(time.Hour + time.Minute)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("multirotate_set.test", "rotation_set.0.expiration", timeCheckWithVariance(n.Add(time.Hour*3), "5s")),
					resource.TestCheckResourceAttrWith("multirotate_set.test", "rotation_set.1.expiration", timeCheckWithVariance(n.Add(time.Hour*2), "5s")),
					resource.TestCheckResourceAttr("multirotate_set.test", "current_rotation", "0"),
				),
			},
			{
				Config: testAccMultirotateSetResourceConfig(n.Add(time.Hour*2 + time.Minute)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("multirotate_set.test", "rotation_set.0.expiration", timeCheckWithVariance(n.Add(time.Hour*3), "5s")),
					resource.TestCheckResourceAttrWith("multirotate_set.test", "rotation_set.1.expiration", timeCheckWithVariance(n.Add(time.Hour*4), "5s")),
					resource.TestCheckResourceAttr("multirotate_set.test", "current_rotation", "1"),
				),
			},
		},
	})
}

func TestAccMultirotateSetPlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
resource "multirotate_set" "test" {
rotation_period = "1h"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						delayCheck{2 * time.Second},
					},
				},
			},
			{
				Config: `
resource "multirotate_set" "test" {
rotation_period = "1h"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						delayCheck{2 * time.Second},
					},
				},
			},
		},
	})
}

func testAccMultirotateSetResourceConfig(t time.Time) string {
	return fmt.Sprintf(`
provider "multirotate" {
  timestamp = %q
}

resource "multirotate_set" "test" {
  rotation_period = "1h"
}
`, t.Format(time.RFC3339))
}

type delayCheck struct {
	Delay time.Duration
}

func (d delayCheck) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	time.Sleep(d.Delay)
}
