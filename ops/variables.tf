variable "aws_region" {
  type    = string
  default = "us-west-2"
}

variable "app" {
  type    = string
  default = "popg"
}

variable "port" {
  type    = number
  default = 1234
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "repository" {
  type    = string
  default = "ghcr.io/maroda/popg"
}

variable "release" {
  type    = string
  default = "latest"
}

variable "tcount" {
  type    = number
  default = 1
}

/* Secrets - populate terraform.tfvars to use locally */

variable "dnsapex" {
  description = "Domain apex"
  type        = string
}

variable "dnstoken" {
  description = "API Token for DNSimple access"
  type        = string
}

variable "dnsaccount" {
  description = "Account ID for DNSimple access"
  type        = number
}

variable "dnscertid" {
  description = "Certificate ID for domain apex"
  type        = number
}

variable "gm_password" {
  description = "GM password for wheel access"
  type        = string
  sensitive   = true
}