resource "google_cloud_run_service" "service" {
  name     = var.service_name
  location = var.region

  template {
    spec {
      containers {
        image = var.image
        ports {
          container_port = var.container_port
        }

        dynamic "env" {
          for_each = var.env_vars
          content {
            name  = env.key
            value = env.value
          }
        }

        dynamic "env" {
          for_each = var.secrets
          content {
            name = env.key
            value_source {
              secret_key_ref {
                secret  = env.value
                version = "latest"
              }
            }
          }
        }

        resources {
          limits = {
            memory = var.memory
            cpu    = var.cpu
          }
        }
      }

      container_concurrency = var.cpu_always_on ? 0 : 1
      timeout_seconds       = var.cpu_always_on ? 900 : 300
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale"         = tostring(var.max_instances)
        "autoscaling.knative.dev/minScale"         = "0"
        "run.googleapis.com/cpu-throttling"        = tostring(!var.cpu_always_on)
        "run.googleapis.com/execution-environment" = "gen2"
        "run.googleapis.com/startup-cpu-boost"     = "true"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  autogenerate_revision_name = true

  lifecycle {
    ignore_changes = [
      metadata[0].annotations["run.googleapis.com/operation-id"],
      metadata[0].annotations["run.googleapis.com/revision"],
      metadata[0].annotations["run.googleapis.com/serving-state"],
    ]
  }
}

output "service_url" {
  description = "URL of the deployed Cloud Run service"
  value       = google_cloud_run_service.service.status[0].url
}

output "service_name" {
  description = "Name of the Cloud Run service"
  value       = google_cloud_run_service.service.name
}
