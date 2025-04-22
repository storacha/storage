#!/usr/bin/env bash
if [ ! -f /var/lib/curio/storage.json ]; then
  curio cli --machine=127.0.0.1:12300 \
      storage attach --init --seal --store /var/lib/curio/piecepark/;
fi