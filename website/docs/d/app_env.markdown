---
layout: "ns"
page_title: "Nullstone: ns_app_env"
sidebar_current: "docs-ns-app-env"
description: |-
  Data source to read information about an application in a specific environment.
---

# ns_app_env

Data source to read information about an application in a specific environment.

## Example Usage

#### Basic example

```hcl
data "ns_workspace" "this" {}

data "ns_app_env" "this" {
  app_id = data.ns_block.ns_workspace.block_id
  env_id = data.ns_env.ns_workspace.env_id
}

locals {
  // app_version is typically used to set the version on the service infrastructure
  app_version = data.ns_app_env.this.version
}
```

## Arguments Reference

* `app_id` - (Required) ID of application in nullstone. (Block ID of the App's block is the same as the Application ID)
* `env_id` - (Required) ID of environment in nullstone.

## Attributes Reference

* `version` - The version configured in nullstone for the application in the specific environment.
