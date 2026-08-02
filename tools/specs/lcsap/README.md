# TS 29.171 Location Response development pipeline

`extract.py` deterministically recovers the six ASN.1 modules that define the
Release 16.4.0 Location Response path from `docs/specs/29171-g40.zip`.  Its
manifest records source and normalized module hashes, DOCX paragraph ranges,
and the supported dependency closure.

Run the complete development-only workflow explicitly:

```sh
tools/specs/lcsap/scripts/extract.sh
tools/specs/lcsap/reference-codec/install.sh
tools/specs/lcsap/scripts/generate-fixtures.sh
tools/specs/lcsap/scripts/verify.sh
```

The reference workflow is never used by production builds or runtime.  It pins
Python `asn1tools` for independently compiling the leaf APER declarations, and
uses the checked-in `generate.py` APER implementation to compose the complete
open-type PDU fixtures.  Go tests compare every complete fixture byte-for-byte.
