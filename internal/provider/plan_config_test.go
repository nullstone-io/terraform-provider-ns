package provider

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestPlanConfigFromFile(t *testing.T) {
	got, err := PlanConfigFromFile(filepath.Join("test-fixtures", ".nullstone.json"))
	require.NoError(t, err, "unexpected error")
	want := PlanConfig{
		OrgName:   "nullstone",
		StackId:   100,
		StackName: "demo",
		BlockId:   101,
		BlockName: "fargate0",
		BlockRef:  "yellow-giraffe",
		EnvId:     102,
		EnvName:   "dev",
	}
	assert.Equal(t, want, got)
}
