# ECID provide-location fixtures

`generate.py` uses the pinned development-only `asn1tools` UPER compiler
against the normalized Release 16.4.0 LPP module. The resulting fixture
metadata is an independent codec oracle; production Go code never reads it.
