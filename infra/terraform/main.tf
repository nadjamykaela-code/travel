terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Enable required GCP APIs
resource "google_project_service" "cloudrun" {
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "cloudscheduler" {
  service            = "cloudscheduler.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "firestore" {
  service            = "firestore.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "secretmanager" {
  service            = "secretmanager.googleapis.com"
  disable_on_destroy = false
}

# Cloud Run API service
module "cloud_run_api" {
  source       = "./modules/cloud_run"
  service_name = "travel-bot-api"
  image        = var.api_image
  region       = var.region
  env_vars = {
    PORT           = "8080"
    GCP_PROJECT_ID = var.project_id
    LOG_LEVEL      = "info"
  }
  secrets       = var.api_env_secrets
  max_instances = 1
  cpu_always_on = false

  depends_on = [google_project_service.cloudrun]
}

# Cloud Run Worker service
module "cloud_run_worker" {
  source       = "./modules/cloud_run"
  service_name = "travel-bot-worker"
  image        = var.worker_image
  region       = var.region
  env_vars = {
    PORT           = "8080"
    GCP_PROJECT_ID = var.project_id
    LOG_LEVEL      = "info"
  }
  secrets       = var.worker_env_secrets
  max_instances = 1
  cpu_always_on = true

  depends_on = [google_project_service.cloudrun]
}

# IAM: allow Scheduler SA to invoke the worker
resource "google_cloud_run_service_iam_member" "worker_invoker" {
  project  = var.project_id
  location = var.region
  service  = module.cloud_run_worker.service_name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${var.scheduler_sa_email}"

  depends_on = [module.cloud_run_worker]
}

# Cloud Scheduler triggers the worker every 15 min
resource "google_cloud_scheduler_job" "trigger_worker" {
  name             = "trigger-travel-bot-worker"
  schedule         = "*/15 * * * *"
  time_zone        = "America/Sao_Paulo"
  attempt_deadline = "600s"

  http_target {
    uri         = module.cloud_run_worker.service_url
    http_method = "GET"
    oidc_token {
      service_account_email = var.scheduler_sa_email
    }
  }

  depends_on = [google_project_service.cloudscheduler]
}

# Firestore database (Native mode)
resource "google_firestore_database" "database" {
  project     = var.project_id
  name        = "(default)"
  location_id = var.region
  type        = "FIRESTORE_NATIVE"

  depends_on = [google_project_service.firestore]
}

# Firestore composite index for filter queries
resource "google_firestore_index" "filter_user_index" {
  project    = var.project_id
  collection = "filters"
  fields {
    field_path = "userId"
    order      = "ASCENDING"
  }
  fields {
    field_path = "isActive"
    order      = "ASCENDING"
  }

  depends_on = [google_firestore_database.database]
}
