package provider

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var (
	secretRefRegex               = regexp.MustCompile("{{\\s*secret\\((.+)\\)\\s*}}")
	interpolationRefRegexPattern = "{{\\s*%s\\s*}}"
)

type EnvVars map[string]EnvVar

func NewEnvVars(envVars map[string]string, secrets map[string]string) EnvVars {
	mixed := EnvVars{}
	for k, v := range envVars {
		mixed[k] = EnvVar{Value: v}
	}
	for k, v := range secrets {
		mixed[k] = EnvVar{Value: v, IsSensitive: true}
	}
	return mixed
}

type EnvVar struct {
	Value       string
	IsSensitive bool
	SecretRef   *string
}

func (m EnvVars) EnvVars() map[string]string {
	result := map[string]string{}
	for k, v := range m {
		if v.SecretRef == nil && !v.IsSensitive {
			result[k] = v.Value
		}
	}
	return result
}

func (m EnvVars) Secrets() map[string]string {
	result := map[string]string{}
	for k, v := range m {
		if v.IsSensitive {
			result[k] = v.Value
		}
	}
	return result
}

func (m EnvVars) SecretRefs() map[string]string {
	result := map[string]string{}
	for k, v := range m {
		if v.SecretRef != nil {
			result[k] = *v.SecretRef
		}
	}
	return result
}

func (m EnvVars) SecretKeys() []string {
	result := make([]string, 0)
	for k, v := range m {
		if v.IsSensitive {
			result = append(result, k)
		}
	}
	slices.SortStableFunc(result, strings.Compare)
	return result
}

func (m EnvVars) Interpolate() {
	// 1. Mark env var values with secret ref
	// Scan all env vars, checking for `{{ secret(...) }}`
	// Extract the secret ref and attach to the value
	for k, v := range m {
		result := secretRefRegex.FindStringSubmatch(v.Value)
		if len(result) > 1 {
			secretRef := result[1]
			v.SecretRef = &secretRef
			m[k] = v
		}
	}

	// 2. Interpolate secrets onto other env vars
	// This has the potential promote env vars to secrets
	for k1, v1 := range m.Secrets() {
		replacer := regexp.MustCompile(fmt.Sprintf(interpolationRefRegexPattern, k1))
		for k2, v2 := range m.EnvVars() {
			result := replacer.ReplaceAllString(v2, v1)
			// if a match was found and replaced, this env variable is now a secret
			if result != v2 {
				entry := m[k2]
				entry.IsSensitive = true
				entry.Value = result
				m[k2] = entry
			}
		}
		for k2, v2 := range m.Secrets() {
			if k2 != k1 {
				result := replacer.ReplaceAllString(v2, v1)
				if result != v2 {
					entry := m[k2]
					entry.IsSensitive = true
					entry.Value = result
					m[k2] = entry
				}
			}
		}
	}

	// 3. Interpolate env vars onto other env vars/secrets
	// This will not promote anybody to a secret
	for k1, v1 := range m.EnvVars() {
		regex := regexp.MustCompile(fmt.Sprintf(interpolationRefRegexPattern, k1))
		for k2, v2 := range m.EnvVars() {
			// we don't want to replace the env variable with itself (this will prevent an infinite loop)
			if k2 != k1 {
				result := regex.ReplaceAllString(v2, v1)
				if result != v2 {
					entry := m[k2]
					entry.Value = result
					m[k2] = entry
				}
			}
		}
		for k2, v2 := range m.Secrets() {
			result := regex.ReplaceAllString(v2, v1)
			if result != v2 {
				entry := m[k2]
				entry.Value = result
				m[k2] = entry
			}
		}
	}
}

func (m EnvVars) Hash() string {
	hashString := ""
	for k, v := range m {
		sensitive := ""
		if v.IsSensitive {
			sensitive = "+"
		}
		hashString += fmt.Sprintf("%s=%s%s;", k, v.Value, sensitive)
	}

	sum := sha256.Sum256([]byte(hashString))
	return fmt.Sprintf("%x", sum)
}

func (m EnvVars) KeysHash() string {
	hashString := ""
	for k, v := range m {
		sensitive := ""
		if v.IsSensitive {
			sensitive = "+"
		}
		hashString += fmt.Sprintf("%s%s;", k, sensitive)
	}

	sum := sha256.Sum256([]byte(hashString))
	return fmt.Sprintf("%x", sum)
}
