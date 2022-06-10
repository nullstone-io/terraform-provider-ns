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

## Argument Reference

There are no arguments to this data source.

## Attributes Reference

* `id` - The fully qualified workspace ID. This follows the form `{stack_id}/{block_id}/{env_id}`.
* `stack_id` - Workspace stack ID. (Environment variable: `NULLSTONE_STACK_ID`)
* `stack_name` - Workspace stack name. (Environment variable: `NULLSTONE_STACK_NAME`)
* `block_id` - Workspace block ID. (Environment variable: `NULLSTONE_BLOCK_ID`)
* `block_name` - Workspace block name. (Environment variable: `NULLSTONE_BLOCK_NAME`)
* `block_ref` - Workspace block reference. Unique name used for constructing resource names. (Environment variable: `NULLSTONE_BLOCK_REF`)
* `env_id` - Workspace environment ID. (Environment variable: `NULLSTONE_ENV_ID`)
* `env_name` - Workspace environment name. (Environment variable: `NULLSTONE_ENV_NAME`)
* `tags` (`map`) - A default list of tags including all nullstone configuration for this workspace.

#### Deprecated

* `workspace_id` - Use `id` instead.
* `stack` - Use `stack_name` instead.
* `env` - Use `env_name` instead.
* `block` - Use `block_name` instead.
* `hyphenated_name`
* `slashed_name` 
