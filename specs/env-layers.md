# Layered env data sources: `ns_env_layout`, `ns_env_values`, `ns_env_platform_data`

Status: SPEC, 2026-09-25 (NUL-132, Proposal A); Phase 2c decisions D14–D16 folded in 2026-09-28. Additive: `ns_env_variables` and `ns_secret_keys` are unchanged and remain supported; modules upgrade one at a time.

## Problem

Every app module re-derives the same partition in ~60 lines of HCL: merge five layers of env vars, split secrets into managed (module creates the cloud secret) vs unmanaged (`{{ secret(...) }}` points at an existing one), and pull out k8s `valueFrom` refs. The partition is also what the Nullstone UI needs to show, along with *where each variable came from* and *where each managed secret lives*. This spec moves the derivation into the provider and emits the platform-data `env` record with a `source` per key and a `ref` per managed secret.

## Why three data sources

Terraform defers a data source read to apply when any input is unknown. Capability secret values are frequently unknown at plan (a password from a resource being created), but modules `for_each` over managed secret keys to create cloud secrets, so the key set must be known at plan. `ns_env_layout` takes keys only and is always resolvable at plan. `ns_env_values` takes values and may defer. This mirrors today's `ns_secret_keys` / `ns_env_variables` split.

The ids of the managed secrets only exist after the module creates them, and `ns_env_values` must not depend on those resources (it feeds their values). `ns_env_platform_data` (D14) is the third step: it takes the record from `ns_env_values` plus the ids and emits the record Nullstone reads.

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

### D16 — Secrets always win

If any secrets layer (`capability_secrets`/`capability_secret_keys` or `user_secrets`/`user_secret_keys`) sets a key, a plain layer at the same key **never** overrides it, regardless of layer order: the plain layer's value, template, and source are ignored and the secret's are kept. Precedence among secret layers still follows layer order (capability < user). Example: `user_env.PG_PASSWORD` vs a capability secret `PG_PASSWORD` → managed secret, source `capability`.

Rationale: a plain override of a secret would either leak the secret's key as plaintext or silently drop the secret; neither is ever intended.

## Platform

`platform` (required string) is validated against `platformdata.LookupPlatform` (module repo). Identifiers are always cloud-prefixed and unambiguous on their own: `aws_ecs`, `aws_batch`, `aws_lambda`, `aws_beanstalk`, `aws_ec2`, `aws_s3`, `aws_eks`, `gcp_gke`, `gcp_cloudrun`, `gcp_cloudfunctions`, `gcp_composer`, `gcp_gce`, `gcp_gcs`, `azure_aks`, `azure_container_app`, `azure_function`, `azure_app_service`, `azure_static_web_app`. Kubernetes is split per cloud (`aws_eks` / `gcp_gke` / `azure_aks`); all three support k8s refs. It gates templates:

- `k8s.field/configMap/resourceField/fileKey(...)` → error unless `SupportsK8sRefs`.
- `secret(...)` → error unless `SupportsSecretRefs`.

Unknown platform → error at plan.

## Inputs

### D15 — `ns_env_layout` takes keys only for non-user layers

| attribute | type | notes |
|---|---|---|
| `platform` | string, required | see above |
| `standard_keys` | set(string), optional | `keys(local.standard_env_vars)` |
| `cloud_keys` | set(string), optional | `keys(local.google_env_vars)` |
| `otel_keys` | set(string), optional | `keys(local.otel_env_vars)` |
| `capability_env_keys` | list(object({capability=string, name=string})), optional | `[for e in local.capabilities.env : { capability = e.capability, name = e.name }]` |
| `capability_secret_keys` | list(object({capability=string, name=string})), optional | same, from `local.capabilities.secrets` |
| `capability_prefixes` | map(string), optional | `local.cap_prefixes` verbatim: capability name → prefix; final key = prefix + name |
| `user_env` | map(string), optional | `var.env_vars` — the **full map**: templates are needed to detect secret promotion (`{{ SOME_SECRET }}`) and `{{ secret(...) }}` refs |
| `user_secret_keys` | set(string), optional | `nonsensitive(keys(var.secrets))` |

Resolution uses `""` as the value for every key-only layer (like `ns_secret_keys` does), so the key classification is available at plan. `standard`/`cloud`/`otel` values are never templates in practice, so dropping them loses nothing.

### `ns_env_values`

| attribute | type | notes |
|---|---|---|
| `platform` | string, required | |
| `standard` | map(string), optional | |
| `cloud` | map(string), optional | |
| `otel` | map(string), optional | |
| `capability_env` | list(object({capability=string, name=string, value=string})), optional | pass `local.capabilities.env` verbatim |
| `capability_secrets` | list(object({capability=string, name=string, value=string})), optional, sensitive | `local.capabilities.secrets` verbatim |
| `capability_prefixes` | map(string), optional | as above |
| `user_env` | map(string), optional | `var.env_vars` |
| `user_secrets` | map(string), optional, sensitive | `var.secrets` |

`capability` is the **capability name** (what users see in the UI), not the legacy `tfId`. The generated `capabilities.tf` (from each module's `capabilities.tf.tmpl`, scaffolded by `nullstone modules generate`) emits items as `merge({ cap_tf_id = mod.tfId, capability = mod.name }, x)` and a `cap_prefixes` local keyed by `mod.name` next to the legacy `cap_env_prefixes`. Terraform drops the extra `cap_tf_id` attribute when converting to the provider's object type (covered by `TestDataEnvValues_ExtraCapabilityAttributesTolerated`), so existing modules that still read `cap_tf_id` are unaffected and upgraded modules pass the locals verbatim.

Use object-typed attributes (not nested blocks); all object attributes are required (tfprotov5 object types have no optionals in provider schemas).

### D16 — plan-time errors

Raised in `Validate` when the relevant values are known, and again in `Read`:

| condition | summary | notes |
|---|---|---|
| a `{{ secret(...) }}` or `{{ k8s.*(...) }}` template inside a secrets input (`capability_secrets` / `user_secrets` values) | `Invalid secret template: KEY` | detail tells the author to move the variable to `env_vars`; secrets hold the value itself and cannot use runtime reference templates. Detected with the same regexes `Interpolate()` uses. Only `ns_env_values` can check this (the layout has no secret values). |
| a capability entry whose `capability` has no key in `capability_prefixes` | `Unknown capability: NAME` | detail names the attribute; an explicit `""` prefix is fine. Skipped at `Validate` while `capability_prefixes` is unknown. |
| empty `capability` in any capability list | `Capability name is required` | |

## Resolution (both data sources)

Both input shapes feed one resolver (`resolveLayers` over `layeredEnvInput`); the layout builds maps with `""` values.

1. Fold layers in precedence order into one `EnvVars` (existing type in `env_vars.go`) keeping, per key: value, is-secret, source, capability id. Capability keys are `prefix + name`. D16: a plain layer never replaces a secret.
2. Validate keys with `validEnvVariableKey`; duplicate keys across layers are allowed (override), duplicate within a layer is impossible (maps) except capability lists → error on duplicate final key.
3. Run the existing `Interpolate()` (secret promotion, `secret(...)` refs, k8s refs, `{{ KEY }}` substitution). Semantics are identical to `ns_env_variables`.
4. Apply platform gates from `platformdata.LookupPlatform`.
5. Classify each key:
   - **unmanaged secret**: has `SecretRef` (from `secret(...)`).
   - **managed secret**: `IsSensitive` (from a secrets input or promoted by interpolation).
   - **k8s ref**: one of the four refs.
   - **plain**: everything else.

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
| `platform_data` | string, **not** sensitive | env v1 JSON record (below); managed secrets carry no ref yet |

`id` for all three data sources: hash of the **sorted** keys (`KeysHash()`, which sorts; the pre-2c implementation iterated a map and was non-deterministic, which also affected legacy `ns_secret_keys`).

## `platform_data` record

`platformdata.EnvV1` JSON as emitted by `ns_env_values`:

```json
{
  "platform": "gcp_gke",
  "variables": {
    "NULLSTONE_ENV":     { "template": "prod", "value": "prod", "source": "standard" },
    "PG_HOST":           { "template": "{{ ... }}", "value": "db.internal", "source": "capability", "capability": "postgres0" },
    "DATABASE_PASSWORD": { "template": "{{ secret(...) }}", "sensitive": true, "ref": { "type": "secret", "id": "arn:..." }, "source": "user" },
    "PG_PASSWORD":       { "sensitive": true, "source": "capability", "capability": "postgres0" },
    "POD_IP":            { "ref": { "type": "k8s_field", "api_version": "v1", "field_path": "status.podIP" }, "source": "user" }
  }
}
```

Rules: `template` = the pre-interpolation input value; `value` only for plain keys (always serialized, so `""` is distinguishable from "no value"); managed secrets are `sensitive` with no value and no ref (the module mints the id later, see D14); unmanaged secrets are `sensitive` + `ref{type: secret}`; k8s refs map to the four `k8s_*` ref types. The record must pass `platformdata.ParseEnvV1` before being emitted; failure is a provider bug → error diagnostic.

## D14 — `ns_env_platform_data`

Completes the record with the ids of the managed secrets.

Inputs:

| attribute | type | notes |
|---|---|---|
| `values` | string, required | `data.ns_env_values.this.platform_data` |
| `secret_ids` | map(string), optional | managed secret key → cloud secret id (AWS ARN, GCP secret resource name, ...) |
| `k8s_secret_refs` | map(object({name=string, key=string})), optional | managed secret key → Kubernetes `secretKeyRef` |

Outputs: `id` (hash of the record's sorted keys), `platform_data` (string, not sensitive).

Read:

1. `platformdata.ParseEnvV1(values)`.
2. For each `secret_ids` entry: the variable must exist and be a managed secret (`Sensitive && Ref == nil`) → `Ref{Type: secret, Id}`.
3. For each `k8s_secret_refs` entry: same rule → `Ref{Type: k8s_secret_key, Name, Key}`; error if the record's platform is known and `!SupportsK8sRefs`.
4. Errors: a key that does not exist or is not a managed secret (`Invalid secret id key: KEY`); a key in both maps (`Secret id declared twice: KEY`).
5. Coverage: every variable still `Sensitive` with `Ref == nil` → one error `Managed secrets without an id: A, B` (every key in `managed_secret_keys` must appear in `secret_ids` or `k8s_secret_refs`).
6. Re-marshal and re-parse (`ParseEnvV1`) before emitting, like `PlatformData()` does.

Validate: if `values` is known, parse it and surface errors; if the id maps are fully known as well, run the whole completion so coverage errors also surface at plan. Everything unknown is skipped.

## Module usage (target shape)

layout → create secrets from `managed_secret_keys` → values → env_platform_data:

```hcl
data "ns_env_layout" "this" {
  platform               = "gcp_gke"
  standard_keys          = keys(local.standard_env_vars)
  cloud_keys             = keys(local.google_env_vars)
  otel_keys              = keys(local.otel_env_vars)
  capability_env_keys    = [for e in local.capabilities.env : { capability = e.capability, name = e.name }]
  capability_secret_keys = [for s in local.capabilities.secrets : { capability = s.capability, name = s.name }]
  capability_prefixes    = local.cap_prefixes
  user_env               = var.env_vars
  user_secret_keys       = nonsensitive(keys(var.secrets))
}

resource "google_secret_manager_secret" "this" {
  for_each  = data.ns_env_layout.this.managed_secret_keys
  secret_id = "${local.resource_name}-${each.key}"
  # ...
}

data "ns_env_values" "this" {
  platform            = "gcp_gke"
  standard            = local.standard_env_vars
  cloud               = local.google_env_vars
  otel                = local.otel_env_vars
  capability_env      = local.capabilities.env
  capability_secrets  = local.capabilities.secrets
  capability_prefixes = local.cap_prefixes
  user_env            = var.env_vars
  user_secrets        = var.secrets
}

resource "google_secret_manager_secret_version" "this" {
  for_each    = data.ns_env_values.this.secrets
  secret      = google_secret_manager_secret.this[each.key].id
  secret_data = each.value
}

data "ns_env_platform_data" "this" {
  values     = data.ns_env_values.this.platform_data
  secret_ids = { for key, secret in google_secret_manager_secret.this : key => secret.id }
  # or, on Kubernetes, k8s_secret_refs = { for key in data.ns_env_layout.this.managed_secret_keys : key => { name = kubernetes_secret.this.metadata[0].name, key = key } }
}
```

Replaces `ns_env_variables.this`, `ns_env_variables.existing`, `ns_secret_keys.this`, `cap_env_vars`/`cap_secrets` loops and the managed/unmanaged set arithmetic. No `ns_platform_data` block is needed for the env kind.

## Arcana

Arcana reads `data.ns_env_platform_data.*` from state directly (already a valid env v1 record, passthrough). Its legacy adapter falls back to `data.ns_env_values.this.platform_data` (valid but without managed-secret refs), then to an explicit `ns_platform_data kind=env`, then to reconstructing from `data.ns_env_variables.this`. A module therefore keeps its env record through every step of the upgrade.

## Non-goals

- Changing `ns_env_variables` / `ns_secret_keys` behaviour (other than the deterministic `id`).
- Emitting per-platform resource shapes (ECS `secrets[]`, k8s `env` blocks). Modules keep that wiring.
- Changing `ns_platform_data`.
