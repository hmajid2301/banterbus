# GitLab Webhook Configuration for Flux CD
# This webhook triggers Flux CD ResourceSet for preview environments

# Get GitLab secrets from OpenBao
data "vault_kv_secret_v2" "gitlab_secrets" {
  mount = "kv"
  name  = "apps/gitlab"
}

# Create GitLab project webhook
resource "gitlab_project_hook" "flux_webhook" {
  project = var.gitlab_project_id
  url     = "https://flux-webhook.homelab.haseebmajid.dev/webhook-receiver"
  token   = data.vault_kv_secret_v2.gitlab_secrets.data["webhook_token"]

  # Enable required events for Flux CD ResourceSet
  merge_requests_events = true
  push_events           = true

  # Optional: Enable other useful events
  issues_events              = false
  confidential_issues_events = false
  tag_push_events            = false
  note_events                = false
  job_events                 = false
  pipeline_events            = false
  wiki_page_events           = false
  deployment_events          = false
  releases_events            = false
  confidential_note_events   = false

  # Webhook configuration
  enable_ssl_verification   = true
  push_events_branch_filter = "" # All branches


}

