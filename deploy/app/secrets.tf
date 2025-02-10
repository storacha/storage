resource "aws_ssm_parameter" "private_key" {
  name        = "/${var.app}/${terraform.workspace}/Secret/PRIVATE_KEY/value"
  description = "private key for the deployed environment"
  type        = "SecureString"
  value       = var.private_key

  tags = {
    environment = "production"
  }
}

resource "aws_ssm_parameter" "external_blob_bucket_access_key_id" {
  count       = var.use_external_blob_bucket ? 1 : 0
  name        = "/${var.app}/${terraform.workspace}/Secret/EXTERNAL_BLOB_BUCKET_ACCESS_KEY_ID/value"
  description = "access key ID for an externally hosted blob bucket"
  type        = "SecureString"
  value       = var.external_blob_bucket_access_key_id

  tags = {
    environment = "production"
  }
}

resource "aws_ssm_parameter" "external_blob_bucket_secret_access_key" {
  count       = var.use_external_blob_bucket ? 1 : 0
  name        = "/${var.app}/${terraform.workspace}/Secret/EXTERNAL_BLOB_BUCKET_SECRET_ACCESS_KEY/value"
  description = "secret access key for an externally hosted blob bucket"
  type        = "SecureString"
  value       = var.external_blob_bucket_secret_access_key

  tags = {
    environment = "production"
  }
}
