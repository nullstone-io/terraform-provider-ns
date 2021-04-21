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
  app_name = data.ns_workspace.this.block
  env_name = data.ns_workspace.this.env
}

locals {
  // app_version is typically used to set the version on the service infrastructure
  app_version = data.ns_app_env.this.version
}
```

## Arguments Reference

* `app` - (Required) Name of application in nullstone.
* `env` - (Required) Name of environment in nullstone.

## Attributes Reference

* `version` - The version configured in nullstone for the application in the specific environment.
