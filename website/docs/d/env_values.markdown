---
layout: "ns"
page_title: "Nullstone: ns_env_values"
sidebar_current: "docs-ns-env-values"
description: |-
  Data source to merge and interpolate the layers of an application's environment and emit the env platform data record.
---

# ns_env_values

Data source that merges the layers of an application's environment (standard, cloud, otel, capability, user), interpolates `{{ VAR }}` references,
splits secrets into managed (the module creates the cloud secret) and unmanaged (`{{ secret(...) }}` points at an existing one),
extracts Kubernetes `valueFrom` refs, and emits the `env` platform data record consumed by Nullstone.

Because it receives secret values, Terraform may defer this data source to apply when any value is unknown at plan.
Use [`ns_env_layout`](env_layout.html) (keys only) wherever the *set* of secrets must be known at plan, such as a `for_each` that creates cloud secrets,
and [`ns_env_platform_data`](env_platform_data.html) to record the ids of the secrets you created in the env record.

Resolution follows the same rules as `ns_env_layout`:

1. Layers are merged in precedence order (lowest to highest): `standard` < `cloud` < `otel` < `capability_env`/`capability_secrets` < `user_env`/`user_secrets`. A later layer overrides an earlier one for the same key.
2. **Secrets always win.** If any secrets input (`capability_secrets`, `user_secrets`) sets a key, a plain layer at the same key never overrides it, regardless of order; the secret keeps its value, template, and source. Secret layers still override each other in order (capability < user).
3. Capability keys are `capability_prefixes[capability] + name`. Every `capability` must have an entry in `capability_prefixes` (an explicit `""` prefix is fine) and must not be empty. A duplicate final key within `capability_env` (or within `capability_secrets`) is an error.
4. Every key is validated (letters, numbers, underscore; cannot begin with a number).
5. `{{ VAR }}` references are interpolated with the same semantics as [`ns_env_variables`](env_variables.html) (see that page for the full template syntax). A value inside `capability_secrets` or `user_secrets` is the secret itself and cannot be a `{{ secret(...) }}` or `{{ k8s.*(...) }}` template; move such a variable to `env_vars`.
6. `platform` gates templates: `{{ secret(...) }}` is only allowed on platforms that support secret refs, and `{{ k8s.*(...) }}` only on Kubernetes. An unknown platform is an error.

Each key is then classified as exactly one of: an **unmanaged secret** (`{{ secret(...) }}`), a **managed secret** (a secrets input, or promoted by interpolating a secret), a **Kubernetes ref**, or a **plain** variable.

## Example Usage

```hcl
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

# Iterate the plan-known keys from ns_env_layout; `secrets` is sensitive (and may be unknown at plan), so it cannot drive for_each.
resource "google_secret_manager_secret_version" "this" {
  for_each    = data.ns_env_layout.this.managed_secret_keys
  secret      = google_secret_manager_secret.this[each.key].id
  secret_data = data.ns_env_values.this.secrets[each.key]
}

data "ns_env_platform_data" "this" {
  values     = data.ns_env_values.this.platform_data
  secret_ids = { for key, secret in google_secret_manager_secret.this : key => secret.id }
}
```

With inputs:

```hcl
standard            = { NULLSTONE_ENV = "prod" }
capability_env      = [{ capability = "postgres0", name = "HOST", value = "db.internal" }]
capability_secrets  = [{ capability = "postgres0", name = "PASSWORD", value = "..." }]
capability_prefixes = { postgres0 = "PG_" }
user_env = {
  DATABASE_URL      = "postgres://app:{{ PG_PASSWORD }}@{{ PG_HOST }}/app"
  DATABASE_PASSWORD = "{{ secret(arn:aws:secretsmanager:us-east-1:123456789012:secret:db) }}"
  POD_IP            = "{{ k8s.field(v1, status.podIP) }}"
}
```

the outputs are:

- `env_variables` = `{ NULLSTONE_ENV = "prod", PG_HOST = "db.internal" }`
- `secrets` = `{ PG_PASSWORD = "...", DATABASE_URL = "postgres://app:...@db.internal/app" }` (promoted because it references a secret)
- `unmanaged_secret_refs` = `{ DATABASE_PASSWORD = "arn:aws:secretsmanager:..." }`
- `field_refs` = `{ POD_IP = { api_version = "v1", field_path = "status.podIP" } }`
- `sources` = `{ NULLSTONE_ENV = "standard", PG_HOST = "capability", PG_PASSWORD = "capability", DATABASE_URL = "user", ... }`
- `capabilities` = `{ PG_HOST = "postgres0", PG_PASSWORD = "postgres0" }`

## Arguments Reference

* `platform` - (Required) The runtime platform of the application: one of `aws_ecs`, `aws_batch`, `aws_lambda`, `aws_beanstalk`, `aws_ec2`, `aws_s3`, `aws_eks`, `gcp_gke`, `gcp_cloudrun`, `gcp_cloudfunctions`, `gcp_composer`, `gcp_gce`, `gcp_gcs`, `azure_aks`, `azure_container_app`, `azure_function`, `azure_app_service`, `azure_static_web_app`.
* `standard` - (Optional) Map of standard Nullstone environment variables (`NULLSTONE_*`). Lowest precedence.
* `cloud` - (Optional) Map of cloud platform environment variables (e.g. `AWS_REGION`, `GOOGLE_CLOUD_PROJECT`).
* `otel` - (Optional) Map of OpenTelemetry environment variables (`OTEL_*`).
* `capability_env` - (Optional) List of `{ capability, name, value }` objects emitted by capabilities (`local.capabilities.env`).
* `capability_secrets` - (Optional, Sensitive) List of `{ capability, name, value }` secret objects emitted by capabilities (`local.capabilities.secrets`).
* `capability_prefixes` - (Optional) Map of capability `capability` to the prefix applied to its variable and secret names. Every capability referenced by `capability_env` or `capability_secrets` must appear here.
* `user_env` - (Optional) Map of user-defined environment variables (`var.env_vars`). Highest precedence among plain layers; a key that any secrets input also sets stays a secret.
* `user_secrets` - (Optional, Sensitive) Map of user-defined secrets (`var.secrets`).

## Attributes Reference

* `id` - (Deprecated) A deterministic hash of the sorted keys. Present only for compatibility and should not be used.
* `env_variables` - Map of plain environment variables after interpolation. Secrets and template refs are excluded.
* `secrets` - (Sensitive) Map of managed secrets after interpolation.
* `unmanaged_secret_refs` - Map of keys that reference an existing secret (`{{ secret(...) }}`) to their secret reference.
* `field_refs` - Map of keys to `{ api_version, field_path }` Kubernetes `fieldRef` objects.
* `config_map_refs` - Map of keys to `{ key, name, optional }` Kubernetes `configMapKeyRef` objects.
* `resource_field_refs` - Map of keys to `{ resource, container, divisor }` Kubernetes `resourceFieldRef` objects.
* `file_key_refs` - Map of keys to `{ key, path, volume_name }` Kubernetes `fileKeyRef` objects.
* `managed_secret_keys` - Set of secret keys whose cloud secret the module must create.
* `unmanaged_secret_keys` - Set of keys that reference an existing cloud secret.
* `all_secret_keys` - Union of `managed_secret_keys` and `unmanaged_secret_keys`.
* `sources` - Map of every key to the layer that supplied it: `standard`, `cloud`, `otel`, `capability`, or `user`.
* `capabilities` - Map of capability-sourced keys to the `capability` of the capability that supplied them.
* `platform_data` - JSON-encoded `env` (version 1) platform data record. It is **not** sensitive: it carries the template, resolved value (plain variables only), sensitivity flag, ref, source, and capability of every variable, and never a secret value. Managed secrets carry no ref yet; pass it to [`ns_env_platform_data`](env_platform_data.html) with the ids of the secrets the module created.
