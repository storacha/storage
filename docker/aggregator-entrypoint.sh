#!/usr/bin/env bash
set -euo pipefail

storage start \
--key-file=/run/service.pem \
--curio-url="${PDP_SERVER_HOST}" \
--pdp-proofset="${PDP_PROOF_SET}"
