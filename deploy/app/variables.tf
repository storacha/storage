variable "app" {
  description = "name of the application"
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

variable "domain" {
  description = "domain name to use for the deployment (will be prefixed with app name)"
  type        = string
  default     = "storacha.network"
}

variable "use_pdp" {
  description = "is this a deployment that uses pdp"
  type        = bool
  default     = false
}

variable "region" {
  description = "aws region for all services"
  type        = string
  default     = "us-west-2"
}

variable "allowed_account_ids" {
  description = "account ids used for AWS"
  type        = list(string)
  default     = ["505595374361"]
}

variable "private_key" {
  description = "private_key for the peer for this deployment"
  type        = string
}

variable "indexing_service_did" {
  description = "did to use for the indexer"
  type        = string
  default     = "did:web:indexer.storacha.network"
}

variable "indexing_service_url" {
  description = "url to use for the indexer"
  type        = string
  default     = "https://indexer.storacha.network"
}

variable "indexing_service_proof" {
  description = "delegatin proof for the indexer"
  type        = string
}

variable "pdp_proofset" {
  description = "proofset used with pdp"
  type        = number
  default     = 0
}

variable "curio_url" {
  description = "url for the curio SP to communicate with"
  type        = string
  default     = ""
}

variable "did" {
  description = "DID for this deployment (did:web:... for example)"
  type        = string
  default     = "did:web:storage.storacha.network"
}

variable "access_logging_log_format" {
  type        = string
  description = "The log format to use for access logging."
  default     = "{\"apiId\": \"$context.apiId\", \"requestId\": \"$context.requestId\", \"extendedRequestId\": \"$context.extendedRequestId\", \"httpMethod\": \"$context.httpMethod\", \"path\": \"$context.path\", \"protocol\": \"$context.protocol\", \"requestTime\": \"$context.requestTime\", \"requestTimeEpoch\": \"$context.requestTimeEpoch\", \"status\": $context.status, \"responseLatency\": $context.responseLatency, \"responseLength\": $context.responseLength}"
}