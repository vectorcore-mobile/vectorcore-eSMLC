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

## Run

Copy `config/esmlc.yaml.example` to `config/esmlc.yaml`, review it, then run:

```sh
make build
go test ./... -count=1
./esmlc -config config/esmlc.yaml
```

The service requires kernel SCTP support for a live association. Unit tests do
not require SCTP, an MME, UE, eNB, GMLC, or live positioning service.

See [architecture](docs/architecture.md) and [limitations](docs/limitations.md).
The source decision is recorded in [the LPP audit](docs/lpp-spec-audit.md).
