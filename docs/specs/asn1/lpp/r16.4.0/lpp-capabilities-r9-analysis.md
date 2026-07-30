# Release 9 bounded capability payload analysis

Source: normalized `LPP-PDU-Definitions.asn1`, SHA-256
`ef5f33c49cb69ba599db59776d7e58b71da5e00aa788ec57a1333f22e69e2724`.

## Exact wrapper paths

`RequestCapabilities` (line 46) is a SEQUENCE containing mandatory
`criticalExtensions`. Its root CHOICE `c1` is index 0; that CHOICE has four
root alternatives: `requestCapabilities-r9` at index 0 and three spare NULL
alternatives. `criticalExtensionsFuture` is index 1 and unsupported.
`requestCapabilities-r9` is `RequestCapabilities-r9-IEs` (line 57).

`ProvideCapabilities` (line 79) has the analogous path: root
`criticalExtensions.c1[0].provideCapabilities-r9[0]` selects
`ProvideCapabilities-r9-IEs` (line 90). Both IE SEQUENCE types have five root
optional fields in declaration order and an extension marker. Extension-present
is rejected for the initial codec.

## Field inventory and ownership

Both R9 IE types have root optional fields for common IEs, A-GNSS, OTDOA,
ECID, and EPDU. The R13 sensor/TBS/WLAN/Bluetooth group and R16 NR groups are
extension additions and are deferred/rejected.

`A-GNSS-RequestCapabilities` has three mandatory BOOLEAN selectors. Its Provide
form carries nested GNSS support and assistance-data capability lists (50
reachable definitions). `OTDOA-RequestCapabilities` is an extension-only empty
SEQUENCE, while its Provide form has mandatory `otdoa-Mode BIT STRING` plus
optional bounded E-UTRA band lists. `ECID-RequestCapabilities` is an
extension-only empty SEQUENCE. `ECID-ProvideCapabilities` has exactly one root
mandatory field: `ecid-MeasSupported BIT STRING (SIZE(1..8))`, with optional
later extension additions.

Provide capability values are target-device support declarations. This ASN.1
inventory does not establish a location-server capability payload in these R9
fields. Either LPP initiator is structurally possible; procedure direction and
ownership need a clause-level policy review before runtime assumptions change.

## Recommended first subset

Select ECID only: the request selector and the provide measurement-support bit
string. It establishes method-family availability without assistance data,
measurements, estimates, coordinate data, or lists. OTDOA is deferred because
its useful provide declaration introduces variable lists and band constraints;
A-GNSS is deferred because useful support declarations reach nested GNSS lists
and assistance support structures.

The complete R9 graph closure has 101 distinct definitions when all root and
extension capability branches are retained. The selected ECID graph has six
wrapper/payload definitions and no unresolved structural reference. The minimum
fixture closure has the same six named definitions. Canonical machine-readable
counts and hashes are in
`tools/specs/lpp/analysis/r16.4.0/capabilities/closures.json`.

## UPER gap

Existing primitives cover wrapper CHOICE indexes, extension presence, and root
optional bitmaps. The ECID provide field requires a new schema-neutral bounded
variable-size BIT STRING primitive for SIZE(1..8): constrained length followed
by MSB-first bits. Proposed APIs are `ReadBitString(min,max)` and
`WriteBitString(bits,min,max)`. No approximation is permitted. The independent
fixtures cover 1-bit and 3-bit values; negative compiler cases cover zero and
9-bit values.

## Compatibility and proposed model

Existing empty wrapper fixtures remain valid: all selected R9 fields are root
OPTIONAL. Future Go models should preserve this distinction:

```go
type RequestCapabilitiesR9IEs struct { ECIDRequested *struct{} }
type ProvideCapabilitiesR9IEs struct { ECIDMeasurementSupport *BitString }
```

Place these bounded types and their schema functions in `internal/lpp/capability`
or a private sibling consumed by `internal/lpp`; that avoids expanding outer
envelope code and creates no import cycle because capability depends on UPER,
while the outer codec depends on capability. Procedure would carry typed values
but must not select a method. Duplicate fingerprints will need a comparable
canonical bit-string value rather than pointer identity.

Maximum selected payload: 8 content bits plus bounded length and wrapper
bitmaps; no lists, open types, recursion, or attacker-driven allocation are
permitted. Unsupported extension additions and lengths outside 1..8 fail
closed.
