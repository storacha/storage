resource "aws_security_group" "pdp_node_sg" {
  name        = "pdp_node_sg"
  description = "Security group for PDP Node"
  vpc_id      = aws_vpc.pdp_vpc.id

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

data "aws_route53_zone" "pdp_zone" {
  name = "pdp.storacha.network"
}

resource "aws_route53_record" "pdp_record" {
  zone_id = data.aws_route53_zone.pdp_zone.zone_id
  name    = "pdp.storacha.network"  # Apex record
  type    = "A"
  ttl     = 300
  records = [aws_instance.pdp_node.public_ip]
}

resource "aws_instance" "pdp_node" {
  ami                     = "ami-00c257e12d6828491" #"ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20250115"
  instance_type           = var.instance_type
  subnet_id               = aws_subnet.pdp_public_subnet.id
  vpc_security_group_ids  = [aws_security_group.pdp_node_sg.id]

  key_name = "forrest-lotus-curio"

  user_data = local.cloud_init

  tags = {
    Name = var.app
  }

  root_block_device {
    volume_size = var.volume_size
    volume_type = var.volume_type
  }
}
