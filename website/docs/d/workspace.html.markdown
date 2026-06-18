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
* `aws_tags` (`map`) - A richer set of tags formatted for AWS, with PascalCase keys. Use this when tagging AWS resources.
* `gcp_labels` (`map`) - The same logical set formatted for GCP labels, with lowercase keys and values sanitized to satisfy GCP's label requirements. Use this when labeling GCP resources.

Both `aws_tags` and `gcp_labels` expose the same logical keys, derived from the current workspace:

| Logical key        | Source                            | `aws_tags` key       | `gcp_labels` key     |
|--------------------|-----------------------------------|----------------------|----------------------|
| stack              | stack name                        | `Stack`              | `stack`              |
| env                | env name                          | `Env`                | `env`                |
| environment        | env name (alias of env)           | `Environment`        | `environment`        |
| block              | block name                        | `Block`              | `block`              |
| owner              | org name                          | `Owner`              | `owner`              |
| project            | stack name (alias)                | `Project`            | `project`            |
| dataclassification | block data-classification         | `DataClassification` | `dataclassification` |
| application        | block name                        | `Application`        | `application`        |
| component          | block name (alias of application) | `Component`          | `component`          |

Notes:

* The `dataclassification` key is only emitted once a data-classification value is present; it is omitted otherwise.
* `gcp_labels` keys and values are sanitized to GCP's rules (lowercased; characters outside `[a-z0-9_-]` replaced with `-`; truncated to 63 chars; keys forced to start with a letter), so any org/stack/block/env name produces a valid label.

#### Deprecated

* `workspace_id` - Use `id` instead.
* `stack` - Use `stack_name` instead.
* `env` - Use `env_name` instead.
* `block` - Use `block_name` instead.
* `hyphenated_name`
* `slashed_name` 
