---
layout: "ns"
page_title: "Nullstone: ns_env"
sidebar_current: "docs-ns-env"
description: |-
  Data source to configure module based on current nullstone environment.
---

# ns_env

Data source to configure module based on current nullstone environment.

This data source is affected by Plan Config. See [the main provider documentation](../index.html) for more details.

## Example Usage

```hcl
data "ns_env" "this" {
}
```

## Attributes Reference

* `id` (number) - The Environment's ID in nullstone. (Environment variable: `NULLSTONE_ENV_ID`)
* `stack_id` (number) - The ID of the Stack where the Environment resides. (Environment variable: `NULLSTONE_STACK_ID`)
* `name` (string) - Name of the Environment. (Environment variable: `NULLSTONE_ENV_NAME`)
