package provider

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMixedEnvVars_Interpolate(t *testing.T) {
	tests := []struct {
		inputEnvVars   map[string]string
		inputSecrets   map[string]string
		wantEnvVars    map[string]string
		wantSecrets    map[string]string
		wantSecretRefs map[string]string
		wantSecretKeys []string
	}{
		{
			inputEnvVars: map[string]string{
				"NULLSTONE_STACK":    "primary",
				"NULLSTONE_BLOCK":    "acme-api",
				"NULLSTONE_ENV":      "dev",
				"FEATURE_FLAG_0115":  "true",
				"ANOTHER_IDENTIFIER": "{{ IDENTIFIER }}",
				"DATABASE_URL":       "{{POSTGRES_URL}}",
				"IDENTIFIER":         "{{ NULLSTONE_STACK }}.{{ NULLSTONE_BLOCK }}.{{ NULLSTONE_ENV }}",
				"DUPLICATE_TEST":     "{{ SECRET_KEY_BASE }}/{{ POSTGRES_URL }}",
				"VAR_WITH_REF":       "{{ secret(arn:aws:something) }}",
			},
			inputSecrets: map[string]string{
				"POSTGRES_URL":    "fake-value1",
				"SECRET_KEY_BASE": "fake-value2",
			},
			wantEnvVars: map[string]string{
				"NULLSTONE_STACK":    "primary",
				"NULLSTONE_BLOCK":    "acme-api",
				"NULLSTONE_ENV":      "dev",
				"FEATURE_FLAG_0115":  "true",
				"IDENTIFIER":         "primary.acme-api.dev",
				"ANOTHER_IDENTIFIER": "primary.acme-api.dev",
			},
			wantSecrets: map[string]string{
				"DATABASE_URL":    "fake-value1",
				"POSTGRES_URL":    "fake-value1",
				"SECRET_KEY_BASE": "fake-value2",
				"DUPLICATE_TEST":  "fake-value2/fake-value1",
			},
			wantSecretRefs: map[string]string{
				"VAR_WITH_REF": "arn:aws:something",
			},
			wantSecretKeys: []string{
				"DATABASE_URL",
				"DUPLICATE_TEST",
				"POSTGRES_URL",
				"SECRET_KEY_BASE",
			},
		},
		{
			// 3-hop chain: A → B → DATABASE_URL (secret)
			// B gets promoted to secret in step 2; A must then also be promoted
			inputEnvVars: map[string]string{
				"A": "{{ B }}",
				"B": "{{ DATABASE_URL }}",
			},
			inputSecrets: map[string]string{
				"DATABASE_URL": "fake-db-url",
			},
			wantEnvVars: map[string]string{},
			wantSecrets: map[string]string{
				"A":            "fake-db-url",
				"B":            "fake-db-url",
				"DATABASE_URL": "fake-db-url",
			},
			wantSecretRefs: map[string]string{},
			wantSecretKeys: []string{"A", "B", "DATABASE_URL"},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := NewEnvVars(test.inputEnvVars, test.inputSecrets)
			got.Interpolate()
			if diff := cmp.Diff(test.wantEnvVars, got.EnvVars()); diff != "" {
				t.Errorf("mismatched env vars (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantSecrets, got.Secrets()); diff != "" {
				t.Errorf("mismatched secrets (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantSecretRefs, got.SecretRefs()); diff != "" {
				t.Errorf("mismatched secret refs (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantSecretKeys, got.SecretKeys()); diff != "" {
				t.Errorf("mismatched secret keys (-want, +got):\n%s", diff)
			}
		})
	}
}
