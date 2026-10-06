terraform {
  required_version = ">= 1.5"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

resource "google_compute_instance" "sre_demo" {
  name         = "sre-demo"
  machine_type = var.machine_type
  zone         = var.zone
  tags         = ["sre-demo"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 30
    }
  }

  network_interface {
    network = "default"
    access_config {} # ephemeral public IP
  }

  metadata_startup_script = file("${path.module}/startup.sh")

  labels = {
    purpose = "sre-demo"
  }
}

# Dashboards are only reachable from your own IP, not the whole internet.
resource "google_compute_firewall" "sre_demo_ui" {
  name    = "sre-demo-ui"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["3000", "5601", "8080", "9090", "9200"]
  }

  source_ranges = [var.my_ip_cidr]
  target_tags   = ["sre-demo"]
}
