variable "project_id" {
  description = "Your personal GCP project ID"
  type        = string
}

variable "region" {
  type    = string
  default = "asia-southeast2" # Jakarta
}

variable "zone" {
  type    = string
  default = "asia-southeast2-a"
}

variable "machine_type" {
  description = "8 GB RAM is enough for Elasticsearch + Kibana + the rest"
  type        = string
  default     = "e2-standard-2"
}

variable "my_ip_cidr" {
  description = "Your public IP in CIDR form, e.g. 203.0.113.10/32"
  type        = string
}
