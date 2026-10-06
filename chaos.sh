#!/usr/bin/env bash
# Inject or clear failures during the demo.
# Usage: ./chaos.sh <VM_IP> errors|slow|both|off
set -euo pipefail
IP="${1:?VM IP required}"
case "${2:-off}" in
  errors) Q="error_rate=0.3&latency_ms=0" ;;
  slow)   Q="error_rate=0&latency_ms=800" ;;
  both)   Q="error_rate=0.3&latency_ms=800" ;;
  off)    Q="error_rate=0&latency_ms=0" ;;
  *) echo "mode must be errors|slow|both|off"; exit 1 ;;
esac
curl -s "http://$IP:8080/admin/chaos?$Q"; echo
