# Development-only UPER reference codec

| Candidate | Result | Notes |
|---|---|---|
| `asn1tools` 0.167.0 | selected for attempted validation | MIT; Python; exposes `uper`; pinned in `requirements-lock.txt` |
| `asn1c` | not selected initially | no installed binary; C-generated output would remain development-only but adds a heavier toolchain |

Install explicitly (network may be needed once):

```bash
tools/specs/lpp/reference-codec/install.sh
```

Then compile only with the repository-local environment:

```bash
tools/specs/lpp/reference-codec/.venv/bin/python tools/specs/lpp/reference-codec/compile.py
```

Validate generated fixtures explicitly with `tools/specs/lpp/reference-codec/test.sh`.

The compiler is never invoked by `go build`, `go test`, `make build`, `make test`, or `make vet`. It is not linked into the E-SMLC and cannot be used at runtime. The source inputs are corrected V2 normalized ASN.1 only.
