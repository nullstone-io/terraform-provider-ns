---
layout: "ns"
page_title: "Nullstone: ns_env_variables"
sidebar_current: "docs-ns-env-variables"
description: |-
  Data source to interpolate environment variables and secrets into their final values.
---

# ns_env_variables

Data source to interpolate environment variables and secrets into their final values.
It resolves `{{ VAR }}` references between variables, promotes variables to secrets when they reference a secret, and extracts special template syntax for secret refs and Kubernetes `valueFrom` sources.

## Example Usage

#### Basic example

```hcl
data "ns_env_variables" "this" {
  input_env_variables = {
    NULLSTONE_STACK = "primary"
    NULLSTONE_BLOCK = "acme-api"
    NULLSTONE_ENV   = "dev"
    IDENTIFIER      = "{{ NULLSTONE_STACK }}.{{ NULLSTONE_BLOCK }}.{{ NULLSTONE_ENV }}"
    DATABASE_URL    = "{{ POSTGRES_URL }}"
  }
  input_secrets = {
    POSTGRES_URL = "postgres://user:pass@host:5432/mydb"
  }
}
```

After interpolation:
- `env_variables["IDENTIFIER"]` = `"primary.acme-api.dev"`
- `secrets["DATABASE_URL"]` = `"postgres://user:pass@host:5432/mydb"` (promoted because it references a secret)

#### Secret ref example

```hcl
data "ns_env_variables" "this" {
  input_env_variables = {
    MY_SECRET = "{{ secret(arn:aws:secretsmanager:us-east-1:123456789:secret:my-secret) }}"
  }
  input_secrets = {}
}
```

- `secret_refs["MY_SECRET"]` = `"arn:aws:secretsmanager:us-east-1:123456789:secret:my-secret"`

#### Kubernetes valueFrom example

```hcl
data "ns_env_variables" "this" {
  input_env_variables = {
    POD_IP    = "{{ k8s.field(v1, status.podIP) }}"
    APP_CFG   = "{{ k8s.configMap(app-key, my-configmap, true) }}"
    CPU_LIMIT = "{{ k8s.resourceField(limits.cpu) }}"
    FILE_VAR  = "{{ k8s.fileKey(MY_VAR, config.env, config-vol) }}"
  }
  input_secrets = {}
}
```

- `field_refs["POD_IP"]` = `{ api_version = "v1", field_path = "status.podIP" }`
- `config_map_refs["APP_CFG"]` = `{ key = "app-key", name = "my-configmap", optional = true }`
- `resource_field_refs["CPU_LIMIT"]` = `{ resource = "limits.cpu", container = "", divisor = "" }`
- `file_key_refs["FILE_VAR"]` = `{ key = "MY_VAR", path = "config.env", volume_name = "config-vol" }`

## Template Syntax

Values in `input_env_variables` can use the following template patterns:

| Template | Description | Example |
|---|---|---|
| `{{ VAR_NAME }}` | Interpolate another variable's value | `{{ POSTGRES_URL }}` |
| `{{ secret(<ref>) }}` | Reference an external secret | `{{ secret(arn:aws:...) }}` |
| `{{ k8s.field(<apiVersion>, <fieldPath>) }}` | Kubernetes `fieldRef` | `{{ k8s.field(v1, status.podIP) }}` |
| `{{ k8s.configMap(<key>, <name>[, <optional>]) }}` | Kubernetes `configMapKeyRef` | `{{ k8s.configMap(my-key, my-cm, true) }}` |
| `{{ k8s.resourceField(<resource>[, <container>, <divisor>]) }}` | Kubernetes `resourceFieldRef` | `{{ k8s.resourceField(limits.cpu) }}` |
| `{{ k8s.fileKey(<key>, <path>, <volumeName>) }}` | Kubernetes `fileKeyRef` | `{{ k8s.fileKey(MY_VAR, config.env, config-vol) }}` |

### Template argument details

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
- `key` (required): The key to read from the env file
- `path` (required): Filename within the volume (e.g. `config.env`)
- `volumeName` (required): Name of the volume containing the file

Entries matched by `secret(...)` or any `k8s.*` template are removed from `env_variables` and appear only in their respective ref output. Unknown template types (e.g. `{{ k8s.future(a,b) }}`) are left as-is in `env_variables`.

Malformed templates (wrong number of arguments, empty arguments) produce a diagnostic error at plan time.

## Arguments Reference

* `input_env_variables` - (Required) A map of environment variable names to their raw values before interpolation.
* `input_secrets` - (Required, Sensitive) A map of secret names to their raw values before interpolation.

## Attributes Reference

* `env_variables` - A map of non-sensitive environment variables after interpolation. Variables that were promoted to secrets or matched a template pattern are excluded.
* `secrets` - (Sensitive) A map of secret environment variables after interpolation. Includes original secrets and any variables promoted to secrets by referencing a secret.
* `secret_refs` - A map of environment variable names to their extracted secret reference strings (from `{{ secret(...) }}`).
* `field_refs` - A map of environment variable names to objects with:
    * `api_version` - The Kubernetes API version (e.g. `v1`).
    * `field_path` - The field path (e.g. `status.podIP`).
* `config_map_refs` - A map of environment variable names to objects with:
    * `key` - The key within the ConfigMap.
    * `name` - The ConfigMap name.
    * `optional` - Whether the ConfigMap/key must exist (`true`/`false`).
* `resource_field_refs` - A map of environment variable names to objects with:
    * `resource` - The resource name (e.g. `limits.cpu`).
    * `container` - The container name (empty string if omitted).
    * `divisor` - The resource divisor (empty string if omitted).
* `file_key_refs` - A map of environment variable names to objects with:
    * `key` - The key to read from the env file.
    * `path` - The filename within the volume.
    * `volume_name` - The name of the volume containing the file.
