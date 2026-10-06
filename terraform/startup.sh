#!/bin/bash
set -e
# Elasticsearch needs a higher mmap count.
sysctl -w vm.max_map_count=262144
echo "vm.max_map_count=262144" > /etc/sysctl.d/99-elasticsearch.conf
# Docker Engine + Compose plugin.
if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
fi
