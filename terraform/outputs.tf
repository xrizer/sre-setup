output "external_ip" {
  value = google_compute_instance.sre_demo.network_interface[0].access_config[0].nat_ip
}

output "urls" {
  value = {
    grafana    = "http://${google_compute_instance.sre_demo.network_interface[0].access_config[0].nat_ip}:3000"
    prometheus = "http://${google_compute_instance.sre_demo.network_interface[0].access_config[0].nat_ip}:9090"
    kibana     = "http://${google_compute_instance.sre_demo.network_interface[0].access_config[0].nat_ip}:5601"
    api        = "http://${google_compute_instance.sre_demo.network_interface[0].access_config[0].nat_ip}:8080/book"
  }
}
