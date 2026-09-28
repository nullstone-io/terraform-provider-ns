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
			Platform: platformdata.PlatformAwsEcs,
			Standard: map[string]string{"A": "standard", "B": "standard", "C": "standard", "D": "standard", "E": "standard"},
			Cloud:    map[string]string{"B": "cloud", "C": "cloud", "D": "cloud", "E": "cloud"},
			Otel:     map[string]string{"C": "otel", "D": "otel", "E": "otel"},
			CapabilityEnv: []capabilityEntry{
				{Capability: "cap_pg", Name: "D", Value: "capability"},
				{Capability: "cap_pg", Name: "E", Value: "capability"},
			},
			CapabilityPrefixes: map[string]string{"cap_pg": ""},
			UserEnv:            map[string]string{"E": "user-{{ A }}"},
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
			Platform: platformdata.PlatformAwsEcs,
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
		assert.Equal(t, platformdata.PlatformAwsEcs, record.Platform)
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
			Platform: platformdata.PlatformGcpGke,
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
			Platform: platformdata.PlatformAwsEcs,
			UserEnv:  map[string]string{"POD_IP": "{{ k8s.field(v1, status.podIP) }}"},
		})
		require.Len(t, diags, 1)
		assert.Contains(t, diags[0].Summary, "POD_IP")
	})

	t.Run("secret ref rejected on platform without secret refs", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{
			Platform: platformdata.PlatformAwsS3,
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
			Platform: platformdata.PlatformAwsEcs,
			CapabilityEnv: []capabilityEntry{
				{Capability: "cap_a", Name: "HOST", Value: "a"},
				{Capability: "cap_b", Name: "HOST", Value: "b"},
			},
			CapabilityPrefixes: map[string]string{"cap_a": "", "cap_b": ""},
		})
		require.Len(t, diags, 1)
		assert.Equal(t, "Duplicate capability environment variable: HOST", diags[0].Summary)
	})

	t.Run("invalid key", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{
			Platform: platformdata.PlatformAwsEcs,
			UserEnv:  map[string]string{"BAD-KEY": "x"},
		})
		require.Len(t, diags, 1)
		assert.Equal(t, "Invalid environment variable key: BAD-KEY", diags[0].Summary)
	})
}

// D16: secrets always win; capability entries must be named and prefixed; secrets cannot carry ref templates.
func TestResolveLayers_D16(t *testing.T) {
	t.Run("secret wins over a later plain layer", func(t *testing.T) {
		in := layeredEnvInput{
			Platform: platformdata.PlatformAwsEcs,
			CapabilitySecrets: []capabilityEntry{
				{Capability: "postgres0", Name: "PASSWORD", Value: "hunter2"},
			},
			CapabilityPrefixes: map[string]string{"postgres0": "PG_"},
			UserEnv:            map[string]string{"PG_PASSWORD": "plaintext-override"},
		}
		result, diags := resolveLayers(in)
		require.Empty(t, diags)

		assert.Equal(t, []string{"PG_PASSWORD"}, result.Managed)
		assert.Empty(t, result.Plain)
		assert.Equal(t, "hunter2", result.ManagedValues()["PG_PASSWORD"])
		assert.Equal(t, layeredEnvKey{Source: platformdata.SourceCapability, Capability: "postgres0", Template: "hunter2", SecretInput: true}, result.Keys["PG_PASSWORD"])

		raw, err := result.PlatformData()
		require.NoError(t, err)
		assert.NotContains(t, raw, "plaintext-override")
		record, err := platformdata.ParseEnvV1(json.RawMessage(raw))
		require.NoError(t, err)
		assert.Equal(t, platformdata.EnvV1Variable{Sensitive: true, Source: platformdata.SourceCapability, Capability: "postgres0"}, record.Variables["PG_PASSWORD"])
	})

	t.Run("secret layers still override each other in order", func(t *testing.T) {
		in := layeredEnvInput{
			Platform: platformdata.PlatformAwsEcs,
			CapabilitySecrets: []capabilityEntry{
				{Capability: "postgres0", Name: "PASSWORD", Value: "from-cap"},
			},
			CapabilityPrefixes: map[string]string{"postgres0": "PG_"},
			UserSecrets:        map[string]string{"PG_PASSWORD": "from-user"},
		}
		result, diags := resolveLayers(in)
		require.Empty(t, diags)
		assert.Equal(t, "from-user", result.ManagedValues()["PG_PASSWORD"])
		assert.Equal(t, platformdata.SourceUser, result.Sources()["PG_PASSWORD"])
	})

	t.Run("secret wins in key-only layout shape", func(t *testing.T) {
		// ns_env_layout feeds "" for every secret; a plain user_env at the same key must still be ignored.
		in := layeredEnvInput{
			Platform:    platformdata.PlatformAwsEcs,
			UserEnv:     map[string]string{"API_TOKEN": "{{ secret(arn:aws:secretsmanager:us-east-1:0123456789012:secret:x) }}"},
			UserSecrets: map[string]string{"API_TOKEN": ""},
		}
		result, diags := resolveLayers(in)
		require.Empty(t, diags)
		assert.Equal(t, []string{"API_TOKEN"}, result.Managed)
		assert.Empty(t, result.Unmanaged)
	})

	t.Run("capability missing from capability_prefixes", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{
			Platform: platformdata.PlatformAwsEcs,
			CapabilityEnv: []capabilityEntry{
				{Capability: "postgres0", Name: "HOST", Value: "db"},
			},
		})
		require.Len(t, diags, 1)
		assert.Equal(t, "Unknown capability: postgres0", diags[0].Summary)
		assert.Contains(t, diags[0].Detail, "capability_env")
		assert.Contains(t, diags[0].Detail, "capability_prefixes")
	})

	t.Run("explicit empty prefix is fine", func(t *testing.T) {
		result, diags := resolveLayers(layeredEnvInput{
			Platform: platformdata.PlatformAwsEcs,
			CapabilityEnv: []capabilityEntry{
				{Capability: "postgres0", Name: "HOST", Value: "db"},
			},
			CapabilityPrefixes: map[string]string{"postgres0": ""},
		})
		require.Empty(t, diags)
		assert.Equal(t, []string{"HOST"}, result.Plain)
	})

	t.Run("empty capability name", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{
			Platform: platformdata.PlatformAwsEcs,
			CapabilitySecrets: []capabilityEntry{
				{Capability: "", Name: "PASSWORD", Value: "x"},
			},
			CapabilityPrefixes: map[string]string{"": "PG_"},
		})
		require.Len(t, diags, 1)
		assert.Equal(t, "Capability name is required", diags[0].Summary)
		assert.Contains(t, diags[0].Detail, "capability_secrets")
	})

	t.Run("secret ref template inside a secrets input", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{
			Platform:    platformdata.PlatformAwsEcs,
			UserSecrets: map[string]string{"DB_PASSWORD": "{{ secret(arn:aws:secretsmanager:us-east-1:0123456789012:secret:x) }}"},
		})
		require.Len(t, diags, 1)
		assert.Equal(t, "Invalid secret template: DB_PASSWORD", diags[0].Summary)
		assert.Contains(t, diags[0].Detail, "env_vars")
	})

	t.Run("k8s template inside a capability secret", func(t *testing.T) {
		_, diags := resolveLayers(layeredEnvInput{
			Platform: platformdata.PlatformGcpGke,
			CapabilitySecrets: []capabilityEntry{
				{Capability: "postgres0", Name: "PASSWORD", Value: "{{ k8s.field(v1, status.podIP) }}"},
			},
			CapabilityPrefixes: map[string]string{"postgres0": "PG_"},
		})
		require.Len(t, diags, 1)
		assert.Equal(t, "Invalid secret template: PG_PASSWORD", diags[0].Summary)
	})
}

func TestHasRuntimeRefTemplate(t *testing.T) {
	assert.True(t, hasRuntimeRefTemplate("{{ secret(arn) }}"))
	assert.True(t, hasRuntimeRefTemplate("{{secret(arn)}}"))
	assert.True(t, hasRuntimeRefTemplate("{{ k8s.field(v1, status.podIP) }}"))
	assert.True(t, hasRuntimeRefTemplate("{{ k8s.configMap(key, name) }}"))
	assert.True(t, hasRuntimeRefTemplate("{{ k8s.resourceField(limits.cpu) }}"))
	assert.True(t, hasRuntimeRefTemplate("{{ k8s.fileKey(key, path, vol) }}"))
	assert.False(t, hasRuntimeRefTemplate("{{ OTHER_VAR }}"))
	assert.False(t, hasRuntimeRefTemplate("plain"))
}
