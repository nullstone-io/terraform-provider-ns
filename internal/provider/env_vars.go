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
	k8sFieldRefRegex             = regexp.MustCompile("{{\\s*k8s\\.field\\((.+)\\)\\s*}}")
	k8sConfigMapRefRegex         = regexp.MustCompile("{{\\s*k8s\\.configMap\\((.+)\\)\\s*}}")
	k8sResourceFieldRefRegex     = regexp.MustCompile("{{\\s*k8s\\.resourceField\\((.+)\\)\\s*}}")
	k8sFileKeyRefRegex           = regexp.MustCompile("{{\\s*k8s\\.fileKey\\((.+)\\)\\s*}}")
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
	Value            string
	IsSensitive      bool
	SecretRef        *string
	FieldRef         *FieldRef
	ConfigMapRef     *ConfigMapRef
	ResourceFieldRef *ResourceFieldRef
	FileKeyRef       *FileKeyRef
}

type FieldRef struct {
	ApiVersion string
	FieldPath  string
}

type ConfigMapRef struct {
	Key      string
	Name     string
	Optional bool
}

type ResourceFieldRef struct {
	Resource  string
	Container string
	Divisor   string
}

type FileKeyRef struct {
	Key        string
	Path       string
	VolumeName string
}

func (m EnvVars) EnvVars() map[string]string {
	result := map[string]string{}
	for k, v := range m {
		if v.SecretRef == nil && v.FieldRef == nil && v.ConfigMapRef == nil && v.ResourceFieldRef == nil && v.FileKeyRef == nil && !v.IsSensitive {
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

func (m EnvVars) FieldRefs() map[string]FieldRef {
	result := map[string]FieldRef{}
	for k, v := range m {
		if v.FieldRef != nil {
			result[k] = *v.FieldRef
		}
	}
	return result
}

func (m EnvVars) ConfigMapRefs() map[string]ConfigMapRef {
	result := map[string]ConfigMapRef{}
	for k, v := range m {
		if v.ConfigMapRef != nil {
			result[k] = *v.ConfigMapRef
		}
	}
	return result
}

func (m EnvVars) ResourceFieldRefs() map[string]ResourceFieldRef {
	result := map[string]ResourceFieldRef{}
	for k, v := range m {
		if v.ResourceFieldRef != nil {
			result[k] = *v.ResourceFieldRef
		}
	}
	return result
}

func (m EnvVars) FileKeyRefs() map[string]FileKeyRef {
	result := map[string]FileKeyRef{}
	for k, v := range m {
		if v.FileKeyRef != nil {
			result[k] = *v.FileKeyRef
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

func parseTemplateArgs(raw string) []string {
	parts := strings.Split(raw, ",")
	args := make([]string, 0, len(parts))
	for _, p := range parts {
		args = append(args, strings.TrimSpace(p))
	}
	return args
}

func (m EnvVars) Interpolate() []error {
	var errs []error

	// 1. Mark env var values with secret ref or k8s valueFrom refs
	// Scan all env vars, checking for `{{ secret(...) }}`, `{{ k8s.field(...) }}`, etc.
	// Extract the ref and attach to the value
	for k, v := range m {
		result := secretRefRegex.FindStringSubmatch(v.Value)
		if len(result) > 1 {
			secretRef := result[1]
			v.SecretRef = &secretRef
			m[k] = v
			continue
		}

		if result = k8sFieldRefRegex.FindStringSubmatch(v.Value); len(result) > 1 {
			args := parseTemplateArgs(result[1])
			if len(args) != 2 || args[0] == "" || args[1] == "" {
				errs = append(errs, fmt.Errorf("env var %q: k8s.field requires 2 arguments (apiVersion, fieldPath), got %q", k, result[1]))
				continue
			}
			v.FieldRef = &FieldRef{ApiVersion: args[0], FieldPath: args[1]}
			m[k] = v
			continue
		}

		if result = k8sConfigMapRefRegex.FindStringSubmatch(v.Value); len(result) > 1 {
			args := parseTemplateArgs(result[1])
			if len(args) < 2 || len(args) > 3 || args[0] == "" || args[1] == "" {
				errs = append(errs, fmt.Errorf("env var %q: k8s.configMap requires 2-3 arguments (key, name[, optional]), got %q", k, result[1]))
				continue
			}
			ref := ConfigMapRef{Key: args[0], Name: args[1]}
			if len(args) == 3 {
				ref.Optional = args[2] == "true"
			}
			v.ConfigMapRef = &ref
			m[k] = v
			continue
		}

		if result = k8sResourceFieldRefRegex.FindStringSubmatch(v.Value); len(result) > 1 {
			args := parseTemplateArgs(result[1])
			if len(args) < 1 || len(args) > 3 || args[0] == "" {
				errs = append(errs, fmt.Errorf("env var %q: k8s.resourceField requires 1-3 arguments (resource[, container, divisor]), got %q", k, result[1]))
				continue
			}
			ref := ResourceFieldRef{Resource: args[0]}
			if len(args) >= 2 {
				ref.Container = args[1]
			}
			if len(args) == 3 {
				ref.Divisor = args[2]
			}
			v.ResourceFieldRef = &ref
			m[k] = v
			continue
		}

		if result = k8sFileKeyRefRegex.FindStringSubmatch(v.Value); len(result) > 1 {
			args := parseTemplateArgs(result[1])
			if len(args) != 3 || args[0] == "" || args[1] == "" || args[2] == "" {
				errs = append(errs, fmt.Errorf("env var %q: k8s.fileKey requires 3 arguments (key, path, volumeName), got %q", k, result[1]))
				continue
			}
			v.FileKeyRef = &FileKeyRef{Key: args[0], Path: args[1], VolumeName: args[2]}
			m[k] = v
			continue
		}
	}

	// 2. Interpolate secrets onto other env vars
	// This has the potential to promote env vars to secrets
	// Loop until convergence to handle chained references (e.g. A → B → SECRET)
	for changed := true; changed; {
		changed = false
		for k1, v1 := range m.Secrets() {
			replacer := regexp.MustCompile(fmt.Sprintf(interpolationRefRegexPattern, k1))
			for k2, v2 := range m.EnvVars() {
				result := replacer.ReplaceAllString(v2, v1)
				// if a match was found and replaced, this env variable is now a secret
				if result != v2 {
					changed = true
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
						changed = true
						entry := m[k2]
						entry.IsSensitive = true
						entry.Value = result
						m[k2] = entry
					}
				}
			}
		}
	}

	// 3. Interpolate env vars onto other env vars/secrets
	// This will not promote anybody to a secret
	// Loop until convergence to handle chained references (e.g. A → B → C)
	for changed := true; changed; {
		changed = false
		for k1, v1 := range m.EnvVars() {
			regex := regexp.MustCompile(fmt.Sprintf(interpolationRefRegexPattern, k1))
			for k2, v2 := range m.EnvVars() {
				// we don't want to replace the env variable with itself (this will prevent an infinite loop)
				if k2 != k1 {
					result := regex.ReplaceAllString(v2, v1)
					if result != v2 {
						changed = true
						entry := m[k2]
						entry.Value = result
						m[k2] = entry
					}
				}
			}
			for k2, v2 := range m.Secrets() {
				result := regex.ReplaceAllString(v2, v1)
				if result != v2 {
					changed = true
					entry := m[k2]
					entry.Value = result
					m[k2] = entry
				}
			}
		}
	}

	return errs
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
