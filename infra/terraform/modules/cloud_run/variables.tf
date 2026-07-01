variable "service_name" {
  description = "Name of the Cloud Run service"
  type        = string
}

variable "image" {
  description = "Docker image URL (gcr.io or artifact registry)"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "env_vars" {
  description = "Non-sensitive environment variables"
  type        = map(string)
  default     = {}
}

variable "secrets" {
  description = "Sensitive env vars (injected via Secret Manager as env vars)"
  type        = map(string)
  default     = {}
}

variable "max_instances" {
  description = "Maximum number of concurrent instances (Free tier max = 1)"
  type        = number
  default     = 1
}

variable "cpu_always_on" {
  description = "Allocate CPU continuously (true for workers, false for request-based APIs)"
  type        = bool
  default     = false
}

variable "memory" {
  description = "Memory limit per instance"
  type        = string
  default     = "512Mi"
}

variable "cpu" {
  description = "CPU cores per instance"
  type        = string
  default     = "1"
}

variable "container_port" {
  description = "Port the container listens on"
  type        = number
  default     = 8080
}

variable "service_account_name" {
  description = "Custom service account email (empty = default compute SA)"
  type        = string
  default     = ""
}
