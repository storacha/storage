# Storage Node Deployment

This repository contains Terraform/OpenTofu configuration for deploying a storage node on AWS EC2.

## Overview

This project provisions:
- EC2 instance with attached EBS volume
- Networking configuration (VPC, subnets, security groups)
- DNS configuration
- NGINX as a reverse proxy
- Storage service as a systemd service

## Prerequisites

- Terraform 1.0+ or OpenTofu installed
- AWS credentials configured
- Registered domain name
- SSH key pair for EC2 access

## Deployment

1. Copy the example configuration:
   ```
   cp tofu.tfvars.example tofu.tfvars
   ```

2. Edit `tofu.tfvars` with your specific values

3. Initialize:
   ```
   tofu init
   ```

4. Plan the deployment:
   ```
   tofu plan -vars-file=tofu.tfvars
   ```

5. Apply:
   ```
   tofu apply -vars-file=tofu.tfvars
   ```

## Configuration

See `tofu.tfvars.example` for a complete example.

**Note on storage_principal_mapping**: This variable requires proper JSON encoding. Use `jsonencode()` to ensure proper formatting as shown in the example file.

## Outputs

- `instance_id`: EC2 instance ID
- `public_ip`: Instance public IP address
- `storage_endpoint`: Full URL of the storage service

## Maintenance

### SSH Access
```
ssh -i /path/to/ssh_key_name/file ubuntu@<public_ip>
```

### Update Configuration
```
tofu apply
```

### Destroy Resources
```
   tofu destroy -vars-file=tofu.tfvars
```