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
LPP payload type 0 is recognized and LPPa type 1 is rejected, but neither
payload is interpreted yet. A non-simulated GNSS request returns an LCS
failure rather than inventing a location.

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
