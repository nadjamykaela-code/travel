output "api_service_url" {
  description = "URL of the deployed API Cloud Run service"
  value       = module.cloud_run_api.service_url
}

output "worker_service_url" {
  description = "URL of the deployed Worker Cloud Run service"
  value       = module.cloud_run_worker.service_url
}

output "scheduler_job_name" {
  description = "Name of the Cloud Scheduler job that triggers the worker"
  value       = google_cloud_scheduler_job.trigger_worker.name
}

output "firestore_database" {
  description = "Name of the Firestore database"
  value       = google_firestore_database.database.name
}
