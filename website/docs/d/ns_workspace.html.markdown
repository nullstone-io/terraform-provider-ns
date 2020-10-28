---
layout: "ns"
page_title: "Nullstone: ns_workspace"
sidebar_current: "docs-ns-workspace"
description: |-
  Data source to configure module based on current nullstone workspace.
---

# ns_workspace

Data source to configure module based on current nullstone workspace.

## Example Usage

```hcl
data "ns_workspace" "this" {
}
```

## Attributes Reference

* `stack` - Workspace stack name. (Environment variable: `NULLSTONE_STACK`)
* `env` - Workspace env name. (Environment variable: `NULLSTONE_ENV`)
* `block` - Workspace block name. (Environment variable: `NULLSTONE_BLOCK`)
* `tags` (`map`) - A default list of tags including all nullstone configuration for this workspace.
* `hyphenated_name` - A standard, unique, computed name for the workspace using '-' as a delimiter that is typically used for resource names.
* `slashed_name` - A standard, unique, computed name for the workspace using '/' as a delimiter that is typically used for resource names.