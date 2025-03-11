#!/usr/bin/env bash
set -euxo pipefail

cd /tmp
curl -OL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"

tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"

# really __really__ realllllly wish go wasn't such a baby about its environment variables.....
echo "export PATH=\$PATH:/usr/local/go/bin" >> /etc/profile
echo "export GOPATH=/root/go" >> /etc/profile
echo "export HOME=/root" >> /etc/profile