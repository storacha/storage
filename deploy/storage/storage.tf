# Create the EC2 instance
resource "aws_instance" "storage_node" {
  ami                    = "ami-00c257e12d6828491" #"ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20250115"
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.storage_public_subnet.id
  vpc_security_group_ids = [aws_security_group.storage_node_sg.id]
  key_name               = var.ssh_key_name
  iam_instance_profile   = aws_iam_instance_profile.storage_node_profile.name
  availability_zone      = "${var.region}a" # Same as the EBS volume

  user_data = templatefile("${path.module}/cloud-init.yaml.tpl", {
    setup_certificates_script      = local.setup_certificates_script
    setup_disk_script              = local.setup_disk_script
    storage_service_config         = local.storage_service_config
    nginx_config                   = local.nginx_config
    service_pem_content            = var.service_pem_content
    storage_version                = var.storage_version
    app                            = var.app
    domain                         = var.domain
  })
  user_data_replace_on_change = true

  root_block_device {
    volume_size = var.volume_size
    volume_type = var.volume_type

    tags = {
      Name  = "${var.app}-storage-root"
      Owner = var.owner
      Team  = var.team
      Org   = var.org
    }
  }

  tags = {
    Name  = "${var.app}-storage-node"
    Owner = var.owner
    Team  = var.team
    Org   = var.org
  }

  # Volume will be attached after instance creation
}

# Create an EBS volume for storage data
resource "aws_ebs_volume" "storage_data" {
  availability_zone = "${var.region}a"
  size              = var.data_volume_size
  type              = var.data_volume_type

  tags = {
    Name  = "${var.app}-storage-data"
    Owner = var.owner
    Team  = var.team
    Org   = var.org
  }
}

# Attach the EBS volume to the EC2 instance
resource "aws_volume_attachment" "storage_data" {
  device_name = var.ebs_device_name
  volume_id   = aws_ebs_volume.storage_data.id
  instance_id = aws_instance.storage_node.id
}