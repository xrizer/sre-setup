#!/usr/bin/env bash
# Copies this project to the VM and starts the stack.
# Usage: GRAFANA_PASSWORD=yourpass ./deploy.sh
set -euo pipefail
ZONE="${ZONE:-asia-southeast2-a}"
VM="${VM:-sre-demo}"
GRAFANA_PASSWORD="${GRAFANA_PASSWORD:-changeme}"

cd "$(dirname "$0")"
COPYFILE_DISABLE=1 tar czf /tmp/sre-demo.tgz --exclude=./terraform --exclude=./.git .
gcloud compute scp --zone "$ZONE" /tmp/sre-demo.tgz "$VM":/tmp/sre-demo.tgz

gcloud compute ssh "$VM" --zone "$ZONE" --command "
  set -e
  until command -v docker >/dev/null 2>&1; do echo 'Waiting for Docker install...'; sleep 5; done
  mkdir -p ~/sre-demo && tar xzf /tmp/sre-demo.tgz -C ~/sre-demo
  cd ~/sre-demo
  sudo GRAFANA_PASSWORD='$GRAFANA_PASSWORD' docker compose up -d --build
  echo 'Waiting for Kibana (a few minutes on first start)...'
  until curl -s localhost:5601/api/status | grep -q '\"level\":\"available\"'; do sleep 10; done
  curl -s -X POST localhost:5601/api/data_views/data_view \
    -H 'kbn-xsrf: true' -H 'Content-Type: application/json' \
    -d '{\"override\":true,\"data_view\":{\"id\":\"booking-api-logs\",\"name\":\"booking-api logs\",\"title\":\"filebeat-*\",\"timeFieldName\":\"@timestamp\"}}' >/dev/null
  sudo docker compose ps
"
echo "Done. Run: cd terraform && terraform output urls"
