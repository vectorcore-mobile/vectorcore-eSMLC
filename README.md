# VectorCore E-SMLC

This is a pure-Go recovery of the VectorCore E-SMLC SLs endpoint. It accepts
SCTP associations from VectorCore MME and implements the small LCS-AP APER
surface recovered from the prior service: Location Request/Response, failure
responses, Reset acknowledgement, and bounded Connection-Oriented Information
parsing. SCTP PPID is fixed at 29 and stream 0 is used for outbound messages.

`CGO_ENABLED=0 go build ./...` is supported. There is no native bridge,
foreign-function interface, native ASN.1 runtime, or external process.

## Status

The old service provided a fixed coordinate only. That path is retained as
`positioning.simulation`, disabled by default and marked in logs as
`result_source=simulation`. It is test-only and never substitutes for a real
position. The historical source contained no working LPP, LPPa, UE GNSS,
assistance-data, or estimate-extraction implementation. Connection-Oriented
Information is decoded in the deployed MME form: procedure 1,
Correlation-ID IE 2, payload type IE 15, and APDU IE 1. Outbound messages use
the TS 29.171 Release-16 `reject` procedure criticality; inbound `ignore` is
accepted only for the current MME compatibility deviation.
LPP payload type 0 is recognized and interpreted (ECID, OTDOA, and A-GNSS
method branches). ECID and OTDOA both support a real, catalog-backed
estimator: ECID returns a coarse surveyed serving-cell point, OTDOA solves
2D hyperbolic multilateration from RSTD measurements. A-GNSS is UE-based
(MS-based) and GPS only: the UE reports its own already-computed position
and this service validates and relays it, with no solver and no catalog of
its own. None of the three has an assistance-data source, so a real UE
cannot be expected to produce a usable report against this service without
one — see `docs/limitations.md`. MS-assisted A-GNSS (raw pseudorange
measurements, an E-SMLC-side GNSS solver) is unimplemented.

LPP*a* payload type 1 (E-SMLC&#8596;eNB, TS 36.455) is also recognized and
interpreted, but only the E-CID Measurement Initiation/Report procedure
family: it rides the same SLs connection as LPP rather than a separate
transport, and is opt-in (`positioning.lppa_ecid.enabled`, default `false`)
and top-priority over every UE-based method when enabled, since it needs no
UE round trip at all. `LPPaECIDEstimator` prefers the eNB's own
self-reported antenna position and otherwise falls back to the same cell
catalog ECID uses. Every other LPPa elementary procedure remains
unimplemented — see `docs/limitations.md`.

An optional observability HTTP listener (`observability.enabled`, disabled
by default, loopback-only by default) exposes Prometheus-format `/metrics`,
`/healthz`, `/readyz`, and `POST /admin/reload-catalog` — the catalog reload
path already documented above as "explicit/administrative" but, before this,
had no actual administrative caller in the shipped binary. GMLC/SLg and
subscriber privacy/authorization are confirmed out of scope for this
component (TS 23.271 places those checks entirely at the GMLC/MME, upstream
of every request this service ever sees) — see `docs/limitations.md` and
`docs/roadmap.md`.

Transport security (NDS/IP per TS 33.401/33.210) is correctly delegated to
the network layer below SCTP, not implemented by this codebase — confirmed
by reading TS 33.401 clause 11, not assumed. Multi-MME concurrency has been
audited and race-tested: every piece of live request state is keyed by
association and correlation together, so concurrent MMEs reusing identical
correlation bytes cannot collide — see `docs/architecture.md`.

## Supported Location Types

Five methods can produce a location estimate, each gated by
`positioning.<method>.enabled` in `config/esmlc.yaml` (see
`config/esmlc.yaml.example`). Each populates `GeographicEstimate.Source`
with a distinct value, so the origin of every estimate is traceable —
including in the JSON log (below).

| Method | Config key | How the estimate is derived | `EstimateSource` |
|---|---|---|---|
| E-CID | `positioning.ecid.enabled` | Coarse surveyed serving-cell point, looked up by E-CGI in the operator-maintained catalog (`ecid.cell_data_file`) | `EstimateSourceAuthoritativeServingCell` |
| OTDOA | `positioning.otdoa.enabled` | 2D hyperbolic multilateration solved from UE-reported RSTD measurements, against the same cell catalog's reference/neighbour positions | `EstimateSourceOTDOAMultilateration` |
| A-GNSS | `positioning.agnss.enabled` | UE-based (MS-based) GPS only: the UE reports its own already-computed position; this service validates and relays it, with no solver of its own | `EstimateSourceAGNSSUEReported` |
| LPPa E-CID | `positioning.lppa_ecid.enabled` | eNB's self-reported antenna position over LPPa (TS 36.455), needing no UE round trip; falls back to the E-CID catalog when the eNB has none | `EstimateSourceLPPaAccessPointPosition` |
| Simulation | `positioning.simulation.enabled` | Fixed, operator-configured coordinate for testing only; never installed unless explicitly enabled, and never substitutes for a real estimate | `EstimateSourceSimulation` |

LPPa E-CID takes priority over every UE-based method when enabled, since it
needs no UE round trip. MS-assisted A-GNSS (raw pseudoranges, an
E-SMLC-side solver) and every other LPPa procedure remain unimplemented —
see `docs/limitations.md`.

## JSON Log API

All logging is structured JSON (`slog.NewJSONHandler`), written to stderr.
`service.log_level` (`debug`/`info`/`warn`/`error`) sets the minimum level.
Every entry has `time`, `level`, and `msg`; `msg` is a stable, dotted event
name safe to match on. The events covering one location transaction:

| Event (`msg`) | Level | Fields | Meaning |
|---|---|---|---|
| `esmlc.lcsap.location_request_received` | Info | `association` | LCS-AP Location Request accepted; positioning job started |
| `esmlc.lpp.event` | Info | `association`, `correlation`, `kind` | An LPP procedure event applied to the job (e.g. capability match, measurement report) |
| `esmlc.lppa.no_active_job` | Warn | `association`, `correlation` | LPPa message received with no matching job (association or job already gone) |
| `esmlc.positioning.job_outcome` | Info | `association`, `outcome` | Job reached a terminal state; see `outcome` values below |
| `esmlc.cell_catalog_reloaded` / `esmlc.cell_catalog_reload_failed` | Info / Warn | `version` or `error`, `active_version`, `records` | Result of a `POST /admin/reload-catalog` call |

`outcome` on `esmlc.positioning.job_outcome` is one of: `estimate_available`,
`measurements_without_estimator`, `estimation_failed`, `quality_not_met`,
`no_eligible_method`, `procedure_failure`, `deadline_expired`, `cancelled`.
Other events (`esmlc.startup`, `esmlc.sls.listening`,
`esmlc.sls.association_up`/`_down`, `esmlc.shutdown`, ...) cover process and
transport lifecycle rather than a specific location transaction.

Example lines (one JSON object per line; reformatted here for readability):

```json
{"time":"2026-08-02T17:20:03.114Z","level":"INFO","msg":"esmlc.lcsap.location_request_received","association":"assoc-7"}
{"time":"2026-08-02T17:20:03.118Z","level":"INFO","msg":"esmlc.lpp.event","association":"assoc-7","correlation":"1a2b3c4d","kind":"MeasurementReport"}
{"time":"2026-08-02T17:20:03.121Z","level":"INFO","msg":"esmlc.positioning.job_outcome","association":"assoc-7","outcome":"estimate_available"}
```

A job where no configured method could produce a fix:

```json
{"time":"2026-08-02T17:20:11.402Z","level":"INFO","msg":"esmlc.positioning.job_outcome","association":"assoc-9","outcome":"no_eligible_method"}
```

An operator-triggered catalog reload — the same fields are also the JSON
HTTP response body of `POST /admin/reload-catalog`, e.g.
`{"ActiveChanged":true,"ActiveVersion":"2026-08-01","RecordCount":842}`:

```json
{"time":"2026-08-02T17:22:00.010Z","level":"INFO","msg":"esmlc.cell_catalog_reloaded","version":"2026-08-01","records":842}
```

## Run

Copy `config/esmlc.yaml.example` to `config/esmlc.yaml`, review it, then run:

```sh
make build
go test ./... -count=1
./esmlc -c config/esmlc.yaml
```

Flags: `-c <path>` selects the YAML config file (default
`config/esmlc.yaml`); `-d` forces DEBUG-level logging to the console
regardless of `service.log_level`, for diagnosing SLs association and
LCS-AP/LPP/LPPa message flow.

The service requires kernel SCTP support for a live association. Unit tests do
not require SCTP, an MME, UE, eNB, GMLC, or live positioning service.

See [architecture](docs/architecture.md) and [limitations](docs/limitations.md).
The source decision is recorded in [the LPP audit](docs/lpp-spec-audit.md).
