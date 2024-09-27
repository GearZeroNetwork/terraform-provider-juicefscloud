# Terraform Provider for JuiceFS Cloud

This repository provides a Terraform provider for managing resources on JuiceFS cloud.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Building The Provider

1. Clone the repository:
    ```shell
    git clone <repository-url>
    ```
2. Enter the repository directory:
    ```shell
    cd <repository-directory>
    ```
3. Build the provider using the Go `install` command:
    ```shell
    go install
    ```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the Provider

To use the provider, add the provider configuration to your Terraform scripts. For example:

```hcl
provider "juicefs-cloud" {
  access_key = ..your.access.key..
  secret_key = ..your.secret.key..
}

data "juicefs-cloud_cloud" "aws" {
  name = "AWS"
}

data "juicefs-cloud_region" "aws_oregon" {
  cloud = data.juicefs-cloud_cloud.aws.id
  name = "us-west-2"
}

resource "juicefs-cloud_volume" "test" {
  name   = "example-tf-vol1"
  region = data.juicefs-cloud_region.aws_oregon.id
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
