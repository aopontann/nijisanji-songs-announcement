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

resource "google_cloud_scheduler_job" "job" {
  paused           = true
  name             = "test-job-2"
  description      = "test http job"
  schedule         = "*/8 * * * *"
  time_zone        = "Asia/Tokyo"
  attempt_deadline = "320s"

  retry_config {
    retry_count = 1
  }

  http_target {
    http_method = "POST"
    uri         = "https://example.com/"
    body        = base64encode("{\"foo\":\"bar\"}")
  }
}