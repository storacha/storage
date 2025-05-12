# Security Group for Storage Node
resource "aws_security_group" "storage_node_sg" {
  name        = "${var.app}-storage-node-sg"
  description = "Security group for Storage Node"
  vpc_id      = aws_vpc.storage_vpc.id

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

  ingress {
    description = "Storage Service"
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    cidr_blocks = ["10.1.0.0/16"] # Internal VPC CIDR
    self        = true            # Allow from the security group itself
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = -1
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name  = "${var.app}-storage-node-sg"
    Owner = var.owner
    Team  = var.team
    Org   = var.org
  }
}

# IAM role for the EC2 instance
resource "aws_iam_role" "storage_node_role" {
  name = "${var.app}-storage-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name  = "${var.app}-storage-node-role"
    Owner = var.owner
    Team  = var.team
    Org   = var.org
  }
}

# Attach policies to the IAM role
resource "aws_iam_role_policy_attachment" "ssm_policy" {
  role       = aws_iam_role.storage_node_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# Instance profile for the EC2 instance
resource "aws_iam_instance_profile" "storage_node_profile" {
  name = "${var.app}-storage-node-profile"
  role = aws_iam_role.storage_node_role.name

  tags = {
    Name  = "${var.app}-storage-node-profile"
    Owner = var.owner
    Team  = var.team
    Org   = var.org
  }
}