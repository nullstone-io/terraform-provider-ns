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
data "ns_block" "this" {}

data "ns_domain" "domain" {
  stack_id = data.ns_block.this.stack_id
  block_id = data.ns_block.this.id
}

output "domain_fqdn" {
  value = data.ns_domain.domain.dns_name
}
```

## Argument Reference

- `stack_id` - (Required) ID of the stack that the domain exists in.
- `block_id` - (Required) ID of the domain/block that you are looking to fetch.

## Attributes Reference

* `dns_name` - The domain name that has been configured for this domain. An example would be `nullstone.io`.
