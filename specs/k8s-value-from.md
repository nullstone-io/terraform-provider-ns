# Spec: `ns_env_variables` support for K8s `valueFrom` templates

## Background

Nullstone modules accept environment variables as a flat `map(string)`. For Kubernetes workloads, some env vars need to be sourced from the K8s downward API, ConfigMaps, resource fields, or files rather than set as literal values. Today, the provider already supports a `{{ secret(...) }}` template syntax that gets parsed and routed to `secret_refs`. This spec extends the same pattern to four additional K8s `valueFrom` sources.

## Template Syntax

Users write these as values in the flat env var map:

| Template | K8s `valueFrom` type | Example |
|---|---|---|
| `{{ k8s.field(<apiVersion>, <fieldPath>) }}` | `fieldRef` | `{{ k8s.field(v1, status.podIP) }}` |
| `{{ k8s.configMap(<key>, <name>[, <optional>]) }}` | `configMapKeyRef` | `{{ k8s.configMap(my-key, my-cm, true) }}` |
| `{{ k8s.resourceField(<resource>[, <container>, <divisor>]) }}` | `resourceFieldRef` | `{{ k8s.resourceField(limits.cpu) }}` |
| `{{ k8s.fileKey(<key>, <path>, <volumeName>) }}` | `fileKeyRef` | `{{ k8s.fileKey(CONFIG_VAR, config.env, config-volume) }}` |

All templates follow the existing convention: `{{ <namespace>.<type>(<args>) }}`. The `k8s.` namespace scopes these functions to Kubernetes workloads. The existing `secret(...)` function remains unnamespaced since it is provider-agnostic.

### Argument details

**`k8s.field(apiVersion, fieldPath)`**
- `apiVersion` (required): e.g. `v1`
- `fieldPath` (required): e.g. `status.podIP`, `metadata.name`, `metadata.namespace`

**`k8s.configMap(key, name, optional)`**
- `key` (required): The key within the ConfigMap
- `name` (required): The ConfigMap name
- `optional` (optional, default `false`): Whether the ConfigMap/key must exist. Accepts `true`/`false`.

**`k8s.resourceField(resource, container, divisor)`**
- `resource` (required): e.g. `limits.cpu`, `requests.memory`
- `container` (optional): Container name; omit to use the current container
- `divisor` (optional): Resource divisor; e.g. `1Mi`

**`k8s.fileKey(key, path, volumeName)`** _(K8s 1.34+, alpha, requires `EnvFiles` feature gate)_
- `key` (required): The key to read from the env file (file uses `KEY=VALUE` syntax)
- `path` (required): Filename within the volume (e.g. `config.env`)
- `volumeName` (required): Name of the `emptyDir` volume containing the file

## Provider Changes

### Parsing

In the `ns_env_variables` data source read/compute logic, after processing `secret(...)` refs, scan remaining `input_env_variables` values for the four new patterns: `k8s.field(...)`, `k8s.configMap(...)`, `k8s.resourceField(...)`, `k8s.fileKey(...)`. Use the same `{{ ... }}` envelope detection. Strip matched entries from the plain `env_variables` output (same as secrets are stripped today).

### New Outputs

Add four new computed attributes to `ns_env_variables`:

```
field_refs: map of env var name => object
  - api_version: string
  - field_path:  string

config_map_refs: map of env var name => object
  - key:      string
  - name:     string
  - optional: bool

resource_field_refs: map of env var name => object
  - resource:  string
  - container: string (empty string if omitted)
  - divisor:   string (empty string if omitted)

file_key_refs: map of env var name => object
  - key:         string
  - path:        string
  - volume_name: string
```

All four outputs should be **non-sensitive**.

### Schema definition (pseudocode)

```go
"field_refs": {
    Type: schema.TypeMap,
    Computed: true,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "api_version": {Type: schema.TypeString},
            "field_path":  {Type: schema.TypeString},
        },
    },
},
"config_map_refs": {
    Type: schema.TypeMap,
    Computed: true,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "key":      {Type: schema.TypeString},
            "name":     {Type: schema.TypeString},
            "optional": {Type: schema.TypeBool},
        },
    },
},
"resource_field_refs": {
    Type: schema.TypeMap,
    Computed: true,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "resource":  {Type: schema.TypeString},
            "container": {Type: schema.TypeString},
            "divisor":   {Type: schema.TypeString},
        },
    },
},
"file_key_refs": {
    Type: schema.TypeMap,
    Computed: true,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "key":         {Type: schema.TypeString},
            "path":        {Type: schema.TypeString},
            "volume_name": {Type: schema.TypeString},
        },
    },
},
```

### Filtering

Entries matched by any of the four new patterns MUST be removed from the existing `env_variables` output so they are not also emitted as plain `value` env vars. This is the same filtering behavior already applied for `secret(...)` refs.

### Error handling

- Malformed templates (wrong number of args, empty args) should produce a diagnostic error at plan time with the env var name and the raw value.
- Unknown template types inside `{{ ... }}` should be left as-is in `env_variables` (passthrough) so existing behavior is preserved and future types can be added without breaking.

## Module Changes (consumer side)

In `env_vars.tf`:

```hcl
locals {
  env_var_field_refs          = data.ns_env_variables.this.field_refs
  env_var_config_map_refs     = data.ns_env_variables.this.config_map_refs
  env_var_resource_field_refs = data.ns_env_variables.this.resource_field_refs
  env_var_file_key_refs       = data.ns_env_variables.this.file_key_refs
}
```

In `deployment.tf`, replace the placeholder static `env` block (current lines 302-322) with:

```hcl
dynamic "env" {
  for_each = local.env_var_field_refs
  content {
    name = env.key
    value_from {
      field_ref {
        api_version = env.value.api_version
        field_path  = env.value.field_path
      }
    }
  }
}

dynamic "env" {
  for_each = local.env_var_config_map_refs
  content {
    name = env.key
    value_from {
      config_map_key_ref {
        key      = env.value.key
        name     = env.value.name
        optional = env.value.optional
      }
    }
  }
}

dynamic "env" {
  for_each = local.env_var_resource_field_refs
  content {
    name = env.key
    value_from {
      resource_field_ref {
        resource       = env.value.resource
        container_name = env.value.container
        divisor        = env.value.divisor
      }
    }
  }
}

# Requires K8s 1.34+ and EnvFiles feature gate
dynamic "env" {
  for_each = local.env_var_file_key_refs
  content {
    name = env.key
    value_from {
      file_key_ref {
        key         = env.value.key
        path        = env.value.path
        volume_name = env.value.volume_name
      }
    }
  }
}
```

## Test Cases

### Provider unit tests

| Input | Expected output | Expected removed from `env_variables` |
|---|---|---|
| `FOO = "bar"` | No refs; `env_variables["FOO"] = "bar"` | No |
| `IP = "{{ k8s.field(v1, status.podIP) }}"` | `field_refs["IP"] = {api_version: "v1", field_path: "status.podIP"}` | Yes |
| `CM = "{{ k8s.configMap(key, name, true) }}"` | `config_map_refs["CM"] = {key: "key", name: "name", optional: true}` | Yes |
| `CM2 = "{{ k8s.configMap(key, name) }}"` | `config_map_refs["CM2"] = {key: "key", name: "name", optional: false}` | Yes |
| `RES = "{{ k8s.resourceField(limits.cpu) }}"` | `resource_field_refs["RES"] = {resource: "limits.cpu", container: "", divisor: ""}` | Yes |
| `RES2 = "{{ k8s.resourceField(limits.memory, main, 1Mi) }}"` | `resource_field_refs["RES2"] = {resource: "limits.memory", container: "main", divisor: "1Mi"}` | Yes |
| `FK = "{{ k8s.fileKey(MY_VAR, config.env, config-vol) }}"` | `file_key_refs["FK"] = {key: "MY_VAR", path: "config.env", volume_name: "config-vol"}` | Yes |
| `SEC = "{{ secret(my-secret) }}"` | Existing `secret_refs` behavior unchanged | Yes |
| `BAD = "{{ k8s.field() }}"` | Plan-time error: missing required args | N/A |
| `UNK = "{{ k8s.future(a,b) }}"` | No refs; `env_variables["UNK"] = "{{ k8s.future(a,b) }}"` (passthrough) | No |

### Integration / acceptance tests

- Deploy with a mix of plain, secret, field_ref, config_map_ref, resource_field_ref, and file_key_ref env vars and verify the resulting K8s pod spec has the correct env structure.
