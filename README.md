# Practice of Practice Games

Tools for resilience games.

> **Module In Development**: [_Wheel of Expertise_](https://www.popg.xyz/2024/05/23/wheelofexpertise/)

## How To Use

### Wheel of Expertise Server

Run this in the background or in its own terminal window:
```zsh
$ ./popg
2026/03/12 21:15:57 INFO Starting server on port 1234 token=559aae76-1e93-11f1-951f-decaae22ca3c
```

### Client CLI

Spin the wheel using the built-in CLI client:
```zsh
$ ./popg -url 'http://localhost:1234/s/test' -client 559aae76-1e93-11f1-951f-decaae22ca3c -json '{ "id": "331c7a00-1e70-11f1-85c8-53e0cbee6e98", "version": "0.1.0", "event_type": "spin.custom.wheel", "timestamp": "2026-03-12T10:00:00Z", "data": { "entries": ["one", "two", "three", "four", "five", "six", "seven"] } }'
2026/03/12 21:10:56 INFO Wheel spun!
five
```

### One-Off Artist Search

```zsh
$ ./popg -artist Autechre
Looking for Autechre in MusicBrainz database...
2026/03/12 21:14:20 INFO Data fetched status=200 url="https://musicbrainz.org/ws/2/artist/?query=artist:Autechre&fmt=json"
Found ::: Autechre
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
