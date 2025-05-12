#cloud-config

package_update: true
package_upgrade: true

packages:
  - git
  - curl
  - wget
  - jq
  - build-essential
  - lsb-release
  - nginx
  - certbot
  - python3-certbot-nginx
  - ripgrep

write_files:
  - path: /etc/nginx/sites-available/storage
    permissions: '0644'
    encoding: b64
    content: ${nginx_config}

  - path: /usr/local/bin/setup-certificates.sh
    permissions: '0755'
    encoding: b64
    content: ${setup_certificates_script}

  - path: /etc/systemd/system/storage.service
    permissions: '0644'
    encoding: b64
    content: ${storage_service_config}

  - path: /usr/local/bin/setup-disk.sh
    permissions: '0755'
    encoding: b64
    content: ${setup_disk_script}

  - path: /opt/service.pem
    permissions: '0600'
    encoding: b64
    content: ${base64encode(service_pem_content)}

runcmd:
  # Setup the data volume
  - chown ubuntu:ubuntu /opt/service.pem
  - /usr/local/bin/setup-disk.sh

  # Install storage binary from release
  - mkdir -p /tmp/storage-release
  - wget -O /tmp/storage-release/storage.tar.gz "https://github.com/storacha/storage/releases/download/v${storage_version}/storage_${storage_version}_linux_amd64.tar.gz"
  - tar -xzf /tmp/storage-release/storage.tar.gz -C /tmp/storage-release
  - mv /tmp/storage-release/storage /usr/local/bin/
  - chmod +x /usr/local/bin/storage

  # Set up Nginx and certificates
  - /usr/local/bin/setup-certificates.sh

  # Enable and start the storage service
  - systemctl daemon-reload
  - systemctl enable storage.service
  - systemctl start storage.service