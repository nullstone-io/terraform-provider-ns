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
The `capability_name` that is normally used in `ns_connection` is ignored in this data source.

## Local Module Development

This data source is only used in **capability modules**. It is not supported in app modules.

### Manually Changing Connections

You can manually override connection targets in `.nullstone/active-workspace.yml` using the root `connections` map.
Each key is the connection name (matching the `name` argument of a `ns_app_connection` data source).

`ns_app_connection` always reads from the root `connections` map, even when the provider specifies `capability_name`.
This is by design — it allows a capability to access the application's connections rather than its own.
For capability-scoped connections, use [`ns_connection`](connection.html) instead.

#### Example: Override an application connection from a capability

Given the following Terraform configuration in a capability module:

```hcl
data "ns_app_connection" "cluster" {
  name     = "cluster"
  contract = "cluster/aws/ecs:fargate"
}
```

Override where `cluster` resolves by editing `.nullstone/active-workspace.yml`.
Note that the connection goes in the root `connections` (not under `capabilities`):

```yaml
org_name: nullstone
stack_id: 100
stack_name: core
block_id: 101
block_name: my-app
block_ref: yellow-giraffe
env_id: 102
env_name: dev
connections:
  cluster:
    stack_id: 100
    block_id: 200
    block_name: my-other-cluster
capabilities:
  my-cap:
    connections:
      log-destination:
        stack_id: 100
        block_id: 400
        block_name: my-log-bucket
```

In this example, `data.ns_app_connection.cluster` resolves from the root `connections` (pointing at `my-other-cluster`),
while any `ns_connection` data sources using the `ns.cap` provider would resolve from `capabilities.my-cap.connections`.

#### Connection target fields

Each connection target supports the following fields:

* `stack_id` - (Required) The stack ID of the target workspace.
* `block_id` - (Required) The block ID of the target workspace.
* `block_name` - (Required) The block name of the target workspace.
* `env_id` - (Optional) The environment ID of the target workspace. If omitted, defaults to the current workspace's `env_id`.

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
  contract = "network/aws/vpc"
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
