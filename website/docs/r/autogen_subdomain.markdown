---
layout: "ns"
page_title: "Nullstone: ns_autogen_subdomain"
sidebar_current: "docs-ns-autogen-subdomain"
description: |-
  Resource to configure an autogen subdomain in nullstone.
---

# ns_autogen_subdomain

Nullstone can generate autogen subdomains for users that look like `random-subdomain.nullstone.app`.
This resource allows users to delegate that subdomain to their own DNS zone.

## Example Usage

#### AWS Example

```hcl
resource "ns_autogen_subdomain" "subdomain" {
}

resource "aws_route53_zone" "this" {
  name = ns_autogen_subdomain.subdomain.fqdn
}

resource "ns_autogen_subdomain_delegation" "to_aws" {
  subdomain   = var.subdomain
  nameservers = aws_route53_zone.this.name_servers
}
```

## Argument Reference

## Attributes Reference

- `name` - Name of created auto-generated subdomain. This does not include the domain name (typically `nullstone.app`).
- `domain_name` - The domain name configured for this autogen subdomain. Typically this is `nullstone.app`.
- `fqdn` - The fully-qualified domain name for this auto-generated subdomain. This is composed of `{name}.{domain_name}.`.
