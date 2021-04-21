---
layout: "ns"
page_title: "Provider: Nullstone"
sidebar_current: "docs-ns-index"
description: |-
  Terraform provider Nullstone.
---

# Scaffolding Provider

Use this paragraph to give a high-level overview of your provider, and any configuration it requires.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
provider "ns" {
}

# Example resource configuration
resource "ns_workspace" "example" {
}
```

## Server Authentication

This provider communicates with Nullstone APIs using an API Key.
To configure, set the `NULLSTONE_API_KEY` environment variable.
Visit your [Nullstone profile](https://app.nullstone.io/profile) to create API Keys in Nullstone.

By default, this provider will use `https://api.nullstone.io`.
To override, set the `NULLSTONE_ADDR` environment variable.

Nullstone implements the state backend protocol for Terraform Cloud.
This provider will default the address to `https://api.nullstone.io`.
To override, set the `NULLSTONE_ADDR` environment variable.

A nullstone API key is necessary to communicate as well.
Set `NULSTONE_API_KEY` to your nullstone API key. 

## Plan Config

When running inside a Nullstone runner, Nullstone will automatically configure the plan configuration all resources in this provider.
However, if you want to run locally, you may configure the current organization and workspace through a plan config.
This plan config is loaded by environment variables or from `.nullstone.json`.

The following is an example `.nullstone.json`.
```json
{
  "orgName": "nullstone",
  "stackName": "core",
  "envName": "prod",
  "blockName": "fargate0"
}
```

The following environment file describes the same information as above.
```
NULLSTONE_ORG=nullstone
NULLSTONE_BLOCK=core
NULLSTONE_ENV=prod
NULLSTONE_STACK=fargate0
```
