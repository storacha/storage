terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.73.0"
    }
  }
  backend "s3" {
    bucket = "storacha-terraform-state"
    key    = "storacha/storage/shared.tfstate"
    region = "us-west-2"
  }
}

provider "aws" {
  region              = var.region
  allowed_account_ids = var.allowed_account_ids
  default_tags {
    
    tags = {
      "Environment" = terraform.workspace
      "ManagedBy"   = "OpenTofu"
      Owner         = "storacha"
      Team          = "Storacha Engineer"
      Organization  = "Storacha"
      Project       = "${var.app}"
    }
  }
}

resource "aws_route53_zone" "primary" {
  name = "${var.app}.storacha.network"
}

output "primary_zone" {
  value = aws_route53_zone.primary
}