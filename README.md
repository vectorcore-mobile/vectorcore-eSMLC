# VectorCore E-SMLC

E-SMLC (Evolved Serving Mobile Location Centre) is the component in an
LTE network responsible for finding the physical location of a phone, for
example for emergency calls (E911) or other lawful location requests. It
talks to the carrier's MME (the network's control-plane server) to receive a
request to locate a specific phone, works out that phone's position, and
sends the answer back.


## Features

- SCTP-based SLs endpoint (PPID 29, stream 0) implementing LCS-AP
  Location Request/Response, failure responses, and Reset acknowledgement
- Connection-Oriented Information parsing in the deployed MME form
- LPP payload handling with real, catalog-backed estimators:
  - **E-CID** — coarse surveyed serving-cell point
  - **OTDOA** — 2D hyperbolic multilateration from RSTD measurements
  - **A-GNSS** (UE-based, GPS only) — validates and relays the UE-reported
    position
- LPPa E-CID Measurement Initiation/Report support (E-SMLC↔eNB, TS 36.455),
  opt-in and prioritized over UE-based methods when enabled
- Structured JSON logging (`slog.NewJSONHandler`) for every location
  transaction
- Optional observability HTTP listener exposing Prometheus-format
  `/metrics`, `/healthz`, `/readyz`, and `POST /admin/reload-catalog`
- Test-only simulation mode for a fixed, operator-configured coordinate,
  disabled by default

## Build Requirements

- Go 1.26.0+
- `CGO_ENABLED=0` (no native bridge, FFI, native ASN.1 runtime, or external
  process required)
- Kernel SCTP support (only needed to run a live association; unit tests do
  not require SCTP, an MME, UE, eNB, or GMLC)

## Build

```sh
make clean
make
```
