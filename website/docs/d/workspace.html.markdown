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
* `data_classification` - The data classification (sensitivity) level configured for this workspace, e.g. `customer-content`. Empty when unclassified. (Environment variable: `NULLSTONE_DATA_CLASSIFICATION`)
* `aws_tags` (`map`) - A richer set of tags formatted for AWS, with PascalCase keys. Use this when tagging AWS resources.
* `gcp_labels` (`map`) - The same logical set formatted for GCP labels, with lowercase keys and values sanitized to satisfy GCP's label requirements. Use this when labeling GCP resources.
* `k8s_labels` (`map`) - The recommended Kubernetes labels (`app.kubernetes.io/*`) plus `nullstone.io/*` labels for this workspace, with values sanitized to satisfy Kubernetes' label value requirements. Use this when labeling Kubernetes resources.
* `azure_tags` (`map`) - The same logical set formatted for Azure tags, with PascalCase keys and values sanitized to satisfy Azure's tag requirements. Use this when tagging Azure resources.

`aws_tags`, `gcp_labels`, and `azure_tags` expose the same logical keys, derived from the current workspace:

| Logical key        | Source                            | `aws_tags` key       | `gcp_labels` key     | `azure_tags` key     |
|--------------------|-----------------------------------|----------------------|----------------------|----------------------|
| stack              | stack name                        | `Stack`              | `stack`              | `Stack`              |
| env                | env name                          | `Env`                | `env`                | `Env`                |
| environment        | env name (alias of env)           | `Environment`        | `environment`        | `Environment`        |
| block              | block name                        | `Block`              | `block`              | `Block`              |
| owner              | org name                          | `Owner`              | `owner`              | `Owner`              |
| project            | stack name (alias)                | `Project`            | `project`            | `Project`            |
| dataclassification | block data-classification         | `DataClassification` | `dataclassification` | `DataClassification` |
| application        | block name                        | `Application`        | `application`        | `Application`        |
| component          | block name (alias of application) | `Component`          | `component`          | `Component`          |

Notes:

* The `dataclassification` key is only emitted once a data-classification level is set on the block; it is omitted otherwise. Its value is the composite `<#>-<slug>` form (e.g. `2-customer-content`), which is valid across AWS tags, GCP labels, Azure tags, and Kubernetes label values.
* `gcp_labels` keys and values are sanitized to GCP's rules (lowercased; characters outside `[a-z0-9_-]` replaced with `-`; truncated to 63 chars; keys forced to start with a letter), so any org/stack/block/env name produces a valid label.
* `azure_tags` keys drop the characters Azure disallows in tag names (`< > % & \ ? /`) and are truncated to 512 chars; values are truncated to 256 chars.

### `k8s_labels`

`k8s_labels` exposes the following Kubernetes labels, derived from the current workspace:

| Label                          | Source                                                         |
|--------------------------------|----------------------------------------------------------------|
| `app.kubernetes.io/name`       | block name                                                     |
| `app.kubernetes.io/version`    | _(left blank; intended to be set by the consuming module)_     |
| `app.kubernetes.io/component`  | _(left blank; intended to be set by the consuming module)_     |
| `app.kubernetes.io/part-of`    | stack name                                                     |
| `app.kubernetes.io/managed-by` | `nullstone`                                                    |
| `nullstone.io/block`           | block name                                                     |
| `nullstone.io/stack`           | stack name                                                     |
| `nullstone.io/env`             | env name                                                       |
| `nullstone.io/block-ref`       | block reference                                                |
| `nullstone.io/data-classification` | block data-classification (composite `<#>-<slug>`)         |

Notes:

* Labels with a blank value (e.g. `app.kubernetes.io/version` and `app.kubernetes.io/component`) are omitted; the consuming module is expected to set them (typically via `merge`).
* `nullstone.io/data-classification` is only emitted once a data-classification level is set on the block; it is omitted otherwise.
* Values are sanitized to Kubernetes' label value rules (characters outside `[A-Za-z0-9_.-]` replaced with `-`; truncated to 63 chars; trimmed so they begin and end with an alphanumeric).

#### Deprecated

* `tags` (`map`) - A default list of tags including all nullstone configuration for this workspace. Use `aws_tags` or `gcp_labels` instead.
