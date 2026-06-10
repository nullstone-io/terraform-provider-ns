---
layout: "ns"
page_title: "Nullstone: ns_secret_keys"
sidebar_current: "docs-ns-secret-keys"
description: |-
  Data source to compute the full set of secret keys after interpolation.
---

# ns_secret_keys

Data source to compute the full set of secret keys after interpolation.

It interpolates `{{ VAR }}` references between environment variables and secrets, then returns the keys of every value that resolves to a secret.
A variable is promoted to a secret when it references one of the `input_secret_keys`.
Unlike [`ns_env_variables`](env_variables.html), this data source only exposes the secret *keys* — never their values — so it can be used safely where the secret values are not yet known.

## Example Usage

```hcl
data "ns_secret_keys" "this" {
  input_env_variables = {
    DATABASE_URL = "{{ POSTGRES_URL }}"
    LOG_LEVEL    = "info"
  }
  input_secret_keys = [
    "POSTGRES_URL",
  ]
}
```

After interpolation:
- `secret_keys` = `["DATABASE_URL", "POSTGRES_URL"]` (`DATABASE_URL` is promoted because it references a secret; keys are sorted alphabetically)

## Arguments Reference

* `input_env_variables` - (Required) A map of environment variable names to their raw values before interpolation.
* `input_secret_keys` - (Required, Sensitive) A set of secret key names before interpolation.

Keys in `input_env_variables` and `input_secret_keys` must contain only letters, numbers, and underscores, and may not begin with a number.

## Attributes Reference

* `secret_keys` - The set of all keys that resolve to a secret. Includes the original `input_secret_keys` plus any variables promoted to secrets by referencing a secret.
