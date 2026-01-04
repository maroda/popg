# Practice Automation

Tools for resilience games in development.

> _Current development is to retrieve data for game contents._

## How To Use

1. Clone
2. `go build .`
3. Run: `./popg <ARTIST>`

### Example

This commandline app takes one argument:
- **Artist name** to lookup in the MusicBrainz database

```zsh
$ ./popg Autechre
Looking for Autechre in MusicBrainz database...
2026/01/04 15:25:04 INFO Data fetched status=200 url="https://musicbrainz.org/ws/2/artist/?query=artist:Autechre&fmt=json"
Found it! ::: Autechre
```

## OpenTelemetry
Without any configuration, it will expect a local collector. If one is not running, this (harmless) error will show up in the logs. 
```log
2026/01/04 15:32:32 traces export: Post "https://localhost:4318/v1/traces": dial tcp [::1]:4318: connect: connection refused
```

### Grafana Cloud
Grafana Cloud expects the following settings, put these in `.env` and load before running with: `set -a; source ./.env`
```dotenv
OTEL_RESOURCE_ATTRIBUTES="service.name=popg-datafetcher"
OTEL_EXPORTER_OTLP_ENDPOINT="https://otlp-gateway-prod-us-west-0.grafana.net/otlp"
OTEL_EXPORTER_OTLP_HEADERS="Authorization=Basic <GRAFANA_CLOUD_TOKEN>"
```
