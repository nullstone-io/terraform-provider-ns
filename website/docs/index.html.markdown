---
layout: "ns"
page_title: "Provider: Nullstone"
sidebar_current: "docs-ns-index"
description: |-
  Terraform provider Nullstone.
---

# Scaffolding Provider

Use this paragraph to give a high-level overview of your provider, and any configuration it requires.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
provider "ns" {
}

# Example resource configuration
resource "ns_workspace" "example" {
  # ...
}
```

## Plan Config

When running inside a Nullstone runner, Nullstone will automatically configure the plan configuration all resources in this provider.
However, if you want to run locally, you may configure the current organization, workspace, and connections through a plan config.
This plan config is loaded by environment variables or from `.nullstone.json`.

The following is an example `.nullstone.json`.
```json
{
  "org": "nullstone",
  "stack": "core",
  "env": "prod",
  "block": "fargate0",
  "connections": {
    "network": "network0"
  }
}
```

The following environment file describes the same information as above.
```
NULLSTONE_ORG=nullstone
NULLSTONE_BLOCK=core
NULLSTONE_ENV=prod
NULLSTONE_STACK=fargate0
NULLSTONE_CONNECTION_network=network0
```
