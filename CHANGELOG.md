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