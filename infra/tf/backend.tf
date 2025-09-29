terraform {
  backend "http" {
    # Backend configuration will be provided via -backend-config in CI
    # Using separate state files for environment separation:
    # - State: dev (for development environment)
    # - State: prod (for production environment)

    # The following variables will be set via CI/CD pipeline:
    # address = "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/terraform/state/${ENVIRONMENT}"
    # lock_address = "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/terraform/state/${ENVIRONMENT}/lock"
    # unlock_address = "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/terraform/state/${ENVIRONMENT}/lock"
    # username = "gitlab-ci-token"
    # password = "${CI_JOB_TOKEN}"
    # lock_method = "POST"
    # unlock_method = "DELETE"
    # retry_wait_min = 5
  }
}