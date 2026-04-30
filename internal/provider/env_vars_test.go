package provider

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMixedEnvVars_Interpolate(t *testing.T) {
	tests := []struct {
		name                  string
		inputEnvVars          map[string]string
		inputSecrets          map[string]string
		wantEnvVars           map[string]string
		wantSecrets           map[string]string
		wantSecretRefs        map[string]string
		wantFieldRefs         map[string]FieldRef
		wantConfigMapRefs     map[string]ConfigMapRef
		wantResourceFieldRefs map[string]ResourceFieldRef
		wantFileKeyRefs       map[string]FileKeyRef
		wantSecretKeys        []string
		wantErrors            int
	}{
		{
			name: "full interpolation with all ref types",
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
				"IP":                 "{{ k8s.field(v1, status.podIP) }}",
				"CM":                 "{{ k8s.configMap(my-key, my-cm, true) }}",
				"CM2":                "{{ k8s.configMap(key, name) }}",
				"RES":                "{{ k8s.resourceField(limits.cpu) }}",
				"RES2":               "{{ k8s.resourceField(limits.memory, main, 1Mi) }}",
				"FK":                 "{{ k8s.fileKey(MY_VAR, config.env, config-vol) }}",
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
			wantFieldRefs: map[string]FieldRef{
				"IP": {ApiVersion: "v1", FieldPath: "status.podIP"},
			},
			wantConfigMapRefs: map[string]ConfigMapRef{
				"CM":  {Key: "my-key", Name: "my-cm", Optional: true},
				"CM2": {Key: "key", Name: "name", Optional: false},
			},
			wantResourceFieldRefs: map[string]ResourceFieldRef{
				"RES":  {Resource: "limits.cpu"},
				"RES2": {Resource: "limits.memory", Container: "main", Divisor: "1Mi"},
			},
			wantFileKeyRefs: map[string]FileKeyRef{
				"FK": {Key: "MY_VAR", Path: "config.env", VolumeName: "config-vol"},
			},
			wantSecretKeys: []string{
				"DATABASE_URL",
				"DUPLICATE_TEST",
				"POSTGRES_URL",
				"SECRET_KEY_BASE",
			},
		},
		{
			name: "3-hop chain promotes to secret",
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
			wantSecretRefs:        map[string]string{},
			wantFieldRefs:         map[string]FieldRef{},
			wantConfigMapRefs:     map[string]ConfigMapRef{},
			wantResourceFieldRefs: map[string]ResourceFieldRef{},
			wantFileKeyRefs:       map[string]FileKeyRef{},
			wantSecretKeys:        []string{"A", "B", "DATABASE_URL"},
		},
		{
			name: "unknown k8s template passes through",
			inputEnvVars: map[string]string{
				"FOO": "bar",
				"UNK": "{{ k8s.future(a,b) }}",
			},
			inputSecrets:          map[string]string{},
			wantEnvVars:           map[string]string{"FOO": "bar", "UNK": "{{ k8s.future(a,b) }}"},
			wantSecrets:           map[string]string{},
			wantSecretRefs:        map[string]string{},
			wantFieldRefs:         map[string]FieldRef{},
			wantConfigMapRefs:     map[string]ConfigMapRef{},
			wantResourceFieldRefs: map[string]ResourceFieldRef{},
			wantFileKeyRefs:       map[string]FileKeyRef{},
			wantSecretKeys:        []string{},
		},
		{
			name: "malformed k8s.field wrong arg count returns error",
			inputEnvVars: map[string]string{
				"BAD": "{{ k8s.field(only-one) }}",
			},
			inputSecrets: map[string]string{},
			wantErrors:   1,
		},
		{
			name: "malformed k8s.configMap returns error",
			inputEnvVars: map[string]string{
				"BAD": "{{ k8s.configMap(only-one) }}",
			},
			inputSecrets: map[string]string{},
			wantErrors:   1,
		},
		{
			name: "malformed k8s.fileKey returns error",
			inputEnvVars: map[string]string{
				"BAD": "{{ k8s.fileKey(one, two) }}",
			},
			inputSecrets: map[string]string{},
			wantErrors:   1,
		},
		{
			// Reproduces a bug where Go's regex.ReplaceAllString interprets `$` in
			// the replacement (e.g., `$1`, `$abc`) as capture-group references,
			// clipping the value. Both secret-onto-envvar (step 2) and
			// envvar-onto-envvar/secret (step 3) must preserve `$` literally.
			name: "values containing $ are preserved during interpolation",
			inputEnvVars: map[string]string{
				"DATABASE_URL":   "{{ POSTGRES_PASSWORD }}",
				"API_BASE":       "prefix/{{ API_KEY }}/suffix",
				"EV_WITH_DOLLAR": "abc$xyz$1",
				"USES_EV":        "got:{{ EV_WITH_DOLLAR }}",
			},
			inputSecrets: map[string]string{
				"POSTGRES_PASSWORD": "p$ssw0rd$abc",
				"API_KEY":           "key$1$2",
			},
			wantEnvVars: map[string]string{
				"EV_WITH_DOLLAR": "abc$xyz$1",
				"USES_EV":        "got:abc$xyz$1",
			},
			wantSecrets: map[string]string{
				"POSTGRES_PASSWORD": "p$ssw0rd$abc",
				"DATABASE_URL":      "p$ssw0rd$abc",
				"API_KEY":           "key$1$2",
				"API_BASE":          "prefix/key$1$2/suffix",
			},
			wantSecretRefs:        map[string]string{},
			wantFieldRefs:         map[string]FieldRef{},
			wantConfigMapRefs:     map[string]ConfigMapRef{},
			wantResourceFieldRefs: map[string]ResourceFieldRef{},
			wantFileKeyRefs:       map[string]FileKeyRef{},
			wantSecretKeys:        []string{"API_BASE", "API_KEY", "DATABASE_URL", "POSTGRES_PASSWORD"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewEnvVars(test.inputEnvVars, test.inputSecrets)
			errs := got.Interpolate()

			if test.wantErrors > 0 {
				if len(errs) != test.wantErrors {
					t.Errorf("expected %d errors, got %d: %v", test.wantErrors, len(errs), errs)
				}
				return
			}
			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if diff := cmp.Diff(test.wantEnvVars, got.EnvVars()); diff != "" {
				t.Errorf("mismatched env vars (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantSecrets, got.Secrets()); diff != "" {
				t.Errorf("mismatched secrets (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantSecretRefs, got.SecretRefs()); diff != "" {
				t.Errorf("mismatched secret refs (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantFieldRefs, got.FieldRefs()); diff != "" {
				t.Errorf("mismatched field refs (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantConfigMapRefs, got.ConfigMapRefs()); diff != "" {
				t.Errorf("mismatched config map refs (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantResourceFieldRefs, got.ResourceFieldRefs()); diff != "" {
				t.Errorf("mismatched resource field refs (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantFileKeyRefs, got.FileKeyRefs()); diff != "" {
				t.Errorf("mismatched file key refs (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantSecretKeys, got.SecretKeys()); diff != "" {
				t.Errorf("mismatched secret keys (-want, +got):\n%s", diff)
			}
		})
	}
}
