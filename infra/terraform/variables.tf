variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "api_image" {
  description = "Docker image for the API service"
  type        = string
}

variable "worker_image" {
  description = "Docker image for the worker service"
  type        = string
}

variable "scheduler_sa_email" {
  description = "Service account email for Cloud Scheduler OIDC"
  type        = string
}

variable "api_env_secrets" {
  description = "Sensitive env vars injected via Secret Manager for the API"
  type        = map(string)
  default     = {}
}

variable "worker_env_secrets" {
  description = "Sensitive env vars (API keys) for the worker"
  type        = map(string)
  default     = {}
}
