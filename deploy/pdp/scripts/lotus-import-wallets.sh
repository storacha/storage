#!/usr/bin/env bash
set -euxo pipefail

# Only import if the file exists and is not empty
if [[ -s "/opt/lotus_wallet_bls.json" ]]; then
  echo "Importing BLS wallet: /opt/lotus_wallet_bls.json"
  lotus --repo=/var/lib/lotus wallet import --format=json-lotus /opt/lotus_wallet_bls.json
else
  echo "No BLS wallet file to import."
fi

if [[ -s "/opt/lotus_wallet_delegated.json" ]]; then
  echo "Importing delegated wallet: /opt/lotus_wallet_delegated.json"
  lotus --repo=/var/lib/lotus wallet import --format=json-lotus /opt/lotus_wallet_delegated.json
else
  echo "No delegated wallet file to import."
fi