---
layout: "ns"
page_title: "Nullstone: ns_block"
sidebar_current: "docs-ns-block"
description: |-
  Data source to configure module based on current nullstone block.
---

# ns_block

Data source to configure module based on current nullstone block.

This data source is affected by Plan Config. See [the main provider documentation](../index.html) for more details.

## Example Usage

```hcl
data "ns_block" "this" {
}
```

## Attributes Reference

* `id` (number) - The Block's ID in nullstone. (Environment variable: `NULLSTONE_BLOCK_ID`)
* `stack_id` (number) - The ID of the stack where the Block resides. (Environment variable: `NULLSTONE_STACK_ID`)
* `name` (string) - Name of Block. (Environment variable: `NULLSTONE_BLOCK_NAME`)
* `ref` (string) - Block's reference; a unique slug that is typically used to name real infrastructure. (Environment variable: `NULLSTONE_BLOCK_REF`)
