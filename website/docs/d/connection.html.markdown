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

## Local Module Development

For local module development, download the [Nullstone CLI](https://docs.nullstone.io/getting-started/setup/install-configure-cli.html).
The `nullstone workspaces select` command prompts you when you define a new `ns_connection` in your module and a target workspace is not configured.
This configuration information is stored in `.nullstone/active-workspace.yml`; refer to the [main provider documentation](../index.html) for more information.

### Manually Changing Connections

You can manually override connection targets in `.nullstone/active-workspace.yml`.
The provider checks local connections first before fetching from the Nullstone API, so any connection defined here takes priority.

Where you define connections in the YAML depends on whether you are developing an **app module** or a **capability module**.

#### In an app module

When the provider does not specify `capability_name`, `ns_connection` reads from the root `connections` map.
Each key is the connection name (matching the `name` argument of a `ns_connection` data source).

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
  network:
    stack_id: 100
    block_id: 300
    block_name: dev-vpc
    env_id: 105
```

This causes `data.ns_connection.cluster` to resolve outputs from `my-other-cluster` and `data.ns_connection.network` to resolve from `dev-vpc`.

#### In a capability module

When the provider specifies `capability_name`, `ns_connection` reads from `capabilities.<capability_name>.connections` instead of the root `connections`.
This keeps capability connections scoped separately from application connections.

Given the following Terraform configuration in a capability module:

```hcl
provider "ns" {
  capability_name = "my-logging-cap"
  alias           = "cap"
}

data "ns_connection" "log_destination" {
  provider = ns.cap
  name     = "log-destination"
  contract = "datastore/aws/s3"
}
```

Override the connection in `.nullstone/active-workspace.yml`:

```yaml
org_name: nullstone
stack_id: 100
stack_name: core
block_id: 101
block_name: my-app
block_ref: yellow-giraffe
env_id: 102
env_name: dev
capabilities:
  my-logging-cap:
    connections:
      log-destination:
        stack_id: 100
        block_id: 400
        block_name: my-log-bucket
```

You can define both root connections and capability connections in the same file:

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
    block_name: my-cluster
capabilities:
  my-logging-cap:
    connections:
      log-destination:
        stack_id: 100
        block_id: 400
        block_name: my-log-bucket
  my-metrics-cap:
    connections:
      metrics-sink:
        stack_id: 100
        block_id: 500
        block_name: my-metrics-store
```

#### Connection target fields

Each connection target supports the following fields:

* `stack_id` - (Required) The stack ID of the target workspace.
* `block_id` - (Required) The block ID of the target workspace.
* `block_name` - (Required) The block name of the target workspace.
* `env_id` - (Optional) The environment ID of the target workspace. If omitted, defaults to the current workspace's `env_id`.

These local overrides also apply when resolving `via` references. If a `via` connection points through a locally overridden connection, the local target is used for the traversal.

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
  contract = "network/aws/vpc"
  via      = data.ns_connection.cluster.name
}
```

#### Example using `via` through another `via` connection

```hcl
data "ns_connection" "app" {
  name     = "app"
  contract = "app:container/aws/ecs"
}

data "ns_connection" "cluster" {
  name     = "cluster"
  contract = "cluster/aws/ecs"
  via      = data.ns_connection.app.name
}

data "ns_connection" "network" {
  name     = "network"
  contract = "network/aws/vpc"
  via      = "${data.ns_connection.app.name}/${data.ns_connection.cluster.name}"
}
```

```hcl
# cluster configuration
data "ns_connection" "network" {
  name     = "network"
  contract = "network/aws/vpc"
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
