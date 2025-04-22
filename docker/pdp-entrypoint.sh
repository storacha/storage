#!/usr/bin/env bash
set -euo pipefail

PDP_ADDRESS=$(storage wallet import /run/wallet.hex)

storage serve pdp \
--lotus-client-host="${LOTUS_CLIENT_HOST}" \
--eth-client-host="${LOTUS_CLIENT_HOST}" \
--pdp-address="${PDP_ADDRESS}"


