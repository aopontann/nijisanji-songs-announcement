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
    uri         = "${var.cloud_run_service_url_dev}/ping"
  }
}