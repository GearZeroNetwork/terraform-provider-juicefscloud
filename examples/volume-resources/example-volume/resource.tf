terraform {
  required_providers {
    juicefscloud = {
      source  = "GearZeroNetwork/juicefscloud"
      version = "1.0.0"
    }
  }
}

provider "juicefscloud" {
  access_key = "<KEY>"
  secret_key = "KEY"
}

data "juicefscloud_cloud" "aws" {
  name = "AWS"
}

data "juicefscloud_region" "aws_oregon" {
  cloud = data.juicefscloud_cloud.aws.id
  name  = "us-west-2"
}

resource "juicefscloud_volume" "test" {
  name   = "example-tf-vol1"
  region = data.juicefscloud_region.aws_oregon.id
}
