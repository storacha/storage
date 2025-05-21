output "instance_id" {
  description = "ID of the EC2 instance"
  value       = aws_instance.storage_node.id
}

output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.storage_node.public_ip
}

output "instance_dns" {
  description = "DNS name of the storage node"
  value       = aws_route53_record.pdp_record.fqdn
}

output "data_volume_id" {
  description = "ID of the EBS data volume"
  value       = aws_ebs_volume.storage_data.id
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.storage_vpc.id
}

output "subnet_id" {
  description = "ID of the subnet"
  value       = aws_subnet.storage_public_subnet.id
}

output "security_group_id" {
  description = "ID of the security group"
  value       = aws_security_group.storage_node_sg.id
}