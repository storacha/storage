resource "aws_ssm_parameter" "private_key" {
  name        = "/${var.app}/${terraform.workspace}/Secret/PRIVATE_KEY/value"
  description = "private key for the deployed environment"
  type        = "SecureString"
  value       = var.private_key

  tags = {
    environment = "production"
  }
}