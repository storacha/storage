#!/usr/bin/env bash
set -euxo pipefail

cd /opt
mkdir -p yugabyte && cd yugabyte
curl -OL "https://downloads.yugabyte.com/releases/${YUGABYTE_VERSION}/yugabyte-${YUGABYTE_VERSION}-b1-linux-x86_64.tar.gz"

tar xzf "yugabyte-${YUGABYTE_VERSION}-b1-linux-x86_64.tar.gz" --strip-components=1
ln -sf "/opt/yugabyte/bin/yugabyted" /usr/local/bin/yugabyted
