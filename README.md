# Practice of Practice Games

[![Release](https://github.com/maroda/popg/actions/workflows/release.yml/badge.svg)](https://github.com/maroda/popg/actions/workflows/release.yml)

Tools for applied resilience engineering games.

## Wheel of Expertise server
> **Module In Development**: [_Wheel of Expertise_](https://www.popg.xyz/2024/05/23/wheelofexpertise/)
### Docker

This will start the game server and use the values in a local `.env` file (e.g. OTEL vars):
```zsh
docker run -d --env-file ./.env --rm --name popg -p 1234:1234 ghcr.io/maroda/popg:latest
```

> Orbstack default networking makes this available at: https://popg.orb.local

## Playing

- Participants can browse to the main front page and view the wheel, in the example above: `https://popg.orb.local`
- The Game Master sets up the server with a `GM_PASSWORD` environment variable.
When ready to facilitate a game, the GM browses to `https://popg.orb.local/?gm=<GM_PASSWORD>` and authenticates to access game controls.

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
