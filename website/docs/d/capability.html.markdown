---
layout: "ns"
page_title: "Nullstone: ns_capability"
sidebar_current: "docs-ns-capability"
description: |-
  Data source to read the current nullstone capability.
---

# ns_capability

Data source to read the current nullstone capability.

This data source is affected by Plan Config. See [the main provider documentation](../index.html) for more details.
It reads the `id` and `name` of the capability the module is currently deployed as.
These are empty/zero when the module is not running as a capability.

## Example Usage

```hcl
data "ns_capability" "this" {
}
```

## Argument Reference

There are no arguments to this data source.

## Attributes Reference

* `id` (number) - The ID of the capability this module is deployed as.
* `name` (string) - The name of the capability this module is deployed as.
