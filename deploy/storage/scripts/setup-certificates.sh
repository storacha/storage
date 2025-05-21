#!/bin/bash
set -e

# Stop nginx before certificate setup
systemctl stop nginx

# Get certificate with standalone mode (since nginx is stopped)
# shellcheck disable=SC2154
certbot certonly --standalone --agree-tos --non-interactive --email "admin@${domain}" -d "${app}.${domain}"

# Enable the site
ln -sf /etc/nginx/sites-available/storage /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Start nginx again
systemctl restart nginx

echo "Certificate setup completed"