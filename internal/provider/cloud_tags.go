package provider

import (
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Cloud tag/label constraints.
//
// AWS: https://docs.aws.amazon.com/tag-editor/latest/userguide/tagging.html
//   - key <= 128 chars, value <= 256 chars
//   - allowed chars: letters, numbers, spaces, and + - = . _ : / @
//   - the "aws:" prefix is reserved and must never be emitted
//
// GCP: https://docs.cloud.google.com/compute/docs/labeling-resources#requirements
//   - keys 1-63 chars, values 0-63 chars
//   - only lowercase letters, digits, underscores, dashes (international letters allowed)
//   - keys must start with a lowercase letter
//
// Azure: https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources
//   - tag name <= 512 chars, value <= 256 chars
//   - names may not contain < > % & \ ? /
const (
	awsTagKeyMaxLen     = 128
	awsTagValueMaxLen   = 256
	gcpLabelMaxLen      = 63
	k8sLabelMaxLen      = 63
	azureTagKeyMaxLen   = 512
	azureTagValueMaxLen = 256
)

// labelSource is the cloud-agnostic set of workspace values that get formatted
// into per-cloud tags/labels. Keys are assigned per-cloud; values are sanitized
// per-cloud.
type labelSource struct {
	stackName          string
	envName            string
	blockName          string
	blockRef           string
	orgName            string
	dataClassification string
}

// labelEntry pairs an AWS (PascalCase) key and a GCP (lowercase) key with the
// underlying value.
type labelEntry struct {
	awsKey string
	gcpKey string
	value  string
}

// entries returns the logical label set in a stable order. Entries with an empty
// value (e.g. dataClassification on an unclassified workspace) are skipped by
// the builders.
func (s labelSource) entries() []labelEntry {
	return []labelEntry{
		{"Stack", "stack", s.stackName},
		{"Env", "env", s.envName},
		{"Environment", "environment", s.envName},
		{"Block", "block", s.blockName},
		{"Owner", "owner", s.orgName},
		{"Project", "project", s.stackName},
		{"DataClassification", "dataclassification", s.dataClassification},
		{"Application", "application", s.blockName},
		{"Component", "component", s.blockName},
	}
}

// buildAwsTags formats the workspace values as AWS tags (PascalCase keys, lightly
// sanitized values). Empty values are omitted.
func buildAwsTags(s labelSource) map[string]string {
	out := map[string]string{}
	for _, e := range s.entries() {
		if e.value == "" {
			continue
		}
		key := sanitizeAwsKey(e.awsKey)
		if key == "" || strings.HasPrefix(strings.ToLower(key), "aws:") {
			continue
		}
		out[key] = sanitizeAwsValue(e.value)
	}
	return out
}

// buildGcpLabels formats the workspace values as GCP labels (lowercase keys and
// values, fully sanitized). Empty values are omitted.
func buildGcpLabels(s labelSource) map[string]string {
	out := map[string]string{}
	for _, e := range s.entries() {
		if e.value == "" {
			continue
		}
		key := sanitizeGcpLabel(e.gcpKey, true)
		if key == "" {
			continue
		}
		out[key] = sanitizeGcpLabel(e.value, false)
	}
	return out
}

// buildAzureTags formats the workspace values as Azure tags (PascalCase keys,
// sanitized values). Azure tag names may not contain < > % & \ ? / and are
// limited to 512 chars; values are limited to 256 chars. Empty values are
// omitted. Owner is not classification — it is the org name (matching aws_tags).
func buildAzureTags(s labelSource) map[string]string {
	out := map[string]string{}
	for _, e := range s.entries() {
		if e.value == "" {
			continue
		}
		key := sanitizeAzureKey(e.awsKey)
		if key == "" {
			continue
		}
		out[key] = sanitizeAzureValue(e.value)
	}
	return out
}

// isAzureKeyChar reports whether r is allowed in an Azure tag name.
func isAzureKeyChar(r rune) bool {
	switch r {
	case '<', '>', '%', '&', '\\', '?', '/':
		return false
	}
	return true
}

func sanitizeAzureKey(k string) string {
	return truncateRunes(filterRunes(k, isAzureKeyChar), azureTagKeyMaxLen)
}

func sanitizeAzureValue(v string) string {
	return truncateRunes(v, azureTagValueMaxLen)
}

// isAwsTagChar reports whether r is allowed in an AWS tag key or value.
func isAwsTagChar(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}
	switch r {
	case ' ', '+', '-', '=', '.', '_', ':', '/', '@':
		return true
	}
	return false
}

func sanitizeAwsKey(k string) string {
	return truncateRunes(filterRunes(k, isAwsTagChar), awsTagKeyMaxLen)
}

func sanitizeAwsValue(v string) string {
	return truncateRunes(filterRunes(v, isAwsTagChar), awsTagValueMaxLen)
}

// sanitizeGcpLabel coerces v into a valid GCP label key or value:
//   - lowercases the value;
//   - replaces any disallowed char with a dash;
//   - truncates to 63 chars;
//   - for keys, strips leading chars until the label starts with a letter.
func sanitizeGcpLabel(v string, isKey bool) string {
	var b strings.Builder
	for _, r := range strings.ToLower(v) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		case unicode.IsLetter(r):
			// International letters are permitted by GCP (UTF-8).
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := truncateRunes(b.String(), gcpLabelMaxLen)
	if isKey {
		out = ensureGcpKeyStart(out)
	}
	return out
}

// ensureGcpKeyStart strips leading characters until the key begins with a letter,
// as required for GCP label keys. Returns "" if no letter is present.
func ensureGcpKeyStart(s string) string {
	for i, r := range s {
		if unicode.IsLetter(r) {
			return s[i:]
		}
	}
	return ""
}

func filterRunes(s string, keep func(rune) bool) string {
	var b strings.Builder
	for _, r := range s {
		if keep(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

// buildK8sLabels formats the workspace values as the recommended Kubernetes
// labels (https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/)
// plus nullstone.io/* labels. Keys are fixed, valid label keys; values are
// sanitized to Kubernetes' value rules. Empty values are omitted.
func buildK8sLabels(s labelSource) map[string]string {
	type kv struct{ key, value string }
	pairs := []kv{
		{"app.kubernetes.io/name", s.blockName},
		// version and component are intentionally left blank here; the consuming
		// module populates them (e.g. via merge). Empty values are omitted below.
		{"app.kubernetes.io/version", ""},
		{"app.kubernetes.io/component", ""},
		{"app.kubernetes.io/part-of", s.stackName},
		{"app.kubernetes.io/managed-by", "nullstone"},
		{"nullstone.io/block", s.blockName},
		{"nullstone.io/stack", s.stackName},
		{"nullstone.io/env", s.envName},
		{"nullstone.io/block-ref", s.blockRef},
		{"nullstone.io/data-classification", s.dataClassification},
	}

	out := map[string]string{}
	for _, p := range pairs {
		v := sanitizeK8sLabelValue(p.value)
		if v == "" {
			continue
		}
		out[p.key] = v
	}
	return out
}

// sanitizeK8sLabelValue coerces v into a valid Kubernetes label value:
//   - replaces any char outside [A-Za-z0-9_.-] with a dash;
//   - truncates to 63 chars;
//   - trims leading/trailing chars until it begins and ends with an alphanumeric.
//
// An empty input (or a value that reduces to nothing) returns "".
func sanitizeK8sLabelValue(v string) string {
	if v == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return trimToAlphanumeric(truncateRunes(b.String(), k8sLabelMaxLen))
}

func isAsciiAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// trimToAlphanumeric strips leading and trailing characters until the string
// begins and ends with an alphanumeric, as required for Kubernetes label values.
func trimToAlphanumeric(s string) string {
	runes := []rune(s)
	start := 0
	for start < len(runes) && !isAsciiAlphanumeric(runes[start]) {
		start++
	}
	end := len(runes)
	for end > start && !isAsciiAlphanumeric(runes[end-1]) {
		end--
	}
	return string(runes[start:end])
}

// toTfStringMap converts a Go string map into a tftypes string-valued map.
func toTfStringMap(m map[string]string) map[string]tftypes.Value {
	out := make(map[string]tftypes.Value, len(m))
	for k, v := range m {
		out[k] = tftypes.NewValue(tftypes.String, v)
	}
	return out
}
