// Alloy configuration for BanterBus
// This configuration scrapes metrics, logs, and traces and forwards them to Grafana Cloud

// Remote write for Prometheus metrics
prometheus.remote_write "grafana_cloud" {
  endpoint {
    url = "${prometheus_url}"

    basic_auth {
      username = "${stack_id}"
      password = "${api_key}"
    }
  }
}

// Loki logs forwarding
loki.write "grafana_cloud" {
  endpoint {
    url = "${loki_url}"

    basic_auth {
      username = "${stack_id}"  
      password = "${api_key}"
    }
  }
}

// OpenTelemetry traces forwarding
otelcol.exporter.otlp "grafana_cloud" {
  client {
    endpoint = "${tempo_url}"
    
    auth = otelcol.auth.basic.grafana_cloud.handler
  }
}

otelcol.auth.basic "grafana_cloud" {
  username = "${stack_id}"
  password = "${api_key}"
}

// Scrape BanterBus application metrics
prometheus.scrape "banterbus_app" {
  targets = [
    {"__address__" = "localhost:8080", "__metrics_path__" = "/metrics"},
  ]
  forward_to = [prometheus.remote_write.grafana_cloud.receiver]
}

// Collect application logs
loki.source.file "banterbus_logs" {
  targets = [
    {__path__ = "/var/log/banterbus/*.log", job = "banterbus"},
    {__path__ = "/var/log/banterbus/app.log", job = "banterbus-app"},
  ]
  forward_to = [loki.write.grafana_cloud.receiver]
}

// OpenTelemetry receiver for traces
otelcol.receiver.otlp "default" {
  grpc {
    endpoint = "0.0.0.0:4317"
  }

  http {
    endpoint = "0.0.0.0:4318"
  }

  output {
    traces = [otelcol.exporter.otlp.grafana_cloud.input]
  }
}

// Add instance labels
prometheus.scrape "node_exporter" {
  targets = [
    {"__address__" = "localhost:9100"},
  ]
  forward_to = [prometheus.remote_write.grafana_cloud.receiver]
}