terraform {
  required_providers {
    juicefs-cloud = {
      source  = "github.com/GearZeroNetwork/juicefscloud"
      version = "1.0.0"
    }
  }
}

provider "juicefscloud" {
  access_key = "<KEY>"
  secret_key = "<KEY>"
}

data "juicefscloud_cloud" "aws" {
  name = "AWS"
}

data "juicefscloud_region" "aws" {
  cloud = data.juicefscloud_cloud.aws.id
  name  = "us-west-2"
}

data "juicefscloud_volume" "test" {
  name = "test-tf-vol1"
}

output "debug_var" {
  value = data.juicefscloud_volume.test
}
