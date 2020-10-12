# Nullstone Terraform Provider

- Website: https://nullstone.io
- [![Gitter chat](https://badges.gitter.im/nullstone-io/Lobby.png)](https://gitter.im/nullstone-io/community)

## Overview

This provider enables several capabilities with nullstone.
1. This provider has several data sources to utilize standard block metadata in Terraform plans.
2. This provider has data sources for connecting to nullstone parent blocks.

## Available Resources

### `ns_workspace`

The nullstone workspace data source provides access to information about the current workspace.

```
data "ns_workspace" "this" {}
```

- Attributes
  - `org_name` - Nullstone Org name
  - `stack_name` - Nullstone Stack name
  - `block_name` - Nullstone Block name
  - `env_name` - Nullstone Environment name
  - `tags` - Workspace tags
    - `Stack` - Stack name
    - `Block` - Block name
    - `Env` - Environment name
  - `hyphenated_name` - Unique, interpolated name using `-` as delimiter
  - `slashed_name` - Unique, interpolated name using `/` as delimiter

### `ns_connection`

```
data "ns_connection" "network" {
  need = "network"
}
```

- Attributes
  - `outputs` - map of outputs from connected workspace via parent block + environment combination.
