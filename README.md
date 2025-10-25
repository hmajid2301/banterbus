# Banter Bus

Banter Bus is a multiplayer web-based party game inspired by JackBox games. Players join lobbies to play social deduction and guessing games together in real-time.

<img src="docs/screens/main.png" alt="Home Page" width="1000">

## Table of Contents

- [Getting Started](#getting-started)
- [Technology Stack](#technology-stack)

## Getting Started

## Quick Start with Nix

```bash
# Clone the repository
git clone https://gitlab.com/hmajid2301/banterbus
cd banterbus

# Allow direnv to load the development environment
direnv allow

# Start the development server
task dev
```

The application will be available at `http://localhost:7331` (proxy for templ -> `:8080`).

### Preview Environments

When creating merge requests, you can deploy preview environments:

1. Add the `deploy/flux-preview` label to your merge request
2. CI will automatically build and deploy a preview environment
3. Access your preview at `https://mr-{MR_ID}.dev.banterbus.games`
4. Environment is automatically cleaned up when MR is merged/closed

## Technology Stack

### Backend
- **Go** - Core application language
  - Standard library HTTP server
  - gobwas/ws for WebSocket communication
  - SQLC for type-safe database queries
- **PostgreSQL** - Primary database for game state and user data
- **Redis** - Pub/Sub messaging for real-time events between players
- **templ** - Type-safe HTML templating

### Frontend

- **HTMX** - Dynamic HTML updates without JavaScript frameworks
- **Alpine.js** - Minimal JavaScript for interactive components
- **Tailwind CSS** - Utility-first styling

### Development Experience

- **Nix** - Reproducible development environments
  - gomod2nix for Go dependency management
  - Automated development shells
  - Pre-commit hooks for code quality
  - Docker image builds
- **Task** - Simple task runner (alternative to Make)
- **Air** - Live reload during development
- **SQLC** - Generate type-safe Go code from SQL

### Infrastructure & Deployment

- **Terraform** - Infrastructure as code
  - Single state file with workspace separation (dev/prod)
  - Automated DNS management via Cloudflare
  - Secret management through OpenBao
- **Kubernetes** - Container orchestration
  - GitOps deployment with Flux CD
  - Automatic scaling and health checks
- **GitLab CI/CD** - Continuous integration and deployment
  - Automated testing (unit, integration, e2e)
  - Docker image builds and registry management
  - **Preview Environments** - Automatic deployment for merge requests
    - Temporary environments for testing features
    - Automatic cleanup when MR is closed
    - URL format: `https://mr-{ID}.dev.banterbus.games`

### Monitoring & Observability

- **OpenTelemetry** - Distributed tracing and metrics
- **Grafana** - Metrics visualization and alerting
- **Prometheus** - Metrics collection and storage
- **Loki** - Log aggregation
- **Tempo** - Distributed tracing backend
