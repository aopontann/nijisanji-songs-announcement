terraform {
  backend "gcs" {
    bucket = "nsa-terraform-state"
    prefix = "terraform/state"
  }
}