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
const (
	awsTagKeyMaxLen   = 128
	awsTagValueMaxLen = 256
	gcpLabelMaxLen    = 63
)

// labelSource is the cloud-agnostic set of workspace values that get formatted
// into per-cloud tags/labels. Keys are assigned per-cloud; values are sanitized
// per-cloud.
type labelSource struct {
	stackName          string
	envName            string
	blockName          string
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
// value (e.g. dataClassification before NUL-99 lands) are skipped by the builders.
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

// toTfStringMap converts a Go string map into a tftypes string-valued map.
func toTfStringMap(m map[string]string) map[string]tftypes.Value {
	out := make(map[string]tftypes.Value, len(m))
	for k, v := range m {
		out[k] = tftypes.NewValue(tftypes.String, v)
	}
	return out
}
