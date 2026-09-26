# Layered env data sources: `ns_env_layout` and `ns_env_values`

Status: SPEC, 2026-09-25 (NUL-132, Proposal A). Additive: `ns_env_variables` and `ns_secret_keys` are unchanged and remain supported; modules upgrade one at a time.

## Problem

Every app module re-derives the same partition in ~60 lines of HCL: merge five layers of env vars, split secrets into managed (module creates the cloud secret) vs unmanaged (`{{ secret(...) }}` points at an existing one), and pull out k8s `valueFrom` refs. The partition is also what the Nullstone UI needs to show, along with *where each variable came from*. This spec moves the derivation into the provider and emits the platform-data `env` record with a `source` per key.

## Why two data sources

Terraform defers a data source read to apply when any input is unknown. Capability secret values are frequently unknown at plan (a password from a resource being created), but modules `for_each` over managed secret keys to create cloud secrets, so the key set must be known at plan. `ns_env_layout` takes keys only and is always resolvable at plan. `ns_env_values` takes values and may defer. This mirrors today's `ns_secret_keys` / `ns_env_variables` split.

## Layers and precedence

Lowest to highest; a later layer overrides an earlier one for the same key (matches `merge(standard, cloud, otel, cap, user)` in every module today):

| layer | source label | example |
|---|---|---|
| standard | `standard` | `NULLSTONE_STACK`, `NULLSTONE_VERSION`, `NULLSTONE_PUBLIC_HOSTS` |
| cloud | `cloud` | `GOOGLE_CLOUD_PROJECT`, `AWS_REGION`, `AZURE_CLIENT_ID` |
| otel | `otel` | `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_RESOURCE_ATTRIBUTES` |
| capability | `capability` (+ capability id) | `PG_HOST` from a postgres capability, prefixed |
| user | `user` | `var.env_vars`, `var.secrets` |

Source labels are `platformdata.Source*` in `github.com/nullstone-io/module/platformdata`.

## Platform

`platform` (required string) is validated against `platformdata.LookupPlatform` (module repo). It gates templates:

- `k8s.field/configMap/resourceField/fileKey(...)` → error unless `SupportsK8sRefs`.
- `secret(...)` → error unless `SupportsSecretRefs`.

Unknown platform → error at plan.

## Shared inputs

Both data sources accept:

| attribute | type | notes |
|---|---|---|
| `platform` | string, required | see above |
| `standard` | map(string), optional | |
| `cloud` | map(string), optional | |
| `otel` | map(string), optional | |
| `capability_env` | list(object({capability=string, name=string, value=string})), optional | pass `local.capabilities.env` verbatim |
| `capability_prefixes` | map(string), optional | `local.cap_prefixes` verbatim: capability name → prefix; final key = prefix + name |

`capability` is the **capability name** (what users see in the UI), not the legacy `tfId`. The generated `capabilities.tf` (from each module's `capabilities.tf.tmpl`, scaffolded by `nullstone modules generate`) emits items as `merge({ cap_tf_id = mod.tfId, capability = mod.name }, x)` and a `cap_prefixes` local keyed by `mod.name` next to the legacy `cap_env_prefixes`. Terraform drops the extra `cap_tf_id` attribute when converting to the provider's object type (covered by `TestDataEnvValues_ExtraCapabilityAttributesTolerated`), so existing modules that still read `cap_tf_id` are unaffected and upgraded modules pass the locals verbatim.
| `user_env` | map(string), optional | `var.env_vars` |

Secrets differ:

| data source | attribute | type |
|---|---|---|
| `ns_env_layout` | `capability_secret_keys` | list(object({capability=string, name=string})) |
| `ns_env_layout` | `user_secret_keys` | set(string) — `nonsensitive(keys(var.secrets))` |
| `ns_env_values` | `capability_secrets` | list(object({capability=string, name=string, value=string})), sensitive |
| `ns_env_values` | `user_secrets` | map(string), sensitive — `var.secrets` |

Use object-typed attributes (not nested blocks); all object attributes are required (tfprotov5 object types have no optionals in provider schemas).

## Resolution (both data sources)

1. Fold layers in precedence order into one `EnvVars` (existing type in `env_vars.go`) keeping, per key: value, is-secret, source, capability id. Capability keys are `prefix + name`.
2. Validate keys with `validEnvVariableKey`; duplicate keys across layers are allowed (override), duplicate within a layer is impossible (maps) except `capability_env` list → error on duplicate final key.
3. Run the existing `Interpolate()` (secret promotion, `secret(...)` refs, k8s refs, `{{ KEY }}` substitution). Semantics are identical to `ns_env_variables`.
4. Apply platform gates from `platformdata.LookupPlatform`.
5. Classify each key:
   - **unmanaged secret**: has `SecretRef` (from `secret(...)`).
   - **managed secret**: `IsSensitive` (from a secrets input or promoted by interpolation).
   - **k8s ref**: one of the four refs.
   - **plain**: everything else.

`ns_env_layout` runs the same steps with empty secret values (like `ns_secret_keys` does) so the key classification is available at plan.

## Outputs

`ns_env_layout` (computed):

| attribute | type | meaning |
|---|---|---|
| `managed_secret_keys` | set(string) | secrets the module must create |
| `unmanaged_secret_keys` | set(string) | `secret(...)` refs |
| `all_secret_keys` | set(string) | union |
| `sources` | map(string) | key → source label |
| `capabilities` | map(string) | key → capability (only capability-sourced keys) |

`ns_env_values` (computed):

| attribute | type | meaning |
|---|---|---|
| `env_variables` | map(string) | plain values (post-interpolation) |
| `secrets` | map(string), sensitive | managed secret values |
| `unmanaged_secret_refs` | map(string) | key → secret ref |
| `field_refs`, `config_map_refs`, `resource_field_refs`, `file_key_refs` | maps of the existing object types | k8s refs (empty unless platform supports them) |
| `managed_secret_keys`, `unmanaged_secret_keys`, `all_secret_keys` | set(string) | same as layout |
| `sources`, `capabilities` | map(string) | same as layout |
| `platform_data` | string, **not** sensitive | env v1 JSON record (below) |

`id` for both: hash of sorted keys (reuse `KeysHash()`).

## `platform_data` record

`platformdata.EnvV1` JSON:

```json
{
  "platform": "k8s",
  "variables": {
    "NULLSTONE_ENV":     { "template": "prod", "value": "prod", "source": "standard" },
    "PG_HOST":           { "template": "{{ ... }}", "value": "db.internal", "source": "capability", "capability": "postgres0" },
    "DATABASE_PASSWORD": { "template": "{{ secret(...) }}", "sensitive": true, "ref": { "type": "secret", "id": "arn:..." }, "source": "user" },
    "PG_PASSWORD":       { "sensitive": true, "source": "capability", "capability": "postgres0" },
    "POD_IP":            { "ref": { "type": "k8s_field", "api_version": "v1", "field_path": "status.podIP" }, "source": "user" }
  }
}
```

Rules: `template` = the pre-interpolation input value; `value` only for plain keys; managed secrets are `sensitive` with no value and no ref (the module mints the id later); unmanaged secrets are `sensitive` + `ref{type: secret}`; k8s refs map to the four `k8s_*` ref types. The record must pass `platformdata.ParseEnvV1` before being emitted; failure is a provider bug → error diagnostic.

## Module usage (target shape)

```hcl
data "ns_env_layout" "this" {
  platform            = "k8s"
  standard            = local.standard_env_vars
  cloud               = local.google_env_vars
  otel                = local.otel_env_vars
  capability_env      = local.capabilities.env
  capability_secret_keys = [for s in local.capabilities.secrets : { capability = s.capability, name = s.name }]
  capability_prefixes = local.cap_prefixes
  user_env            = var.env_vars
  user_secret_keys    = nonsensitive(keys(var.secrets))
}

data "ns_env_values" "this" {
  platform            = "k8s"
  standard            = local.standard_env_vars
  cloud               = local.google_env_vars
  otel                = local.otel_env_vars
  capability_env      = local.capabilities.env
  capability_secrets  = local.capabilities.secrets
  capability_prefixes = local.cap_prefixes
  user_env            = var.env_vars
  user_secrets        = var.secrets
}

data "ns_platform_data" "env" {
  kind    = "env"
  version = 1
  data    = data.ns_env_values.this.platform_data
}
```

Replaces `ns_env_variables.this`, `ns_env_variables.existing`, `ns_secret_keys.this`, `cap_env_vars`/`cap_secrets` loops and the managed/unmanaged set arithmetic.

## Arcana

Arcana's legacy adapter gains a second source: when no explicit `ns_platform_data kind=env` exists, it reads `data.ns_env_values.this.platform_data` from state (already a valid env v1 record, passthrough) before falling back to reconstructing from `data.ns_env_variables.this`. A module therefore keeps its env record through the upgrade even before adding the `ns_platform_data` block.

## Non-goals

- Changing `ns_env_variables` / `ns_secret_keys` behaviour.
- Emitting per-platform resource shapes (ECS `secrets[]`, k8s `env` blocks). Modules keep that wiring.
- Managed secret ids in the record (follow-up: `patches` on `ns_platform_data`).
