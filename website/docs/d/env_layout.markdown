---
layout: "ns"
page_title: "Nullstone: ns_env_layout"
sidebar_current: "docs-ns-env-layout"
description: |-
  Data source to classify the secrets of an application's layered environment at plan time.
---

# ns_env_layout

Data source that merges the layers of an application's environment (standard, cloud, otel, capability, user) using **keys only**
and reports which secrets the module must create (`managed_secret_keys`) versus which reference an existing cloud secret (`unmanaged_secret_keys`).

Because it never receives secret values, this data source is always resolvable at plan time, which lets a module `for_each` over `managed_secret_keys` to create cloud secrets.
Pair it with [`ns_env_values`](env_values.html), which takes the same layers with values and produces the final environment.

Both data sources apply the same resolution:

1. Layers are merged in precedence order (lowest to highest): `standard` < `cloud` < `otel` < `capability_env`/secrets < `user_env`/secrets. A later layer overrides an earlier one for the same key.
2. Capability keys are `capability_prefixes[capability] + name`. A duplicate final key within `capability_env` (or within the capability secrets) is an error.
3. Every key is validated (letters, numbers, underscore; cannot begin with a number).
4. `{{ VAR }}` references are interpolated with the same semantics as [`ns_env_variables`](env_variables.html): a variable that references a secret is promoted to a secret; `{{ secret(...) }}` and `{{ k8s.*(...) }}` templates are extracted into refs.
5. `platform` gates templates: `{{ secret(...) }}` is only allowed on platforms that support secret refs, and `{{ k8s.*(...) }}` only on Kubernetes. An unknown platform is an error.

## Example Usage

```hcl
data "ns_env_layout" "this" {
  platform               = "gcp_gke"
  standard               = local.standard_env_vars
  cloud                  = local.google_env_vars
  otel                   = local.otel_env_vars
  capability_env         = local.capabilities.env
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
```

## Arguments Reference

* `platform` - (Required) The runtime platform of the application: one of `aws_ecs`, `aws_batch`, `aws_lambda`, `aws_beanstalk`, `aws_ec2`, `aws_s3`, `aws_eks`, `gcp_gke`, `gcp_cloudrun`, `gcp_cloudfunctions`, `gcp_composer`, `gcp_gce`, `gcp_gcs`, `azure_aks`, `azure_container_app`, `azure_function`, `azure_app_service`, `azure_static_web_app`.
* `standard` - (Optional) Map of standard Nullstone environment variables (`NULLSTONE_*`). Lowest precedence.
* `cloud` - (Optional) Map of cloud platform environment variables (e.g. `AWS_REGION`, `GOOGLE_CLOUD_PROJECT`).
* `otel` - (Optional) Map of OpenTelemetry environment variables (`OTEL_*`).
* `capability_env` - (Optional) List of `{ capability, name, value }` objects emitted by capabilities (`local.capabilities.env`).
* `capability_secret_keys` - (Optional) List of `{ capability, name }` objects naming the secrets emitted by capabilities.
* `capability_prefixes` - (Optional) Map of capability `capability` to the prefix applied to its variable and secret names.
* `user_env` - (Optional) Map of user-defined environment variables (`var.env_vars`). Highest precedence.
* `user_secret_keys` - (Optional) Set of user-defined secret keys (`nonsensitive(keys(var.secrets))`).

## Attributes Reference

* `managed_secret_keys` - Set of secret keys whose cloud secret the module must create. Includes secrets inputs and any variable promoted by interpolating a secret.
* `unmanaged_secret_keys` - Set of keys that reference an existing cloud secret via `{{ secret(...) }}`.
* `all_secret_keys` - Union of `managed_secret_keys` and `unmanaged_secret_keys`.
* `sources` - Map of every key to the layer that supplied it: `standard`, `cloud`, `otel`, `capability`, or `user`.
* `capabilities` - Map of capability-sourced keys to the `capability` of the capability that supplied them.
