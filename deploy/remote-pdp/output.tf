output "instance_public_ip" {
  description = "Public IP of the instance"
  value       = aws_instance.pdp_node.public_ip
}
