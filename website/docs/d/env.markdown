---
layout: "ns"
page_title: "Nullstone: ns_env"
sidebar_current: "docs-ns-env"
description: |-
  Data source to read information about an environment.
---

# ns_env

Data source to read information an environment.

## Example Usage

#### Basic example

```hcl
data "ns_workspace" "this" {}

data "ns_env" "this" {
  stack_id = data.ns_workspace.this.stack_id
  env_id   = data.ns_workspace.this.env_id
}

locals {
  // env_type can be used to make decisions based on what type of environment
  env_type = data.ns_env.this.type
}
```

## Arguments Reference

* `stack_id` - (Required) ID of stack where the application resides in nullstone.
* `env_id` - (Required) ID of environment in nullstone.

## Attributes Reference

* `name` (string) - The name of the environment in Nullstone.
* `type` (string) - The type of environment in Nullstone. Possible values: `PipelineEnv`, `PreviewEnv`, `PreviewsSharedEnv`, `GlobalEnv`.
* `pipeline_order` (number) - [Only for `PipelineEnv`] A number that dictates which order the environment falls in the pipeline.
* `is_prod` (bool) - Indicates whether the environment is marked as a production environment.
