variable "app" {
  description = "The name of the application"
  type        = string
  default     = "storage"
}

variable "owner" {
  description = "owner of the resources"
  type        = string
  default     = "storacha"
}

variable "team" {
  description = "name of team managing working on the project"
  type        = string
  default     = "Storacha Engineer"
}

variable "org" {
  description = "name of the organization managing the project"
  type        = string
  default     = "Storacha"
}

variable "region" {
  description = "aws region for all services"
  type        = string
  default     = "us-west-2"
}

variable "domain" {
  description = "domain name to use for the deployment (will be prefixed with app name)"
  type        = string
  default     = "storacha.network"
}

variable "allowed_account_ids" {
  description = "account IDs used for AWS"
  type        = list(string)
  default     = ["0"]
}
