terraform {
  required_version = "~> 1.9"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.50"
    }
  }
}

variable "region" {
  description = "AWS region to deploy resources into."
  type        = string
  default     = "ap-northeast-1"
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project = "code-stash"
      Managed = "terraform"
    }
  }
}

module "infra" {
  source = "./modules"
}
