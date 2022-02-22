package provider

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestLoadPlanConfig(t *testing.T) {
	original, _ := os.Getwd()
	os.Chdir("test-fixtures")
	got, err := LoadPlanConfig()
	os.Chdir(original)
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
