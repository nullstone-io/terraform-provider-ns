## 0.6.20 (May 15, 2023)

BUG FIXES:

* Fixed issue reading environment type in `ns_env` data source.

## 0.6.19 (May 13, 2023)

FEATURES:

* Added `ns_env` data source to read information about an environment.

## 0.6.18 (Feb 16, 2023)

FEATURES:

* Added `ns_env_variables` which takes in all the environment variables and secrets, performing interpolation and returning the results.
* Added `ns_secret_keys` which takes in all the environment variables and secret keys.
  * This is useful when you need to do a for_each over the set of secret keys; this keeps the result static.

## 0.6.13 (Oct 20, 2022)

FEATURES:

* Added `commit_sha` attribute to `ns_app_env` data source.
* Updated `go-api-client` that includes improved error messages.

## 0.6.12 (Jul 01, 2022)

FEATURES:

* Added support for transitive connections using `ns_connection.via` attribute.

## 0.6.11 (Jun 10, 2022)

FEATURES:

* Added `ns_connection.contract` to migrate modules to use contract-based matching for dependent workspaces.
* Marked `ns_connection.type` for deprecation.
* Added `ns_connection.via` support for local module development.

## 0.6.10 (Mar 29, 2022)

FIXES:

* Upgraded [nullstone-io/go-api-client](https://github.com/nullstone-io/go-api-client) to resolve changes to Nullstone API upgrades for roles/permissions.

## 0.6.7 (Feb 25, 2022)

FIXES:

* Fixed `ns_subdomain` as a result of Nullstone API upgrades.

## 0.6.6 (Feb 24, 2022)

FIXES:

* Fixed loading of nullstone API address when using CLI profile on local machine.

## 0.6.5 (Feb 24, 2022)

FIXES:

* Fixed warning output if `ns_connection` cannot find state file for outputs.

## 0.6.4 (Feb 24, 2022)

FIXES:

* Fixed nil panic when `ns_app_env` is not found.

## 0.6.3 (Feb 24, 2022)

FEATURES:

* `ns_connection` will attempt to resolve through the plan config `.nullstone/active-workspace.yml` first.
This allows for local configuration of connections when iterating on modules.
* Updated Nullstone API endpoints to utilize new stack-based endpoints.

## 0.6.2 (Feb 24, 2022)

FIXES:

* Fixed nil panic in `ns_connection` when the workspace has not been created yet.

## 0.6.1 (Feb 22, 2022)

FIXES:

* Fixed loading of profile so that `NULLSTONE_ADDR` is honored.

## 0.6.0 (Feb 22, 2022)

FEATURES:

* Added support for CLI profiles when loading API configuration for nullstone resources.
* Added support for `.nullstone/active-workspace.yml` to replace `.nullstone.json`.

## 0.5.12 (Aug 18, 2021)

FEATURES:

* Update API client that includes consistent error messages from Nullstone APIs.

## 0.5.11 (Aug 03, 2021)

FIXES:

* Fail gracefully if nullstone fails to create/update autogen subdomain.

## 0.5.10 (Aug 02, 2021)

FEATURES:

* Added `ns_app_connection` to utilize a connection from the owning application.

## 0.5.9 (Jul 30, 2021)

FIXES:

* Fixed usage of capability ID in providers.

## 0.5.7 (Jul 10, 2021)

FEATURES:

* Added initial support for capabilities.

## 0.5.6 (Jun 21, 2021)

FIXES:

* Adjusting `autogen_subdomain` resource to use `env_id` instead of `env`.

## 0.5.5 (Jun 14, 2021)

FIXES:

* Remove unused APIs and fixed tests.

## 0.5.4 (May 24, 2021)

FIXES:

* Fixed issues when autogen_subdomain is missing.

## 0.5.3 (May 24, 2021)

FIXES:

* Fix nil pointer with autogen subdomain resource.

## 0.5.2 (May 21, 2021)

FIXES:

* Fix nil pointer with missing autogen subdomain.

## 0.5.1 (May 21, 2021)

FIXES:

* Updated `ns_app_env` docs to use `ns_workspace`.

## 0.5.0 (May 20, 2021)

FEATURES:

* Added support for `ns_workspace` `block_ref` to enable unique resource names across stacks and blocks.
* Updated usage of stack, block, and env in data sources and docs.
  * `stack_id` instead of `stack`.
  * `block_id` instead of `block`.
  * `app_id` instead of `app`.
  * `env_id` instead of `env`.

## 0.4.5 (Apr 27, 2021)

BREAKING CHANGES:

* Removed `ns_autogen_subdomain` data source.

## 0.4.4 (Apr 23, 2021)

FIXES:

* Fixed loading of `NULLSTONE_ADDR` for state backend access on `ns_connection`.

## 0.4.3 (Apr 23, 2021)

FEATURES:

* Added `ns_app_env` data source.

## 0.4.2 (Apr 23, 2021)

FEATURES:

* Added `ns_subdomain` data source.
* Added `ns_domain` data source.

## 0.4.1 (Apr 21, 2021)

FEATURES:

* Added `ns_autogen_subdomain` resource.
* Added loading of stack, env, block by env vars `NULLSTONE_STACK`, `NULLSTONE_ENV`, and `NULLSTONE_BLOCK`.

## 0.4.0 (Mar 08, 2021)

FEATURES:

* Added `ns_autogen_subdomain` data source.
* Added `ns_autogen_subdomain_delegation` resource.

## 0.3.1 (Feb 22, 2021)

FIXES:

* Updated docs to include authentication to Nullstone API.

## 0.3.0 (Feb 22, 2021)

FEATURES:

* Added support for `ns_connection` `via` to allow for pulling in a transitive connection.

## 0.2.3 (Jan 17, 2021)

FEATURES:

* Added extra debug logging to diagnose workspace loading.

## 0.2.2 (Dec 29, 2020)

FIXES:

* Fixed connection to Nullstone state backend.
* Fixed loading of connections to other workspaces.

## 0.2.1 (Dec 03, 2020)

FIXES:

* Fixed loading of stack, env, block in data sources.

## 0.2.0 (Nov 25, 2020)

FEATURES:

* Added `outputs` attribute to `ns_connection`.

## 0.1.1 (Nov 05, 2020)

FEATURES:

* Created data source `ns_connection`.

## 0.1.0 (Oct 28, 2020)

FEATURES:

* Created provider.
* Created data source `ns_workspace`.
