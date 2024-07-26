package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMultirotateSet(t *testing.T) {
	n := time.Now()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMultirotateSetResourceConfig(n),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("multirotate_set.test", "rotation_set.0.expiration", n.Add(time.Hour).Format(time.RFC3339)),
					resource.TestCheckResourceAttr("multirotate_set.test", "rotation_set.1.expiration", n.Add(time.Hour*2).Format(time.RFC3339)),
					resource.TestCheckResourceAttr("multirotate_set.test", "current_rotation", "1"),
				),
			},
			// Update and Read testing
			{
				Config: testAccMultirotateSetResourceConfig(n.Add(time.Hour + time.Minute)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("multirotate_set.test", "rotation_set.0.expiration", n.Add(time.Hour*3).Format(time.RFC3339)),
					resource.TestCheckResourceAttr("multirotate_set.test", "rotation_set.1.expiration", n.Add(time.Hour*2).Format(time.RFC3339)),
					resource.TestCheckResourceAttr("multirotate_set.test", "current_rotation", "0"),
				),
			},
			{
				Config: testAccMultirotateSetResourceConfig(n.Add(time.Hour*2 + time.Minute)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("multirotate_set.test", "rotation_set.0.expiration", n.Add(time.Hour*3).Format(time.RFC3339)),
					resource.TestCheckResourceAttr("multirotate_set.test", "rotation_set.1.expiration", n.Add(time.Hour*4).Format(time.RFC3339)),
					resource.TestCheckResourceAttr("multirotate_set.test", "current_rotation", "1"),
				),
			},
		},
	})
}

func testAccMultirotateSetResourceConfig(t time.Time) string {
	return fmt.Sprintf(`
resource "multirotate_set" "test" {
  rotation_period = "1h"
  timestamp = %q
}
`, t.Format(time.RFC3339))
}
