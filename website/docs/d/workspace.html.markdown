---
layout: "ns"
page_title: "Nullstone: ns_workspace"
sidebar_current: "docs-ns-workspace"
description: |-
  Data source to configure module based on current nullstone workspace.
---

# ns_workspace

Data source to configure module based on current nullstone workspace.

This data source is affected by Plan Config. See [the main provider documentation](../index.html) for more details.

## Example Usage

```hcl
data "ns_workspace" "this" {
}
```

## Attributes Reference

* `workspace_id` - The fully qualified workspace ID. This follows the form `{stack_id}/{block_id}/{env_id}`.
* `stack_id` - Workspace stack ID. (Environment variable: `NULLSTONE_STACK_ID`)
* `env_id` - Workspace env ID. (Environment variable: `NULLSTONE_ENV_ID`)
* `block_id` - Workspace block ID. (Environment variable: `NULLSTONE_BLOCK_ID`)
* `tags` (`map`) - A default list of tags including all nullstone configuration for this workspace.
