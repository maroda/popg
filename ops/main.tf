terraform {
  required_version = "~> 1.1"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    dnsimple = {
      source  = "dnsimple/dnsimple"
      version = "~> 2.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}
