---
layout: "ns"
page_title: "Nullstone: ns_domain"
sidebar_current: "docs-ns-domain"
description: |-
  Data source to read a nullstone domain.
---

# ns_domain

Nullstone can create and manage domains with a configured dns_name.
This data source allows users to read the dns_name in order to use the configured value when creating a dns zone.

## Example Usage

#### Example

```hcl
data "ns_domain" "domain" {
  stack = data.ns_workspace.stack
  block = data.ns_workspace.block
}

output "domain_fqdn" {
  value = data.ns_domain.dns_name
}
```

## Argument Reference

- `stack` - (Required) Name of the stack that the domain exists in.
- `block` - (Required) Name of the domain/block that you are looking to fetch.

## Attributes Reference

* `dns_name` - The domain name that has been configured for this domain. An example would be `nullstone.io`.
