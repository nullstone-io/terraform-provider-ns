package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerFactories = map[string]func() (*schema.Provider, error){
	"ns": func() (*schema.Provider, error) {
		return New("test")(), nil
	},
}

func TestProvider(t *testing.T) {
	t.Run("runs internal validation without error", func(t *testing.T) {
		if err := New("test")().InternalValidate(); err != nil {
			t.Fatalf("err: %s", err)
		}
	})
}

func testAccPreCheck(t *testing.T) {
	// NOTE: Add prechecks for all tests here
}
