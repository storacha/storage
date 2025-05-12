# Main Terraform configuration file
# Acts as the entry point and references all resources defined in other files

locals {
  common_tags = {
    App       = var.app
    Owner     = var.owner
    Team      = var.team
    Org       = var.org
    ManagedBy = "Terraform"
  }

  # Read and encode external files for cloud-init
  setup_certificates_script = base64encode(templatefile("${path.module}/scripts/setup-certificates.sh", {
    app         = var.app
    domain      = var.domain
  }))

  setup_disk_script = base64encode(templatefile("${path.module}/scripts/setup-disk.sh", {
    device_name = var.ebs_device_name
    mount_point = var.data_mount_point
  }))

  storage_service_config = base64encode(templatefile("${path.module}/configs/storage.service", {
    storage_log_level             = var.storage_log_level
    storage_principal_mapping     = var.storage_principal_mapping
    storage_indexing_service_did  = var.storage_indexing_service_did
    storage_indexing_service_url  = var.storage_indexing_service_url
    storage_upload_service_did    = var.storage_upload_service_did
    storage_upload_service_url    = var.storage_upload_service_url
    storage_curio_url             = var.storage_curio_url
    storage_pdp_proofset          = var.storage_pdp_proofset
    storage_public_url            = var.storage_public_url == "" ? "https://${var.app}.${var.domain}" : var.storage_public_url
    storage_indexing_service_proof = var.storage_indexing_service_proof
  }))

  # Nginx config was already a template
  nginx_config = base64encode(templatefile("${path.module}/nginx.conf.tpl", {
    app         = var.app
    domain      = var.domain
  }))
}