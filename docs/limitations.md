# Limitations and live-test boundary

This recovery decodes LPP payload-type 0 APDUs through the LCS-AP
Connection-Oriented Information carrier and applies the bounded LPP procedure.
It supports the bounded ECID capability and request/provide reporting branches,
a limited capability-gated ECID initiation workflow, and an honest raw-method
result to LCS-outcome boundary. It can produce a coarse authoritative
serving-cell reference estimate only when a fresh operator-supplied catalog is
configured; it has no RF-ranging ECID estimator. Otherwise a received ECID
report produces an estimator-unavailable failure unless the explicitly enabled,
labelled simulation estimator is used for development. It does not yet have
QoS, permitted-method, UE/eNB, or assistance inputs sufficient for general
method selection. GNSS assistance, LPPa, OTDOA, connectionless
information, Error Indication, Location Abort response semantics, privacy
authorization, GMLC, and SLg remain unsupported. No actual GNSS location is
derived.

Live interoperability with a UE/eNB has not been validated. The covered MME
interoperability boundary includes the exact generic SLs APER PDU container,
bounded LPP APDU decoding, correlation round trip, and generated LPP procedure
actions. A live trial still requires complete LCS request semantics, real
UE/eNB capabilities and measurements, and a standards-derived estimate before
simulation is disabled.
