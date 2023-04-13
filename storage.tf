resource "google_storage_bucket" "storage" {
  name          = "nsa-bot-storage-test"
  location      = "asia-northeast1"
  
  force_destroy = true
  uniform_bucket_level_access = true
  lifecycle_rule {
    condition {
      age = 5
    }
    action {
      type = "Delete"
    }
  }
}