variable "app" {
  description = "Name of the application"
  type        = string
  default     = "pdp"
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
  description = "AWS region for all resources"
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


variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.medium"
}

variable "volume_size" {
  description = "Size of the root volume in GB"
  type        = number
  default     = 20
}

variable "volume_type" {
  description = "Type of the root volume"
  type        = string
  default     = "gp3"
}

variable "data_volume_size" {
  description = "Size of the data volume in GB"
  type        = number
  default     = 100
}

variable "data_volume_type" {
  description = "Type of the data volume"
  type        = string
  default     = "gp3"
}

variable "ssh_key_name" {
  description = "Name of the SSH key pair for the EC2 instance"
  type        = string
}

variable "service_pem_content" {
  description = "Content of the service.pem file"
  type        = string
  sensitive   = true
}

variable "storage_principal_mapping" {
  description = "JSON encoded mapping of did:web to did:key"
  type        = string
}

variable "storage_indexing_service_did" {
  description = "DID to use for the indexer"
  type        = string
}

variable "storage_indexing_service_url" {
  description = "URL to use for the indexer"
  type        = string
}

variable "storage_indexing_service_proof" {
  description = "UCAN delegation to prove this storage node can access the indexer"
  type        = string
}

variable "storage_upload_service_did" {
  description = "DID to use for the upload service"
  type        = string
}

variable "storage_upload_service_url" {
  description = "URL to use for the upload service"
  type        = string
}

variable "storage_curio_url" {
  description = "URL for the Curio SP to communicate with"
  type        = string
}

variable "storage_pdp_proofset" {
  description = "Proofset used with PDP"
  type        = number
}

variable "storage_log_level" {
  description = "Log level for the storage node"
  type        = string
  default     = "info"
}

variable "storage_version" {
  description = "Version of the storage binary to install"
  type        = string
  default     = "0.0.3"
}

variable "storage_public_url" {
  description = "Public URL for the storage node"
  type        = string
  default     = ""
}

variable "ebs_device_name" {
  description = "Device name for EBS volume"
  type        = string
  default     = "/dev/sdf"
}

variable "data_mount_point" {
  description = "Mount point for the data volume"
  type        = string
  default     = "/data"
}