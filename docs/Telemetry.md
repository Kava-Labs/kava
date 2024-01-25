# Telemetry

[example metrics emitted by Kava application](./example-prometheus-metrics.txt)

## Enabling Kava Metric Telemetry

To enable the kava app to emit telemetry during operation, update the relevant config values to enable metrics:

`config.toml`

```toml
[instrumentation]

# When true, Prometheus metrics are served under /metrics on
# PrometheusListenAddr.
# Check out the documentation for the list of available metrics.
prometheus = true

# Address to listen for Prometheus collector(s) connections
prometheus_listen_addr = ":8888"
```

`app.toml`

```toml
[telemetry]

# Prefixed with keys to separate services.
service-name = ""

# Enabled enables the application telemetry functionality. When enabled,
# an in-memory sink is also enabled by default. Operators may also enabled
# other sinks such as Prometheus.
enabled = true
```

Then restart the service with the updated settings

## Running local prometheus collector and grafana services

To collect app metrics and visualize them locally, you can run the prometheus collector and grafana services with docker compose from the repo root directory

```bash
docker compose -f prometheus.docker-compose.yml
```

Navigate to localhost:3000 to view the grafana unix

Login with `admin` as the username and `admin` as the password

Hook up grafana to the local prometheus collector by navigating to `http://localhost:3000/connections/datasources/new`, selecting prometheus, entering `http://prometheus:9090` for the url, and clicking `Save & test` at the bottom of the screen

See [grafana docs](https://grafana.com/docs/grafana/latest/dashboards/) for information on how to construct queries and build dashboards

### Collecting from local host

Update [prometheus config](../prometheus.yml) to collect metrics from your local source

```yaml
  metrics_path: /
  static_configs:
    - targets:
      - localhost:8888
```

### Collecting from remote host

Update the kava config on the host and restart using the instructions from `Enabling Kava Metric Emission`

Install [ngrok](https://ngrok.com/download) on the remote host

Run ngrok on the remote host to forward the prometheus metric port

```bash
ngrok http 8888
```

```yaml
scrape_configs:
- job_name: proxy
  scheme: https
  metrics_path: /
  static_configs:
    - targets:
      - 4efb-18-207-102-158.ngrok-free.app
```
