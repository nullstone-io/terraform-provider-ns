---
layout: "ns"
page_title: "Nullstone: ns_env_platform_data"
sidebar_current: "docs-ns-env-platform-data"
description: |-
  Data source to complete the env platform data record with the ids of the managed secrets a module created.
---

# ns_env_platform_data

Data source that takes the `env` platform data record emitted by [`ns_env_values`](env_values.html) and attaches, to every **managed secret**,
the id of the cloud secret (or the Kubernetes `secretKeyRef`) the module created for it.
The result is the record Nullstone reads from state to show an application's environment, so every managed secret must receive exactly one id.

This data source makes no API calls. Nullstone reads `platform_data` directly from `data.ns_env_platform_data.*` in state;
no [`ns_platform_data`](platform_data.html) block is needed.

Rules:

- A key in `secret_ids` or `k8s_secret_refs` must exist in `values` and be a managed secret (sensitive, no ref). Any other key is an error.
- A key may appear in only one of the two maps.
- `k8s_secret_refs` is only valid when the record's `platform` supports Kubernetes refs (`aws_eks`, `gcp_gke`, `azure_aks`).
- Every managed secret in `values` must receive an id; otherwise the error `Managed secrets without an id: ...` lists the missing keys.

## Example Usage

The target module shape is: layout (keys, known at plan) → create one secret per `managed_secret_keys` → values → platform data.

```hcl
data "ns_env_layout" "this" {
  platform               = "aws_ecs"
  standard_keys          = keys(local.standard_env_vars)
  cloud_keys             = keys(local.aws_env_vars)
  otel_keys              = keys(local.otel_env_vars)
  capability_env_keys    = [for e in local.capabilities.env : { capability = e.capability, name = e.name }]
  capability_secret_keys = [for s in local.capabilities.secrets : { capability = s.capability, name = s.name }]
  capability_prefixes    = local.cap_prefixes
  user_env               = var.env_vars
  user_secret_keys       = nonsensitive(keys(var.secrets))
}

resource "aws_secretsmanager_secret" "this" {
  for_each = data.ns_env_layout.this.managed_secret_keys
  name     = "${local.resource_name}/${each.key}"
}

data "ns_env_values" "this" {
  platform            = "aws_ecs"
  standard            = local.standard_env_vars
  cloud               = local.aws_env_vars
  otel                = local.otel_env_vars
  capability_env      = local.capabilities.env
  capability_secrets  = local.capabilities.secrets
  capability_prefixes = local.cap_prefixes
  user_env            = var.env_vars
  user_secrets        = var.secrets
}

resource "aws_secretsmanager_secret_version" "this" {
  for_each      = data.ns_env_layout.this.managed_secret_keys
  secret_id     = aws_secretsmanager_secret.this[each.key].id
  secret_string = data.ns_env_values.this.secrets[each.key]
}

data "ns_env_platform_data" "this" {
  values     = data.ns_env_values.this.platform_data
  secret_ids = { for key, secret in aws_secretsmanager_secret.this : key => secret.arn }
}
```

On Kubernetes, pass the `secretKeyRef` of each managed secret instead:

```hcl
data "ns_env_platform_data" "this" {
  values = data.ns_env_values.this.platform_data
  k8s_secret_refs = {
    for key in data.ns_env_layout.this.managed_secret_keys :
    key => { name = kubernetes_secret.this.metadata[0].name, key = key }
  }
}
```

## Arguments Reference

* `values` - (Required) The env v1 platform data record from `ns_env_values` (`data.ns_env_values.<name>.platform_data`).
* `secret_ids` - (Optional) Map of managed secret key to the id of the cloud secret the module created for it (e.g. an AWS Secrets Manager ARN or a GCP secret resource name).
* `k8s_secret_refs` - (Optional) Map of managed secret key to `{ name, key }`, the Kubernetes `secretKeyRef` the module created for it. Only valid on Kubernetes platforms.

## Attributes Reference

* `id` - (Deprecated) A deterministic hash of the record's sorted keys. Present only for compatibility and should not be used.
* `platform_data` - JSON-encoded `env` (version 1) platform data record in which every managed secret carries a `secret` or `k8s_secret_key` ref. It is **not** sensitive and never contains a secret value.
