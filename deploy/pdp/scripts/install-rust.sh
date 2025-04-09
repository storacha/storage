#!/usr/bin/env bash
set -euxo pipefail

# Install rustup and the latest stable Rust toolchain
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | bash -s -- -y

# Add rustup to path
echo "export PATH=\$PATH:\$HOME/.cargo/bin" >> /etc/profile
