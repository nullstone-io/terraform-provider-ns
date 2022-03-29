# Nullstone Terraform Provider

- Website: https://nullstone.io

## Overview

This provider enables several capabilities with nullstone.
1. This provider has several data sources to utilize standard block metadata in Terraform plans.
2. This provider has data sources for connecting to nullstone parent blocks.

 - A resource and a data source (`internal/provider/`),
 - Documentation (`website/`),
 - Miscellaneous meta files.

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) >= 1.x
-	[Go](https://golang.org/doc/install) >= 1.12

Building The Provider
---------------------

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command: 
```sh
$ go install
```

Adding Dependencies
---------------------

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.


Using the provider
----------------------

This provider is registered on the official Terraform registry. Follow the [docs](https://registry.terraform.io/providers/nullstone-io/ns/latest/docs) to use this provider.

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```