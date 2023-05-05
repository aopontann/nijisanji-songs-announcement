resource "google_cloud_scheduler_job" "nsa_bot_dev_tweet" {
  name             = "nsa-bot-dev-tweet"
  description      = "歌ってみた動画の告知を行う"
  schedule         = "*/5 * * * *"
  time_zone        = "Asia/Tokyo"
  attempt_deadline = "300s"
  paused = true

  retry_config {
    retry_count = 1
  }

  http_target {
    http_method = "POST"
    uri         = "${var.cloud_run_service_url_dev}/tweet"
  }
}

resource "google_cloud_scheduler_job" "nsa_bot_dev_check_new_video" {
  name             = "nsa-bot-dev-check-new-video"
  description      = "歌ってみた動画が新しくアップロードされているかチェックする"
  schedule         = "*/5 * * * *"
  time_zone        = "Asia/Tokyo"
  attempt_deadline = "300s"

  retry_config {
    retry_count = 1
  }

  http_target {
    http_method = "POST"
    uri         = "${var.cloud_run_service_url_dev}/check-new-video"
  }
}