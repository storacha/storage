#!/bin/bash
set -ex

echo "Starting EBS volume setup script"
echo "Expected device name: ${device_name}"
echo "Mount point: ${mount_point}"

# Function to find the actual block device
find_device() {
  local expected_device="$${1}"
  local device=""
  
  # First check if the device exists as specified
  if [ -b "$${expected_device}" ]; then
    echo >&2 "Device exists at $${expected_device}"
    device="$${expected_device}"
    
  # Check for the xvd* naming convention
  elif [ -b "$${expected_device/\/dev\/sd/\/dev\/xvd}" ]; then
    device="$${expected_device/\/dev\/sd/\/dev\/xvd}"
    echo >&2 "Device found with xvd naming at $${device}"
    
  # For Nitro-based instances, use the AWS CLI to find the EBS volume
  else
    # Get instance ID from metadata service
    TOKEN=$(curl -s -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")
    INSTANCE_ID=$(curl -s -H "X-aws-ec2-metadata-token: $${TOKEN}" http://169.254.169.254/latest/meta-data/instance-id)
    
    echo >&2 "Running on instance $${INSTANCE_ID}, checking for attached EBS volumes"
    
    # Get list of non-root NVMe devices (fix whitespace handling)
    NVME_DEVICES=$(lsblk -d -o NAME,MOUNTPOINT -n | grep -v "/" | grep "nvme" | awk '{print $1}' | sort)
    
    if [ -n "$${NVME_DEVICES}" ]; then
      # If we have exactly one non-root NVMe device, use that
      NVME_COUNT=$(echo "$${NVME_DEVICES}" | wc -l | tr -d '[:space:]')
      if [ "$${NVME_COUNT}" -eq 1 ]; then
        device="/dev/$(echo "$${NVME_DEVICES}" | tr -d '[:space:]')"
        echo >&2 "Found single attached NVMe device: $${device}"
      else
        # Multiple NVMe devices - look for unmounted data volumes
        echo >&2 "Found multiple NVMe devices, trying to identify the data volume"
        
        # Find unmounted devices first
        UNMOUNTED_DEVICES=""
        for dev in $${NVME_DEVICES}; do
          if ! mount | grep -q "/dev/$${dev}"; then
            UNMOUNTED_DEVICES="$${UNMOUNTED_DEVICES} $${dev}"
          fi
        done
        
        # If we have exactly one unmounted device, use that
        UNMOUNTED_COUNT=$(echo "$${UNMOUNTED_DEVICES}" | wc -w | tr -d '[:space:]')
        if [ "$${UNMOUNTED_COUNT}" -eq 1 ]; then
          device="/dev/$(echo "$${UNMOUNTED_DEVICES}" | tr -d '[:space:]')"
          echo >&2 "Selected single unmounted NVMe device: $${device}"
        # If we have multiple unmounted devices, try to identify the data volume by size
        elif [ "$${UNMOUNTED_COUNT}" -gt 1 ]; then
          echo >&2 "Multiple unmounted devices found, checking sizes to identify data volume"
          # Look for a volume with the expected data volume size (approximately 100GB)
          for dev in $${UNMOUNTED_DEVICES}; do
            # Get size in GB, rounded
            SIZE_GB=$(lsblk -dno SIZE /dev/$${dev} | tr -d 'G' | awk '{print int($1+0.5)}')
            echo >&2 "Device /dev/$${dev} size: $${SIZE_GB}GB"
            # If close to 100GB (between 90GB and 110GB), it's likely our data volume
            if [ $${SIZE_GB} -ge 90 ] && [ $${SIZE_GB} -le 110 ]; then
              device="/dev/$${dev}"
              echo >&2 "Selected data volume by size (~100GB): $${device}"
              break
            fi
          done
          
          # If still no device found, just use the first unmounted device
          if [ -z "$${device}" ]; then
            device="/dev/$(echo "$${UNMOUNTED_DEVICES}" | awk '{print $1}')"
            echo >&2 "No volume matched expected size, using first unmounted device: $${device}"
          fi
        fi
      fi
    fi
    
    # If still no device, fall back to checking if mount point exists and is already mounted
    if [ -z "$${device}" ] && [ -d "${mount_point}" ] && mountpoint -q "${mount_point}"; then
      echo >&2 "Mount point ${mount_point} already exists and is mounted"
      # Get the device from mount information
      device=$(findmnt -n -o SOURCE --target "${mount_point}")
      echo >&2 "Device $${device} is already mounted at ${mount_point}"
    fi
  fi
  
  # Only output the device path, no debug information
  echo "$${device}"
}

# Find the actual device - the function only outputs the device path, all other output is to stderr
DEVICE=$(find_device "${device_name}")

if [ -z "$${DEVICE}" ]; then
  echo "ERROR: Could not find a suitable block device."
  echo "Available block devices:"
  lsblk
  exit 1
fi

echo "Using block device: $${DEVICE}"

# Check if the mount point already exists and is mounted
if [ -d "${mount_point}" ] && mountpoint -q "${mount_point}"; then
  echo "Mount point ${mount_point} is already mounted with device $(findmnt -n -o SOURCE --target "${mount_point}")"
  
  # Create storage directory if needed
  if [ ! -d "${mount_point}/storage" ]; then
    mkdir -p "${mount_point}/storage"
    chown -R ubuntu:ubuntu "${mount_point}/storage"
  fi
  
  echo "Disk is already set up and mounted."
  exit 0
fi

# Check if the device has a file system
if ! blkid "$${DEVICE}"; then
  echo "Creating ext4 filesystem on $${DEVICE}"
  mkfs.ext4 "$${DEVICE}"
fi

# Create mount point if it doesn't exist
if [ ! -d "${mount_point}" ]; then
  echo "Creating mount point directory: ${mount_point}"
  mkdir -p "${mount_point}"
fi

# Update fstab with UUID instead of device name for better reliability
UUID=$(blkid -s UUID -o value "$${DEVICE}")
if [ -n "$${UUID}" ]; then
  if ! grep -q "UUID=$${UUID}" /etc/fstab; then
    echo "Adding UUID=$${UUID} to /etc/fstab"
    echo "UUID=$${UUID} ${mount_point} ext4 defaults,nofail 0 2" >> /etc/fstab
  else
    echo "Entry for UUID=$${UUID} already exists in /etc/fstab"
  fi
else
  # Fallback to device name if UUID isn't available for some reason
  if ! grep -q "$${DEVICE}" /etc/fstab && ! grep -q "${mount_point}" /etc/fstab; then
    echo "Adding $${DEVICE} to /etc/fstab"
    echo "$${DEVICE} ${mount_point} ext4 defaults,nofail 0 2" >> /etc/fstab
  else
    echo "Entry for device or mount point already exists in /etc/fstab"
  fi
fi

# Mount the device
echo "Mounting ${mount_point}"
mount "${mount_point}"

# Create storage directory
echo "Setting up storage directory"
mkdir -p "${mount_point}/storage"
chown -R ubuntu:ubuntu "${mount_point}/storage"

echo "Disk setup complete. EBS volume mounted at ${mount_point}"
lsblk
df -h "${mount_point}"