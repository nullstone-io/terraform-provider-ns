---
layout: "ns"
page_title: "Nullstone: ns_agent"
sidebar_current: "docs-ns-agent"
description: |-
  Data source to read information about the Nullstone Agent.
---

# ns_agent

Data source to read information about the Nullstone Agent.

This is typically used to grant the Nullstone Agent access to cloud resources (for example, by referencing its AWS user ARN or GCP service account email in an IAM policy, or by allowing its IP addresses through a security group).

## Example Usage

```hcl
data "ns_agent" "this" {
}

locals {
  agent_aws_user_arn = data.ns_agent.this.aws_user_arn
}

# Allow the Nullstone Agent to reach a private endpoint
resource "aws_security_group_rule" "agent_https" {
  type              = "ingress"
  security_group_id = aws_security_group.this.id
  protocol          = "tcp"
  from_port         = 443
  to_port           = 443
  cidr_blocks       = [for ip in data.ns_agent.this.ip_addresses : "${ip}/32"]
}
```

## Argument Reference

There are no arguments to this data source.

## Attributes Reference

* `aws_account_id` (string) - The AWS Account ID of the Nullstone Agent.
* `aws_user_name` (string) - The AWS User Name of the Nullstone Agent.
* `aws_user_arn` (string) - The AWS User ARN of the Nullstone Agent.
* `gcp_project_id` (string) - The GCP Project ID of the Nullstone Agent.
* `gcp_service_account_email` (string) - The GCP Service Account Email of the Nullstone Agent.
* `ip_addresses` (list of strings) - The public IP addresses that the Nullstone Agent connects from. Use these to allow the agent through security groups and firewall rules.
