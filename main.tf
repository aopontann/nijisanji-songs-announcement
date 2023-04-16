terraform {
  backend "gcs" {
    bucket = "nsa-terraform-state"
    prefix = "terraform/state"
  }
}

provider "google" {
  project = var.project_id
  region  = "asia-northeast1"
}
