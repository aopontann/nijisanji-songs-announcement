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

locals {
  github_repo_owner = "aopontann"
  github_repo_name  = "nijisanji-songs-announcement"
}

resource "google_service_account" "cloud-run-jobs-scheduler" {
  account_id   = "cloud-run-jobs-scheduler"
  display_name = "cloud-run-jobs-scheduler"
  description  = "Cloud Run Jobs を定期的に実行する"
}

resource "google_iam_workload_identity_pool" "main" {
  workload_identity_pool_id = "github"
  display_name              = "GitHub"
  description               = "GitHub Actions 用 Workload Identity Pool"
  disabled                  = false
}

resource "google_iam_workload_identity_pool_provider" "main" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.main.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"
  display_name                       = "GitHub"
  description                        = "GitHub Actions 用 Workload Identity Poolプロバイダ"
  disabled                           = false
  attribute_condition                = "assertion.repository_owner == \"${local.github_repo_owner}\""
  attribute_mapping = {
    "google.subject" = "assertion.repository"
  }
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}
