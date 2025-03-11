#!/usr/bin/env bash
set -euxo pipefail

# Create a working directory for snapshots
mkdir -p "/opt/lotus-snapshots"
cd "/opt/lotus-snapshots"

# Download snapshot with aria2c
# Note: '-o' sets output filename
aria2c -x5 "${LOTUS_SNAPSHOT_URL}" -o snapshot.car.zst

# Import the snapshot into Lotus
lotus --repo="/var/lib/lotus" daemon --halt-after-import --import-snapshot "/opt/lotus-snapshots/snapshot.car.zst"

# Clean up snapshot to free disk
rm -f "/opt/lotus-snapshots/snapshot.car.zst"
