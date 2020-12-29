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

This data source is affected by Plan Config. See [the main provider documentation](../index.html) for more details.

## Example Usage

```hcl
# Simple example
data "ns_connection" "network" {
  name = "network"
  type = "network/aws"
}
```


```hcl
# Example using `via`
data "ns_connection" "cluster" {
  name = "cluster"
  type = "cluster/aws-fargate"
}

data "ns_connection" "network" {
  name = "network"
  type = "network/aws"
  via  = data.ns_connection.cluster.workspace
}
```

## Attributes Reference

* `name` - Name of nullstone connection.
* `type` - Type of nullstone module to make connection.
* `optional` - By default, if this connection has not been configured, this causes an error. Set to true to disable. (Default: `false`)
* `workspace` - This refers to the exact workspace used for state files in nullstone.
  This value will always be of the form `{stack}-{env}-{block}`.
  Utilizes environment variable `NULLSTONE_CONNECTION_{name}` to resolve.
  This value can be one of the following formats:
    * `{stack}.{env}.{block}`
    * `{env}.{block}` - (`stack` is pulled from the current workspace)
    * `{block}` - (`stack` and `env` are pulled from the current workspace)
* `via` - Name of workspace to satisfy this connection through. Typically, this is set to `data.ns_connection.other.workspace`.
- `outputs` - An object containing every root-level output in the remote state. This attribute is interchangeable for `data.terraform_remote_state.outputs`.
