terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.73.0"
    }
    archive = {
      source = "hashicorp/archive"
    }
  }
  backend "s3" {
    bucket = "storacha-terraform-state"
    key    = "storacha/storage/terraform.tfstate"
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

provider "aws" {
  alias = "virginia"
  region = "us-east-1"
}