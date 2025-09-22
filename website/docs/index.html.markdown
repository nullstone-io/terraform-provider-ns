---
layout: "ns"
page_title: "Provider: Nullstone"
sidebar_current: "docs-ns-index"
description: |-
  Terraform provider Nullstone.
---

# Nullstone Provider

[Nullstone](https://nullstone.io) is an extensible developer platform.

This provider serves as a bridge between Terraform modules and Nullstone apps, datastores, and domains.
Mostly, this allows for retrieval of Nullstone information, but you can also use this provider to configure domain information in Nullstone.

The Nullstone engine automatically configures this provider with the correct context (i.e. stack, environment, app/datastore/domain/block).
See documentation below for reference on local setup.

Use the navigation to the left to read about the available resources.

## Example Usage

```terraform
terraform {
  required_providers {
    ns = {
      source = "nullstone-io/ns"
    }
  }
}

data "ns_workspace" this {}
```

## Server Authentication

This provider communicates with Nullstone APIs using an API Key.
To configure, set the `NULLSTONE_API_KEY` environment variable.
Visit your [Nullstone profile](https://app.nullstone.io/profile) to create API Keys in Nullstone.

By default, this provider will use `https://api.nullstone.io`.
To override, set the `NULLSTONE_ADDR` environment variable.

Nullstone implements the state backend protocol for Terraform Cloud.
This provider will default the address to `https://api.nullstone.io`.
To override, set the `NULLSTONE_ADDR` environment variable.

A nullstone API key is necessary to communicate as well.
Set `NULSTONE_API_KEY` to your nullstone API key. 

## Plan Config

When running inside a Nullstone runner, Nullstone will automatically configure the plan configuration all resources in this provider.
However, if you want to run locally, you may configure the current organization and workspace through a plan config.
This terraform provider loads the plan config by environment variables or from `.nullstone/active-workspace.yml`.

The following is an example `.nullstone/active-workspace.yml`.
```yaml
org_name: nullstone
stack_id: 100
stack_name: core
block_id: 101
block_name: fargate0
block_ref: yellow-giraffe
env_id: 102
env_name: prod
```

The following environment file describes the same information as above.
```
NULLSTONE_ORG_NAME=nullstone
NULLSTONE_STACK_ID=100
NULLSTONE_STACK_NAME=fargate0
NULLSTONE_BLOCK_ID=101
NULLSTONE_BLOCK_NAME=core
NULLSTONE_BLOCK_REF=yellow-giraffe
NULLSTONE_ENV_ID=102
NULLSTONE_ENV_NAME=prod
```

## Capabilities

When constructing app modules that use capabilities, you can use an aliased provider to scope the module.
This ensures that connections defined within the module pull connection configuration from the capability rather than the application.

```terraform
provider "ns" {
  capability_name = "cap-name" 
  alias           = "cap_5"
}

module "cap_5" {
  source = "nullstone/capability"
  
  providers = {
    ns = ns.cap_5
  }
}
```
