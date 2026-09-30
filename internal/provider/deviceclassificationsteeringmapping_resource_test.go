package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDeviceClassificationSteeringMapping_basic tests steering mapping CRUD.
//
// Requires NETSKOPE_TEST_STEERING_ID to be set to an existing NPA steering
// config ID in the test tenant. Without it the test is skipped.
func TestAccDeviceClassificationSteeringMapping_basic(t *testing.T) {
	steeringID := os.Getenv("NETSKOPE_TEST_STEERING_ID")
	if steeringID == "" {
		t.Skip("Skipping: NETSKOPE_TEST_STEERING_ID not set — requires an existing NPA steering config ID")
	}

	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_device_classification_steering_mapping.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDCSteeringMappingConfig(rName, steeringID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "steering_id"),
					resource.TestCheckResourceAttr(resourceName, "steering_id", steeringID),
					resource.TestCheckResourceAttrSet(resourceName, "data.#"),
					resource.TestCheckResourceAttr(resourceName, "status", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// onprem_detection_ids is not returned by the GET endpoint; the GET response
					// returns `data` (the same IDs) which is the computed representation.
					"onprem_detection_ids",
				},
			},
		},
	})
}

func testAccDCSteeringMappingConfig(name, steeringID string) string {
	return fmt.Sprintf(`
%s

resource "netskope_device_classification_on_prem_detection" "test" {
  name = %q

  config = jsonencode({
    "onpremcheck" = {
      "onprem_use_dns"                     = "1"
      "onprem_host"                        = "internal.example.com"
      "onprem_ip"                          = "10.0.0.1"
      "onprem_http_host"                   = ""
      "onprem_http_tcp_connection_timeout" = "10"
      "onprem_additional_ips"              = []
      "onprem_additional_http_hosts"       = []
      "onprem_detection_method"            = "1"
      "onprem_egress_ips"                  = []
    }
  })
}

resource "netskope_device_classification_steering_mapping" "test" {
  steering_id          = %s
  onprem_detection_ids = [netskope_device_classification_on_prem_detection.test.onprem_id]
}
`, testAccProviderConfig(), name, steeringID)
}
