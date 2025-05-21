data "aws_route53_zone" "pdp_zone" {
  name = "${var.app}.${var.domain}"
}

resource "aws_route53_record" "pdp_record" {
  zone_id = data.aws_route53_zone.pdp_zone.zone_id
  name    = "${var.app}.${var.domain}" # Apex record
  type    = "A"
  ttl     = 300
  records = [aws_instance.storage_node.public_ip]
}