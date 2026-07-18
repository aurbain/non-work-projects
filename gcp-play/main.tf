

# # Terraform configuration for creating a GCP Compute Engine instance with SSH access

# terraform {
#   required_version = ">= 1.0"
# }

# provider "google" {
#   project = var.project_id
#   region  = var.region
# }

# # Firewall rule to allow SSH on port 22 from anywhere
# resource "google_compute_firewall" "allow_ssh" {
#   name    = "allow-ssh"
#   network = var.network

#   allow {
#     protocol = "tcp"
#     ports    = ["22"]
#   }

#   source_ranges = ["0.0.0.0/0"]
# }

# # Compute Engine instance
# resource "google_compute_instance" "example" {
#   name         = "example-instance"
#   machine_type = var.machine_type
#   zone         = var.zone

#   boot_disk {
#     initialize_params {
#       image = var.image_family
#     }
#   }

#   network_interface {
#     network    = var.network
#     access_config {}
#   }

#   tags = ["ssh"]
# }
