# Create Route53 zone for the app
resource "aws_route53_zone" "primary" {
  name = "${var.app}.${var.domain}"

  tags = {
    Name  = "${var.app}-zone"
    Owner = var.owner
    Team  = var.team
    Org   = var.org
  }
}

# Create DNS record for storage node
resource "aws_route53_record" "storage_node" {
  zone_id = aws_route53_zone.primary.zone_id
  name    = "${var.app}.${var.domain}"
  type    = "A"
  ttl     = "300"
  records = [aws_instance.storage_node.public_ip]

  depends_on = [aws_instance.storage_node]

  # Add a lifecycle policy to ensure DNS changes are applied properly
  lifecycle {
    create_before_destroy = true
  }
}