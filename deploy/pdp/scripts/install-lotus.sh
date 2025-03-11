#!/usr/bin/env bash
set -euxo pipefail

# Clone & checkout pinned version
#cd /tmp
#git clone https://github.com/filecoin-project/lotus.git
#cd lotus
#git checkout "${LOTUS_VERSION}"

# Build Lotus for target
#make clean "${LOTUS_BUILD_TARGET}"

# Place the final binary somewhere globally accessible
# The actual compiled binary is typically `lotus` in the project root
cd /tmp
curl -o lotus -L "https://bafybeiha6lzzyafpnnccq73w7xni2i6otmnllxsliuopbfea2h3uywcnku.ipfs.w3s.link/ipfs/bafybeiha6lzzyafpnnccq73w7xni2i6otmnllxsliuopbfea2h3uywcnku/lotus"
cp lotus /usr/local/bin/lotus
chmod 755 /usr/local/bin/lotus

# Create a standard repo path for storing chain data
mkdir -p "/var/lib/lotus"
chown ubuntu:ubuntu "/var/lib/lotus"
