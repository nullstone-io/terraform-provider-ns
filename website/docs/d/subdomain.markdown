---
layout: "ns"
page_title: "Nullstone: ns_subdomain"
sidebar_current: "docs-ns-subdomain"
description: |-
Data source to read a nullstone subdomain.
---

# ns_domain

Nullstone can create and manage subdomains with a configured dns_name.
This data source allows users to read the dns_name in order to use the configured value when creating a dns zone.
The dns_name should be combined with the domain name in order to create a fqdn.

## Example Usage

#### Example

```hcl
data "ns_subdomain" "subdomain" {
  stack = data.ns_workspace.stack
  block = data.ns_workspace.block
}

output "subdomain_fqdn" {
  value = data.ns_subdomain.dns_name
}
```

## Argument Reference

- `stack` - (Required) Name of the stack that the subdomain exists in.
- `block` - (Required) Name of the subdomain/block that you are looking to fetch.

## Attributes Reference

* `id` - The id of the subdomain.
* `dns_name` - The subdomain name that has been configured for this domain. An example would be `api`.
