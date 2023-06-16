resource "google_secret_manager_secret" "secret-dsn" {
  secret_id = "DSN"

  replication {
    user_managed {
      replicas {
        location = "asia-northeast1"
      }
    }
  }
}

resource "google_secret_manager_secret" "secret-youtube-api-key" {
  secret_id = "YOUTUBE_API_KEY"

  replication {
    user_managed {
      replicas {
        location = "asia-northeast1"
      }
    }
  }
}

resource "google_secret_manager_secret" "secret-sntp" {
  secret_id = "SMTP_PASSWORD"

  replication {
    user_managed {
      replicas {
        location = "asia-northeast1"
      }
    }
  }
}

resource "google_secret_manager_secret" "secret-twitter-api-key" {
  secret_id = "TWITTER_API_KEY"

  replication {
    user_managed {
      replicas {
        location = "asia-northeast1"
      }
    }
  }
}

resource "google_secret_manager_secret" "secret-twitter-api-secret-key" {
  secret_id = "TWITTER_API_SECRET_KEY"

  replication {
    user_managed {
      replicas {
        location = "asia-northeast1"
      }
    }
  }
}

resource "google_secret_manager_secret" "secret-twitter-access-token" {
  secret_id = "TWITTER_ACCESS_TOKEN"

  replication {
    user_managed {
      replicas {
        location = "asia-northeast1"
      }
    }
  }
}

resource "google_secret_manager_secret" "secret-twitter-access-token-secret" {
  secret_id = "TWITTER_ACCESS_TOKEN_SECRET"

  replication {
    user_managed {
      replicas {
        location = "asia-northeast1"
      }
    }
  }
}

resource "google_service_account" "cloud-run-jobs" {
  account_id   = "cloud-run-jobs"
  display_name = "cloud-run-jobs"
  description  = "Cloud Run Jobs を起動し、Secret Manager にアクセスする"
}

resource "google_secret_manager_secret_iam_member" "member-dsn" {
  secret_id = google_secret_manager_secret.secret-dsn.secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.cloud-run-jobs.account_id}@${var.project_id}.iam.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "member-youtube-api-key" {
  secret_id = google_secret_manager_secret.secret-youtube-api-key.secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.cloud-run-jobs.account_id}@${var.project_id}.iam.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "member-secret-sntp" {
  secret_id = google_secret_manager_secret.secret-sntp.secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.cloud-run-jobs.account_id}@${var.project_id}.iam.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "member-twitter-api-key" {
  secret_id = google_secret_manager_secret.secret-twitter-api-key.secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.cloud-run-jobs.account_id}@${var.project_id}.iam.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "member-twitter-api-secret-key" {
  secret_id = google_secret_manager_secret.secret-twitter-api-secret-key.secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.cloud-run-jobs.account_id}@${var.project_id}.iam.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "member-twitter-access-token" {
  secret_id = google_secret_manager_secret.secret-twitter-access-token.secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.cloud-run-jobs.account_id}@${var.project_id}.iam.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "member-twitter-access-token-secret" {
  secret_id = google_secret_manager_secret.secret-twitter-access-token-secret.secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.cloud-run-jobs.account_id}@${var.project_id}.iam.gserviceaccount.com"
}