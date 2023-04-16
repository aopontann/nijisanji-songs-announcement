variable "project_id" {
  description = "プロジェクトのID"
  type        = string
}

variable "cloud_run_service_url_dev" {
  description = "開発用CloudRunサービスのURL"
  type        = string
}

variable "cloud_run_service_url_pro" {
  description = "本番用CloudRunサービスのURL"
  type        = string
}