#!/usr/bin/env bash
#set -euxo pipefail

# Clone & checkout pinned version
#cd /tmp
#git clone https://github.com/filecoin-project/curio.git
#cd curio
#git checkout "${CURIO_VERSION}"

# Build Lotus for target
#make clean "${CURIO_BUILD_TARGET}"

# Place the final binary somewhere globally accessible
# The actual compiled binary is typically `curio` in the project root
cd /tmp
#curl -o curio -L "https://bafybeibu66ysikactrnccubi2u2g3wsbvgtp4rqpxzzobdwa62y5lm2vye.ipfs.w3s.link/ipfs/bafybeibu66ysikactrnccubi2u2g3wsbvgtp4rqpxzzobdwa62y5lm2vye/curio"
curl -o curio -L "https://bafybeib2wjsd25tq4kbchjve4wqtqs7zs3qrhtwz2qo3mknoijgctbfjyy.ipfs.w3s.link/ipfs/bafybeib2wjsd25tq4kbchjve4wqtqs7zs3qrhtwz2qo3mknoijgctbfjyy/curio"

cp curio /usr/local/bin/curio
chmod 755 /usr/local/bin/curio

# Create a standard repo path for storing chain data
mkdir -p "/var/lib/curio"
chown ubuntu:ubuntu "/var/lib/curio"
