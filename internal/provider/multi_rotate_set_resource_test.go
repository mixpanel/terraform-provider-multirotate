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
			},
			// Update and Read testing
			{
				Config: testAccMultirotateSetResourceConfig(n.Add(time.Hour + time.Minute)),
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
