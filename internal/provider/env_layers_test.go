package provider

import (
	"encoding/json"
	"testing"

	"github.com/nullstone-io/module/platformdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLayers(t *testing.T) {
	t.Run("precedence: user > capability > otel > cloud > standard", func(t *testing.T) {
		in := layeredEnvInput{
			Platform: platformdata.PlatformEcs,
			Standard: map[string]string{"A": "standard", "B": "standard", "C": "standard", "D": "standard", "E": "standard"},
			Cloud:    map[string]string{"B": "cloud", "C": "cloud", "D": "cloud", "E": "cloud"},
			Otel:     map[string]string{"C": "otel", "D": "otel", "E": "otel"},
			CapabilityEnv: []capabilityEntry{
				{Capability: "cap_pg", Name: "D", Value: "capability"},
				{Capability: "cap_pg", Name: "E", Value: "capability"},
			},
			UserEnv: map[string]string{"E": "user-{{ A }}"},
		}
		result, diags := resolveLayers(in)
		require.Empty(t, diags)

		assert.Equal(t, map[string]string{"A": "standard", "B": "cloud", "C": "otel", "D": "capability", "E": "user-standard"}, result.PlainValues())
		assert.Equal(t, map[string]string{
			"A": platformdata.SourceStandard,
			"B": platformdata.SourceCloud,
			"C": platformdata.SourceOtel,
			"D": platformdata.SourceCapability,
			"E": platformdata.SourceUser,
		}, result.Sources())
		assert.Equal(t, map[string]string{"D": "cap_pg"}, result.Capabilities())
		// template is the raw input value from the winning layer
		assert.Equal(t, "user-{{ A }}", result.Keys["E"].Template)
		assert.Equal(t, "capability", result.Keys["D"].Template)
		assert.Equal(t, "cap_pg", result.Keys["D"].Capability)
		assert.Empty(t, result.Keys["E"].Capability)
	})

	t.Run("capability prefixes, promotion and classification", func(t *testing.T) {
		in := layeredEnvInput{
			Platform: platformdata.PlatformEcs,
			Standard: map[string]string{"NULLSTONE_ENV": "dev"},
			CapabilityEnv: []capabilityEntry{
				{Capability: "postgres0", Name: "HOST", Value: "db.internal"},
			},
			CapabilitySecrets: []capabilityEntry{
				{Capability: "postgres0", Name: "PASSWORD", Value: "hunter2"},
			},
			CapabilityPrefixes: map[string]string{"postgres0": "PG_"},
			UserEnv: map[string]string{
				"DATABASE_URL": "postgres://user:{{ PG_PASSWORD }}@{{ PG_HOST }}/db",
				"EXISTING":     "{{ secret(arn:aws:secretsmanager:us-east-1:0123456789012:secret:x) }}",
			},
			UserSecrets: map[string]string{"API_TOKEN": "tok"},
		}
		result, diags := resolveLayers(in)
		require.Empty(t, diags)

		assert.Equal(t, []string{"NULLSTONE_ENV", "PG_HOST"}, result.Plain)
		assert.Equal(t, []string{"API_TOKEN", "DATABASE_URL", "PG_PASSWORD"}, result.Managed)
		assert.Equal(t, []string{"EXISTING"}, result.Unmanaged)
		assert.Empty(t, result.K8s)
		assert.Equal(t, []string{"API_TOKEN", "DATABASE_URL", "EXISTING", "PG_PASSWORD"}, result.AllSecretKeys())
		assert.Equal(t, "postgres://user:hunter2@db.internal/db", result.ManagedValues()["DATABASE_URL"])
		assert.Equal(t, map[string]string{"EXISTING": "arn:aws:secretsmanager:us-east-1:0123456789012:secret:x"}, result.UnmanagedRefs())
		assert.Equal(t, map[string]string{"PG_HOST": "postgres0", "PG_PASSWORD": "postgres0"}, result.Capabilities())

		raw, err := result.PlatformData()
		require.NoError(t, err)
		assert.NotContains(t, raw, "hunter2")
		assert.NotContains(t, raw, "tok")
		record, err := platformdata.ParseEnvV1(json.RawMessage(raw))
		require.NoError(t, err)
		assert.Equal(t, platformdata.PlatformEcs, record.Platform)
		assert.Equal(t, platformdata.EnvV1Variable{Template: "db.internal", Value: "db.internal", Source: platformdata.SourceCapability, Capability: "postgres0"}, record.Variables["PG_HOST"])
		assert.Equal(t, platformdata.EnvV1Variable{Sensitive: true, Source: platformdata.SourceCapability, Capability: "postgres0"}, record.Variables["PG_PASSWORD"])
		assert.Equal(t, platformdata.EnvV1Variable{Template: "postgres://user:{{ PG_PASSWORD }}@{{ PG_HOST }}/db", Sensitive: true, Source: platformdata.SourceUser}, record.Variables["DATABASE_URL"])
		assert.Equal(t, platformdata.EnvV1Variable{
			Template:  "{{ secret(arn:aws:secretsmanager:us-east-1:0123456789012:secret:x) }}",
			Sensitive: true,
			Ref:       &platformdata.EnvV1Ref{Type: platformdata.RefTypeSecret, Id: "arn:aws:secretsmanager:us-east-1:0123456789012:secret:x"},
			Source:    platformdata.SourceUser,
		}, record.Variables["EXISTING"])
		assert.Equal(t, platformdata.EnvV1Variable{Sensitive: true, Source: platformdata.SourceUser}, record.Variables["API_TOKEN"])
	})

	t.Run("k8s refs on k8s platform", func(t *testing.T) {
		in := layeredEnvInput{
			Platform: platformdata.PlatformK8s,
			UserEnv:  map[string]string{"POD_IP": "{{ k8s.field(v1, status.podIP) }}"},
		}
		result, diags := resolveLayers(in)
		require.Empty(t, diags)
		assert.Equal(t, []string{"POD_IP"}, result.K8s)
		raw, err := result.PlatformData()
		require.NoError(t, err)
		record, err := platformdata.ParseEnvV1(json.RawMessage(raw))
		require.NoError(t, err)
		assert.Equal(t, &platformdata.EnvV1Ref{Type: platformdata.RefTypeK8sField, ApiVersion: "v1", FieldPath: "status.podIP"}, record.Variables["POD_IP"].Ref)
	})

	t.Run("k8s ref rejected on non-k8s platform", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{
			Platform: platformdata.PlatformEcs,
			UserEnv:  map[string]string{"POD_IP": "{{ k8s.field(v1, status.podIP) }}"},
		})
		require.Len(t, diags, 1)
		assert.Contains(t, diags[0].Summary, "POD_IP")
	})

	t.Run("secret ref rejected on platform without secret refs", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{
			Platform: platformdata.PlatformS3,
			UserEnv:  map[string]string{"X": "{{ secret(some-id) }}"},
		})
		require.Len(t, diags, 1)
		assert.Contains(t, diags[0].Summary, "X")
	})

	t.Run("unknown platform", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{Platform: "mainframe"})
		require.Len(t, diags, 1)
		assert.Equal(t, "Unknown platform: mainframe", diags[0].Summary)
	})

	t.Run("duplicate capability keys", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{
			Platform: platformdata.PlatformEcs,
			CapabilityEnv: []capabilityEntry{
				{Capability: "cap_a", Name: "HOST", Value: "a"},
				{Capability: "cap_b", Name: "HOST", Value: "b"},
			},
		})
		require.Len(t, diags, 1)
		assert.Equal(t, "Duplicate capability environment variable: HOST", diags[0].Summary)
	})

	t.Run("invalid key", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{
			Platform: platformdata.PlatformEcs,
			UserEnv:  map[string]string{"BAD-KEY": "x"},
		})
		require.Len(t, diags, 1)
		assert.Equal(t, "Invalid environment variable key: BAD-KEY", diags[0].Summary)
	})
}
