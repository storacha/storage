variable "app" {
  description = "The name of the application"
  type        = string
  default = "storage"
}

variable "use_pdp" {
  description = "is this a deployment that uses pdp"
  type        = bool
  default = false
}

variable "region" {
  description = "aws region for all services"
  type = string
  default = "us-west-2"
}

variable "allowed_account_ids" {
  description = "account ids used for AWS"
  type = list(string)
  default = ["505595374361"]
}

variable "private_key" {
  description = "private_key for the peer for this deployment"
  type = string
}

variable "indexing_service_did" {
  description = "did to use for the indexer"
  type = string
  default = "did:web:indexer.storacha.network"
}

variable "indexing_service_url" {
  description = "url to use for the indexer"
  type = string
  default = "https://indexer.storacha.network"
}

variable "pdp_proofset" {
  description = "proofset used with pdp"
  type = number
  default = 0
}

variable "curio_url" {
  description = "url for the curio SP to communicate with"
  type = string
  default = ""
}