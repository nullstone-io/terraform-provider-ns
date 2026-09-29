---
layout: "ns"
page_title: "Nullstone: ns_platform_data"
sidebar_current: "docs-ns-datasource-platform-data"
description: |-
  No-op data source that persists platform data (data consumed only by Nullstone) into Terraform state.
---

# ns_platform_data

No-op data source that persists platform data into Terraform state.

Platform data is the *platform contract* of a module: data consumed only by Nullstone (UI, API, CLI, deployment tooling).
This is distinct from Terraform outputs, which are the module-to-module contract.
Nullstone extracts platform data from state each time a state version is saved and serves it by `kind`.

This data source makes no API calls. It validates the payload against the schema for `kind`/`version` and echoes the inputs into state.

For the `env` kind, prefer the layered data sources: [`ns_env_values`](env_values.html) emits a valid env record and [`ns_env_platform_data`](env_platform_data.html) completes it with the ids of the managed secrets. Nullstone reads `data.ns_env_platform_data.*` directly from state, so no `ns_platform_data` block is needed in that case. Use `ns_platform_data` for other kinds, or for an `env` record built by hand as below.

Validation follows a forward-compatibility rule:
- A `kind` or `version` that this version of the provider does not recognize produces a **warning** and is stored without validation, so an older provider never blocks a newer module.
- A recognized `kind`/`version` whose `data` is invalid produces an **error** at plan time.
- `data` must always be a JSON-encoded object.

## Example Usage

#### `env` kind, version 1

```hcl
data "ns_env_variables" "this" {
  input_env_variables = var.env_vars
  input_secrets       = var.secrets
}

data "ns_platform_data" "env" {
  kind    = "env"
  version = 1
  data = jsonencode({
    variables = merge(
      { for key, value in data.ns_env_variables.this.env_variables : key => {
        template = lookup(var.env_vars, key, "")
        value    = value
      } },
      { for key, value in nonsensitive(data.ns_env_variables.this.secrets) : key => {
        template  = lookup(var.secrets, key, "")
        sensitive = true
      } },
      { for key, ref in data.ns_env_variables.this.secret_refs : key => {
        sensitive = true
        ref       = { type = "secret", id = ref }
      } },
    )
  })
}
```

Every entry in `variables` is one of:

* a resolved value: `{ template?, value }`
* a sensitive value: `{ template?, sensitive = true, ref? }` (the value is never carried; `ref`, when present, is a `secret` or `k8s_secret_key` ref naming the managed secret)
* a reference resolved at runtime: `{ template?, ref = { type, ... } }`

`ref.type` is one of `secret` (`id`), `k8s_field` (`api_version?`, `field_path`), `k8s_config_map` (`name`, `key`, `optional?`),
`k8s_resource_field` (`resource`, `container?`, `divisor?`), `k8s_file_key` (`volume_name`, `path`, `key`), `k8s_secret_key` (`name`, `key`).
A `secret` or `k8s_secret_key` ref is sensitive by definition. Each entry may also carry `source` (`standard`, `cloud`, `otel`, `capability`, `user`) and, for `capability`, the `capability` name.

The `env` kind never carries secret values: a sensitive entry with a `value`, or an entry with both a `value` and a `ref`, is rejected at plan time.

## Arguments Reference

* `kind` - (Required) The kind of platform data (e.g. `env`). Determines the schema used to validate `data`.
* `version` - (Required) The schema version of `kind` that `data` conforms to. Must be a positive integer.
* `data` - (Required) A JSON-encoded object (typically produced with `jsonencode(...)`) that conforms to the schema for `kind`/`version`.

## Attributes Reference

* `id` - (Deprecated) `<kind>/<version>`. Present only for compatibility and should not be used.
* `kind` - The `kind` as provided.
* `version` - The `version` as provided.
* `data` - The `data` string as provided, unchanged.
