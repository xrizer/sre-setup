# Booking API Observability Stack (SRE Demo on GCP)

Simulasi API booking konsultasi dengan observability lengkap: **Prometheus** (metrics + alert), **Grafana** (dashboard 4 golden signals + SLO), dan **Elastic Stack** (Filebeat → Elasticsearch → Kibana) untuk investigasi log. Semua berjalan di satu VM Compute Engine via Docker Compose, dibuat dengan Terraform.

```
loadgen ──► booking-api (Go) ──► /metrics ──► Prometheus ──► Grafana
                  │                              └─► alert rules
                  └── JSON logs ──► Filebeat ──► Elasticsearch ──► Kibana
node-exporter ──► Prometheus (saturation: CPU, memory)
```

## Deploy (±15 menit)

Prasyarat: `gcloud` dan `terraform` terpasang, sudah `gcloud auth login` dan `gcloud auth application-default login`, pakai **project GCP pribadi** (bukan project kantor/klien).

```bash
gcloud config set project <PROJECT_ID>
gcloud services enable compute.googleapis.com

cd terraform
cp terraform.tfvars.example terraform.tfvars   # isi project_id dan my_ip_cidr
echo "$(curl -s ifconfig.me)/32"               # IP Anda untuk my_ip_cidr
terraform init && terraform apply

cd .. && GRAFANA_PASSWORD='passwordAnda' ./deploy.sh
cd terraform && terraform output urls
```

Pull image dan start Elasticsearch/Kibana pertama kali butuh beberapa menit. Login Grafana: `admin` / password Anda.

> Firewall hanya membuka port untuk IP Anda. Jika pindah jaringan (misal ke hotspot HP), update `my_ip_cidr` lalu `terraform apply` lagi.

## Akses dashboard

| Tool | URL | Fungsi |
|---|---|---|
| Grafana | `http://<IP>:3000` (user `admin`) | Dashboard metrics + panel error logs |
| Prometheus | `http://<IP>:9090/alerts` | Status alert rules |
| Kibana | `http://<IP>:5601` | Investigasi log (Discover) |
| Elasticsearch | `http://<IP>:9200/_cat/indices?v` | REST API penyimpanan log |

Jika port 5601/9090/9200 diblokir jaringan Anda (`ERR_CONNECTION_RESET`), buka tunnel dari Cloud Shell lalu pakai **Web Preview → Change port**:

```bash
gcloud compute ssh sre-demo --zone asia-southeast2-a -- -4 -N -L 5601:localhost:5601 -L 9200:localhost:9200 -L 9090:localhost:9090
```

Lupa password Grafana:

```bash
gcloud compute ssh sre-demo --zone asia-southeast2-a --command "cd ~/sre-demo && sudo docker compose exec -T grafana grafana cli admin reset-admin-password PasswordBaru123"
```

Gunakan password alfanumerik saja; karakter seperti `$ ! ' "` bisa rusak saat diteruskan lewat SSH.

## Skenario demo (±10 menit)

1. **Kondisi normal** — buka Grafana, dashboard *Booking API - Golden Signals*. Availability ~100%, error budget penuh, tidak ada alert.
2. **Inject insiden** — `./chaos.sh <IP> both` (30% error + 800 ms latency).
3. **Deteksi** — panel Errors dan Latency naik melewati threshold. Dalam ±1 menit, `HighErrorRate` dan `HighLatencyP95` berstatus FIRING di Prometheus (`/alerts`) dan panel *Firing alerts* jadi merah.
4. **Investigasi** — Kibana → Discover → data view *booking-api logs*, filter `app.level : "ERROR"`. Pesan `clinic-schedule-db: connection pool exhausted` dari komponen `clinic-scheduler` menunjukkan akar masalah.
   Error log yang sama juga tampil di panel *Error logs (Elasticsearch)* di dashboard Grafana. Elasticsearch bisa diquery langsung: `http://<IP>:9200/_cat/indices?v` atau `http://<IP>:9200/filebeat-*/_search?q=app.level:ERROR&size=5&pretty`.
5. **Mitigasi** — `./chaos.sh <IP> off`. Metrics pulih dan alert resolved.
6. **Tutup** — jelaskan SLO 99,5%, error budget yang terpakai, dan isi postmortem (timeline, root cause, action items).

## Poin untuk dijelaskan

- **Dua pilar observability**: Prometheus + Grafana untuk *metrics* (deteksi: ada masalah? seberapa parah?), Elasticsearch + Kibana untuk *logs* (investigasi: kenapa?). Elasticsearch adalah storage + search engine, Kibana adalah UI-nya, Filebeat pengirim log dari container. Pilar ketiga (traces) bisa ditambah dengan OpenTelemetry + Tempo/Jaeger.

- Alert berbasis gejala yang dirasakan user (error rate, latency), bukan hanya CPU.
- Metrics untuk mendeteksi, logs untuk menjelaskan penyebab.
- Semua as code: Terraform, Compose, alert rules, dashboard provisioning.
- Window alert dibuat 1 menit agar cepat untuk demo; di production pakai window lebih panjang atau multi-window burn-rate alert.
- Next step: Alertmanager ke Slack/PagerDuty, deploy ke GKE dengan Helm, Elasticsearch dengan security aktif.

## Bersihkan setelah interview

```bash
cd terraform && terraform destroy
```
