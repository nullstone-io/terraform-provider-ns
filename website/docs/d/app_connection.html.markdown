---
layout: "ns"
page_title: "Nullstone: ns_app_connection"
sidebar_current: "docs-ns-app-connection"
description: |-
  Data source to configure a connection to another nullstone workspace through a capability's application.
---

# ns_app_connection

Data source to configure connection to another nullstone workspace through a capability's application.
See [capabilities](../index.html#capabilities) for more information.
Normally, `ns_connection` scopes the connection to the capability.
This stanza is a drop-in replacement that allows the capability to retrieve a connection of the application.

This stanza defines the name and type of connection we need.
During terraform execution, nullstone provides outputs from the connected workspace.

Plan Config affects this data source. See [the main provider documentation](../index.html) for more details.
The `capability_id` that is normally used in `ns_connection` is ignored in this data source.

## Example Usage

#### Basic example

```hcl
data "ns_app_connection" "network" {
  name     = "network"
  contract = "network/aws/vpc"
}
```


#### Example using `via`

The following example uses `via` to find the network for the owning application.
Since the application may not have a cluster, we specify `optional = true`.

```hcl
# top-level configuration
data "ns_app_connection" "cluster" {
  name     = "cluster"
  contract = "cluster/aws/ecs:fargate"
  optional = true
}

data "ns_app_connection" "network" {
  name     = "network"
  contract = "network/aws"
  via      = data.ns_connection.cluster.name
}
```

## Attributes Reference

* `name` - Name of nullstone connection.
* `type` - Type of nullstone module to make connection.
* `optional` - By default, if this connection has not been configured, this causes an error. Set to true to disable. (Default: `false`)
* `via` - Name of connection to satisfy this connection through. Typically, this is set to `data.ns_connection.other.name`.
* `workspace_id` - This refers to the workspace in nullstone. This follows the form `{stack_id}/{block_id}/{env_id}`.
- `outputs` - An object containing every root-level output in the remote state. This attribute is interchangeable for `data.terraform_remote_state.outputs`.
