package provider

import (
	"strings"
	"testing"
)

func TestBuildAwsTags(t *testing.T) {
	t.Run("full key set", func(t *testing.T) {
		got := buildAwsTags(labelSource{
			stackName:          "stack0",
			envName:            "env0",
			blockName:          "block0",
			orgName:            "org0",
			dataClassification: "2-customer-content",
		})
		want := map[string]string{
			"Stack":              "stack0",
			"Env":                "env0",
			"Environment":        "env0",
			"Block":              "block0",
			"Owner":              "org0",
			"Project":            "stack0",
			"DataClassification": "2-customer-content",
			"Application":        "block0",
			"Component":          "block0",
		}
		assertStringMapEqual(t, want, got)
	})

	t.Run("dataclassification omitted when empty", func(t *testing.T) {
		got := buildAwsTags(labelSource{
			stackName: "stack0",
			envName:   "env0",
			blockName: "block0",
			orgName:   "org0",
		})
		if _, ok := got["DataClassification"]; ok {
			t.Fatalf("expected DataClassification to be omitted, got %q", got["DataClassification"])
		}
		if len(got) != 8 {
			t.Fatalf("expected 8 tags, got %d: %v", len(got), got)
		}
	})

	t.Run("value length is truncated to 256", func(t *testing.T) {
		long := strings.Repeat("a", 300)
		got := buildAwsTags(labelSource{stackName: long, envName: "env0", blockName: "block0", orgName: "org0"})
		if l := len([]rune(got["Stack"])); l != awsTagValueMaxLen {
			t.Fatalf("expected value truncated to %d runes, got %d", awsTagValueMaxLen, l)
		}
	})

	t.Run("disallowed value chars are dropped", func(t *testing.T) {
		// commas and parentheses are not in the AWS allowlist; spaces and dots are.
		got := buildAwsTags(labelSource{stackName: "s", envName: "e", blockName: "b", orgName: "Acme, Inc. (US)"})
		if got["Owner"] != "Acme Inc. US" {
			t.Fatalf("unexpected sanitized owner: %q", got["Owner"])
		}
	})
}

func TestBuildGcpLabels(t *testing.T) {
	t.Run("full key set", func(t *testing.T) {
		got := buildGcpLabels(labelSource{
			stackName:          "stack0",
			envName:            "env0",
			blockName:          "block0",
			orgName:            "org0",
			dataClassification: "2-customer-content",
		})
		want := map[string]string{
			"stack":              "stack0",
			"env":                "env0",
			"environment":        "env0",
			"block":              "block0",
			"owner":              "org0",
			"project":            "stack0",
			"dataclassification": "2-customer-content",
			"application":        "block0",
			"component":          "block0",
		}
		assertStringMapEqual(t, want, got)
	})

	t.Run("adversarial org name is sanitized", func(t *testing.T) {
		got := buildGcpLabels(labelSource{stackName: "s", envName: "e", blockName: "b", orgName: "Acme Corp, Inc. @HQ"})
		owner := got["owner"]
		assertValidGcpValue(t, owner)
		if owner != "acme-corp--inc---hq" {
			t.Fatalf("unexpected sanitized owner: %q", owner)
		}
	})

	t.Run("dataclassification omitted when empty", func(t *testing.T) {
		got := buildGcpLabels(labelSource{stackName: "stack0", envName: "env0", blockName: "block0", orgName: "org0"})
		if _, ok := got["dataclassification"]; ok {
			t.Fatalf("expected dataclassification to be omitted")
		}
	})
}

func TestSanitizeGcpLabel(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		isKey bool
		want  string
	}{
		{"lowercases", "MyValue", false, "myvalue"},
		{"replaces spaces and punctuation", "a.b/c@d e", false, "a-b-c-d-e"},
		{"keeps digits dashes underscores", "a-b_c1", false, "a-b_c1"},
		{"truncates value to 63", strings.Repeat("a", 80), false, strings.Repeat("a", 63)},
		{"key strips leading digit", "9stack", true, "stack"},
		{"key strips leading dash", "-stack", true, "stack"},
		{"key all non-letters becomes empty", "123", true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeGcpLabel(tt.in, tt.isKey)
			if got != tt.want {
				t.Fatalf("sanitizeGcpLabel(%q, %v) = %q, want %q", tt.in, tt.isKey, got, tt.want)
			}
			if got != "" {
				assertValidGcpValue(t, got)
			}
		})
	}
}

func TestSanitizeAws(t *testing.T) {
	t.Run("key truncated to 128", func(t *testing.T) {
		if l := len([]rune(sanitizeAwsKey(strings.Repeat("k", 200)))); l != awsTagKeyMaxLen {
			t.Fatalf("expected key truncated to %d, got %d", awsTagKeyMaxLen, l)
		}
	})
	t.Run("value truncated to 256", func(t *testing.T) {
		if l := len([]rune(sanitizeAwsValue(strings.Repeat("v", 400)))); l != awsTagValueMaxLen {
			t.Fatalf("expected value truncated to %d, got %d", awsTagValueMaxLen, l)
		}
	})
}

func TestBuildK8sLabels(t *testing.T) {
	src := labelSource{
		stackName: "stack0",
		envName:   "env0",
		blockName: "block0",
		blockRef:  "yellow-giraffe",
		orgName:   "org0",
	}

	t.Run("full label set omits blank labels", func(t *testing.T) {
		got := buildK8sLabels(src)
		want := map[string]string{
			"app.kubernetes.io/name":       "block0",
			"app.kubernetes.io/part-of":    "stack0",
			"app.kubernetes.io/managed-by": "nullstone",
			"nullstone.io/block":           "block0",
			"nullstone.io/stack":           "stack0",
			"nullstone.io/env":             "env0",
			"nullstone.io/block-ref":       "yellow-giraffe",
		}
		assertStringMapEqual(t, want, got)
		// version and component are intentionally blank, so they are omitted.
		if _, ok := got["app.kubernetes.io/version"]; ok {
			t.Fatalf("expected app.kubernetes.io/version to be omitted")
		}
		if _, ok := got["app.kubernetes.io/component"]; ok {
			t.Fatalf("expected app.kubernetes.io/component to be omitted")
		}
		// data-classification is blank in src, so it is omitted.
		if _, ok := got["nullstone.io/data-classification"]; ok {
			t.Fatalf("expected nullstone.io/data-classification to be omitted when empty")
		}
	})

	t.Run("data-classification label emitted when set", func(t *testing.T) {
		classified := src
		classified.dataClassification = "2-customer-content"
		got := buildK8sLabels(classified)
		if got["nullstone.io/data-classification"] != "2-customer-content" {
			t.Fatalf("expected nullstone.io/data-classification=2-customer-content, got %q", got["nullstone.io/data-classification"])
		}
	})
}

func TestBuildAzureTags(t *testing.T) {
	t.Run("full key set", func(t *testing.T) {
		got := buildAzureTags(labelSource{
			stackName:          "stack0",
			envName:            "env0",
			blockName:          "block0",
			orgName:            "org0",
			dataClassification: "2-customer-content",
		})
		want := map[string]string{
			"Stack":              "stack0",
			"Env":                "env0",
			"Environment":        "env0",
			"Block":              "block0",
			"Owner":              "org0",
			"Project":            "stack0",
			"DataClassification": "2-customer-content",
			"Application":        "block0",
			"Component":          "block0",
		}
		assertStringMapEqual(t, want, got)
	})

	t.Run("dataclassification omitted when empty", func(t *testing.T) {
		got := buildAzureTags(labelSource{stackName: "stack0", envName: "env0", blockName: "block0", orgName: "org0"})
		if _, ok := got["DataClassification"]; ok {
			t.Fatalf("expected DataClassification to be omitted")
		}
	})

	t.Run("disallowed key chars are stripped", func(t *testing.T) {
		// Azure tag names may not contain < > %% & \ ? /. The fixed keys here are
		// clean, so verify the sanitizer drops disallowed chars directly.
		if got := sanitizeAzureKey(`a<b>c%d&e\f?g/h`); got != "abcdefgh" {
			t.Fatalf("unexpected sanitized azure key: %q", got)
		}
	})

	t.Run("value truncated to 256", func(t *testing.T) {
		long := strings.Repeat("a", 400)
		got := buildAzureTags(labelSource{stackName: long, envName: "e", blockName: "b", orgName: "o"})
		if l := len([]rune(got["Stack"])); l != azureTagValueMaxLen {
			t.Fatalf("expected value truncated to %d, got %d", azureTagValueMaxLen, l)
		}
	})
}

func TestSanitizeK8sLabelValue(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty stays empty", "", ""},
		{"reference style passes through", "yellow-giraffe", "yellow-giraffe"},
		{"keeps mixed case dots underscores", "My.App_1", "My.App_1"},
		{"replaces spaces and punctuation", "a b/c@d", "a-b-c-d"},
		{"trims leading and trailing non-alphanumeric", "-_.value._-", "value"},
		{"truncates to 63", strings.Repeat("a", 80), strings.Repeat("a", 63)},
		{"truncation then trims trailing dash", strings.Repeat("a", 62) + "-bbb", strings.Repeat("a", 62)},
		{"all punctuation reduces to empty", "@@@", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeK8sLabelValue(tt.in)
			if got != tt.want {
				t.Fatalf("sanitizeK8sLabelValue(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if got != "" {
				assertValidK8sValue(t, got)
			}
		})
	}
}

func assertValidK8sValue(t *testing.T, s string) {
	t.Helper()
	if len([]rune(s)) > k8sLabelMaxLen {
		t.Fatalf("value %q exceeds %d chars", s, k8sLabelMaxLen)
	}
	runes := []rune(s)
	if !isAsciiAlphanumeric(runes[0]) || !isAsciiAlphanumeric(runes[len(runes)-1]) {
		t.Fatalf("value %q must begin and end with an alphanumeric", s)
	}
	for _, r := range runes {
		ok := isAsciiAlphanumeric(r) || r == '-' || r == '_' || r == '.'
		if !ok {
			t.Fatalf("value %q contains disallowed char %q", s, r)
		}
	}
}

func assertValidGcpValue(t *testing.T, s string) {
	t.Helper()
	if len([]rune(s)) > gcpLabelMaxLen {
		t.Fatalf("value %q exceeds %d chars", s, gcpLabelMaxLen)
	}
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-'
		if !ok {
			t.Fatalf("value %q contains disallowed char %q", s, r)
		}
	}
}

func assertStringMapEqual(t *testing.T, want, got map[string]string) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("map length mismatch: want %d (%v), got %d (%v)", len(want), want, len(got), got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("key %q: want %q, got %q", k, v, got[k])
		}
	}
}
