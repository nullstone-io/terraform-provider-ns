package ns

import (
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var testAccProviders = map[string]terraform.ResourceProvider{
	"ns": Provider().(*schema.Provider),
}

func TestProvider(t *testing.T) {
	t.Run("runs internal validation without error", func(t *testing.T) {
		if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
			t.Fatalf("err: %s", err)
		}
	})
}

func testAccPreCheck(t *testing.T) {
	// NOTE: Add prechecks for all tests here
}
