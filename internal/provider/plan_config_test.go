package provider

import (
	"path/filepath"
	"testing"

	"github.com/nullstone-io/terraform-provider-ns/ns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanConfigFromFile(t *testing.T) {
	got, err := PlanConfigFromFile(filepath.Join("test-fixtures", ".nullstone.json"))
	require.NoError(t, err, "unexpected error")
	want := PlanConfig{
		Org: "nullstone",
		WorkspaceLocation: ns.WorkspaceLocation{
			Stack: "demo",
			Env:   "dev",
			Block: "fargate0",
		},
	}
	assert.Equal(t, want, got)
}
