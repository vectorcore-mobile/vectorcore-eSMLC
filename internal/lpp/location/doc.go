// Package location implements bounded TS 37.355 R9 ECID location payloads.
//
// It supports only criticalExtensions.c1[0].requestLocationInformation-r9[0]
// and the ECID root optional field. The five R9 root optional bits remain in
// normative order (common, A-GNSS, OTDOA, ECID, EPDU); all but ECID fail
// closed. Requested measurements are packed MSB-first in an immutable
// uper.BitString and retain their exact SIZE(1..8) bit length.
//
// The matching provide path supports root-only ECID measurement information:
// an optional primary-cell result and a bounded MeasuredResultsList of 1..32
// root-only result elements. It preserves non-aligned UPER composition and
// rejects ECID errors and every extension addition. A-GNSS, OTDOA, common
// location estimates, EPDU, result extensions and provide-side positioning
// policy remain unsupported.
package location
