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
  name = "network"
  type = "network/aws"
}
```


#### Example using `via`

```hcl
# top-level configuration
data "ns_connection" "cluster" {
  name = "cluster"
  type = "cluster/aws-fargate"
}

data "ns_connection" "network" {
  name = "network"
  type = "network/aws"
  via  = data.ns_connection.cluster.name
}
```

```hcl
# cluster configuration
data "ns_connection" "network" {
  name = "network"
  type = "network/aws"
}
```

## Attributes Reference

* `name` - Name of nullstone connection.
* `type` - Type of nullstone module to make connection.
* `optional` - By default, if this connection has not been configured, this causes an error. Set to true to disable. (Default: `false`)
* `via` - Name of connection to satisfy this connection through. Typically, this is set to `data.ns_connection.other.name`.
* `workspace_id` - This refers to the workspace in nullstone. This follows the form `{stack_id}/{block_id}/{env_id}`.
- `outputs` - An object containing every root-level output in the remote state. This attribute is interchangeable for `data.terraform_remote_state.outputs`.
