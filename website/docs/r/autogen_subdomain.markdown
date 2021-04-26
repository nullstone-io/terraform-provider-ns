---
layout: "ns"
page_title: "Nullstone: ns_autogen_subdomain"
sidebar_current: "docs-ns-autogen-subdomain"
description: |-
  Resource to configure an autogen subdomain in nullstone.
---

# ns_autogen_subdomain

Nullstone can create and manage auto-generated subdomains for users that look like `random-subdomain.nullstone.app`.
This resource allows users to delegate DNS records provisioned via Nullstone to a user-managed DNS zone.

## Example Usage

#### AWS Example

```hcl
resource "ns_autogen_subdomain" "autogen_subdomain" {
  subdomain_id = data.ns_subdomain.id
  env = data.ns_workspace.this.env
}

resource "aws_route53_zone" "this" {
  name = ns_autogen_subdomain.autogen_subdomain.fqdn
  tags = data.ns_workspace.this.tags
}

resource "ns_autogen_subdomain_delegation" "to_aws" {
  subdomain_id = data.ns_subdomain.id
  env = data.ns_workspace.this.env
  nameservers = aws_route53_zone.this.name_servers
}
```

## Argument Reference

- `subdomain_id` - (Required) Id of the subdomain that already exists in Nullstone system.
  The subdomain in Nullstone represents the block. This represents the subdomain name created for each environment.
- `env` - (Required) Name of the environment to create an autogen_subdomain in.

## Attributes Reference

* `dns_name` - The random portion of the autogen_subdomain. This is the first part of the fully-qualified domain name (FQDN) that comes before the domain name.
* `domain_name` - The name of the domain that Nullstone administers for this auto-generated subdomain.
* `fqdn` - The fully-qualified domain name (FQDN) for this auto-generated subdomain. It is composed as `{dns_name}.{domain_name}.`.
