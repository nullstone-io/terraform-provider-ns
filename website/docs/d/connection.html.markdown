---
layout: "ns"
page_title: "Nullstone: ns_connection"
sidebar_current: "docs-ns-connection"
description: |-
  Data source to configure a connection to another nullstone workspace.
---

# ns_connection

Data source to configure connection to another nullstone workspace.
This stanza defines the name and type of connection we need.
During terraform execution, nullstone provides outputs from the connected workspace.

Plan Config affects this data source. See [the main provider documentation](../index.html) for more details.
Specific to this data source, if the provider specifies `capability_id`, 
this data source will pull connections from the capability rather than the owning application.

## Example Usage

#### Basic example

```hcl
data "ns_connection" "network" {
  name     = "network"
  contract = "network/aws/vpc"
}
```


#### Example using `via`

```hcl
# top-level configuration
data "ns_connection" "cluster" {
  name     = "cluster"
  contract = "cluster/aws/ecs:fargate"
}

data "ns_connection" "network" {
  name     = "network"
  contract = "network/aws"
  via      = data.ns_connection.cluster.name
}
```

```hcl
# cluster configuration
data "ns_connection" "network" {
  name     = "network"
  contract = "network/aws"
}
```

## Argument Reference

* `name` - (Required) Name of nullstone connection.
* `contract` - (Required) A contract name that enables matching of other workspaces by <category>[:<subcategory>]/<cloud-provider>/<platform>[:<subplatform>].
  This supports wildcard matching of any component in the contract. For example, `datastores/aws/postgres:*` will match any subplatform of `postgres`.
  See more at [https://docs.nullstone.io/extending/contracts/index.html](https://docs.nullstone.io/extending/contracts/index.html).
* `type` - (**DEPRECATED**) Type of nullstone module to make connection.
* `optional` - (Optional) By default, if this connection has not been configured, this causes an error. Set to true to disable. (Default: `false`)
* `via` - (Optional) Name of connection to satisfy this connection through. Typically, this is set to `data.ns_connection.other.name`.

## Attributes Reference

* `workspace_id` - This refers to the workspace in nullstone. This follows the form `{stack_id}/{block_id}/{env_id}`.
* `outputs` - An object containing every root-level output in the remote state. This attribute is interchangeable for `data.terraform_remote_state.outputs`.
