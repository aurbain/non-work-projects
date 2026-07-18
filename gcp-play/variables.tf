variable "project_id" {
  description = "The ID of the GCP project"
  type        = string
}

variable "region" {
  description = "The region to deploy resources in"
  type        = string
  default     = "us-central1"
}

variable "network" {
  description = "The name of the VPC network"
  type        = string
}

variable "machine_type" {
  description = "The machine type for the instance"
  type        = string
  default     = "e2-medium"
}

variable "zone" {
  description = "The zone to deploy resources in"
  type        = string
  default     = "us-central1-a"
}

variable "image_family" {
  description = "The image family for the boot disk"
  type        = string
  default     = "debian-11"
}
