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

## Example Usage

```hcl
data "ns_connection" "network" {
  name = "network"
  type = "aws/network"
}
```

## Attributes Reference

* `name` - Name of nullstone connection.
* `type` - Type of nullstone module to make connection.
* `optional` - By default, if this connection has not been configured, this causes an error. Set to true to disable. (Default: `false`)
* `workspace` - Name of workspace for connection. (Environment variable: `NULLSTONE_CONNECTION_{name}`)
