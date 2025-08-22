# BanterBus Infrastructure

This directory contains the Terraform configuration for managing BanterBus infrastructure across multiple cloud providers and services.

## Components

### Core Infrastructure
- **backend.tf** - Terraform HTTP backend configuration for GitLab state management
- **providers.tf** - Provider configurations for Cloudflare, Grafana, PostgreSQL, Sentry, and BetterStack
- **variables.tf** - Variable definitions for all required configuration

### Database
- **database.tf** - PostgreSQL database setup with proper user permissions and grants

### Monitoring & Observability
- **grafana_cloud.tf** - LGTM (Logs, Grafana, Traces, Metrics) Grafana Cloud stack setup
- **alloy.tf** - Alloy configuration for connecting to homelab instance at `alloy.homelab.haseebmajid.dev`
- **sentry.tf** - Sentry projects for frontend and backend error tracking with player_id support
- **betterstack.tf** - BetterStack monitoring for uptime, logs, metrics, and traces

### DNS
- **dns_records.tf** - Cloudflare DNS records for production and development environments

## Required Environment Variables

Set these variables before running Terraform:

```bash
# GitLab Remote State
export TF_VAR_remote_state_address="https://gitlab.com/api/v4/projects/PROJECT_ID/terraform/state/banterbus"
export TF_VAR_username="your-gitlab-username"
export TF_VAR_access_token="your-gitlab-access-token"

# Cloudflare
export TF_VAR_cloudflare_api_token="your-cloudflare-api-token"
export TF_VAR_zone_id="your-cloudflare-zone-id"

# Grafana Cloud
export TF_VAR_grafana_cloud_access_policy_token="your-grafana-cloud-token"

# PostgreSQL
export TF_VAR_postgres_host="your-postgres-host"
export TF_VAR_postgres_username="your-postgres-username"
export TF_VAR_postgres_password="your-postgres-password"

# Sentry
export TF_VAR_sentry_auth_token="your-sentry-auth-token"
export TF_VAR_sentry_organization="banterbus"

# BetterStack
export TF_VAR_betterstack_api_token="your-betterstack-api-token"
```

## Terraform Backend Initialization

Initialize with GitLab HTTP backend:

```bash
terraform init \
  -backend-config="address=$TF_VAR_remote_state_address" \
  -backend-config="lock_address=$TF_VAR_remote_state_address/lock" \
  -backend-config="unlock_address=$TF_VAR_remote_state_address/lock" \
  -backend-config="username=$TF_VAR_username" \
  -backend-config="password=$TF_VAR_access_token" \
  -backend-config="lock_method=POST" \
  -backend-config="unlock_method=DELETE" \
  -backend-config="retry_wait_min=5"
```

## Usage

1. **Plan the deployment:**
   ```bash
   terraform plan
   ```

2. **Apply the configuration:**
   ```bash
   terraform apply
   ```

3. **Get outputs:**
   ```bash
   terraform output
   ```

## Key Outputs

- `sentry_frontend_dsn` - Frontend Sentry DSN for JavaScript error tracking
- `postgres_connection_string` - PostgreSQL connection string (sensitive)
- `grafana_cloud_prometheus_url` - Prometheus remote write endpoint
- `grafana_cloud_loki_url` - Loki logs endpoint
- `grafana_cloud_tempo_url` - Tempo traces endpoint
- `alloy_config` - Generated Alloy configuration
- `betterstack_heartbeat_url` - Heartbeat URL for health monitoring

## Frontend Integration

For Sentry integration with player_id tracking, use the frontend DSN:

```javascript
import * as Sentry from "@sentry/browser";

Sentry.init({
  dsn: "your-frontend-dsn",
  integrations: [
    new Sentry.BrowserTracing(),
  ],
  tracesSampleRate: 1.0,
});

// Set player context
Sentry.setUser({
  id: playerId,
  username: playerName,
});

// Add custom tags
Sentry.setTag("game_session", lobbyId);
```

## Alloy Configuration

The generated Alloy configuration will be output and can be deployed to `alloy.homelab.haseebmajid.dev`. It includes:

- Prometheus metrics scraping from BanterBus application
- Log forwarding to Grafana Cloud Loki
- OpenTelemetry trace collection and forwarding
- Node exporter metrics collection

## Monitoring Setup

The infrastructure sets up comprehensive monitoring:

1. **Grafana Cloud** - Centralized metrics, logs, and traces
2. **Sentry** - Error tracking for both frontend and backend
3. **BetterStack** - Website uptime monitoring and alerting
4. **Alloy** - Collection agent connecting to homelab endpoint

## Development vs Production

DNS records are configured for both environments:
- Production: `banterbus.app` → `banterbus.fly.dev`
- Development: `dev.banterbus.app` → `banterbus-dev.fly.dev`

## Security Notes

- All sensitive outputs are marked as sensitive
- PostgreSQL passwords are properly handled
- API tokens are managed through environment variables
- Terraform state is stored securely in GitLab