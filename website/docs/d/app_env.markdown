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
  stack_id = data.ns_workspace.this.stack_id
  app_id   = data.ns_workspace.this.block_id
  env_id   = data.ns_workspace.this.env_id
}

locals {
  // app_version is typically used to set the version on the service infrastructure
  app_version = data.ns_app_env.this.version
}
```

## Arguments Reference

* `stack_id` - (Required) ID of stack where the application resides in nullstone.
* `app_id` - (Required) ID of application in nullstone. (Block ID of the App's block is the same as the Application ID)
* `env_id` - (Required) ID of environment in nullstone.

## Attributes Reference

* `version` - The version of the latest deployment of this application in the specific environment.
  The `NULLSTONE_DEPLOY_VERSION` environment variable will override this value.
* `commit_sha` - The commit SHA of the latest deployment of this application in this specific environment.
  The `NULLSTONE_DEPLOY_COMMIT_SHA` environment variable will override this value.
