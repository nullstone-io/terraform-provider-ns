package provider

import (
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanConfigFromFile(t *testing.T) {
	got, err := PlanConfigFromFile(filepath.Join("test-fixtures", ".nullstone.json"))
	require.NoError(t, err, "unexpected error")
	want := PlanConfig{
		WorkspaceTarget: types.WorkspaceTarget{
			OrgName: "nullstone",
			StackName: "demo",
			EnvName:   "dev",
			BlockName: "fargate0",
		},
	}
	assert.Equal(t, want, got)
}
