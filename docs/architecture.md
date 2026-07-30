# Architecture

The `cmd/esmlc` process loads validated YAML, owns shutdown signals, and runs
the SCTP server. Each association has a bounded read buffer and serialized
writer. Incoming PPID 29 messages are decoded by `internal/lcsap`, which uses
the bounded `internal/aper` primitives (constrained integers, criticality and
open types). Unknown reject-criticality IEs and malformed input are rejected.

The LCS-AP PDU/IE encoding intentionally follows the pure-Go MME SLs APER
layout. Connection-Oriented Information is procedure 1 with an initiating
message and TS 29.171 `reject` procedure criticality. The deployed MME emits
`ignore`, which remains accepted only as an inbound compatibility exception.
It uses Correlation-ID (2), Payload-Type (15), and APDU (1), all reject
criticality. LPP is payload-type 0 and LPPa is 1. Routing-ID is not
part of this LPP procedure; it is used by the MME's separate UE-associated
LPPa path. Payload-type 0 is decoded as an LPP transport octet string,
scoped by association plus Correlation-ID, and applied to the transport-neutral
LPP procedure. Procedure actions are encoded back into LPP and carried in a
Connection-Oriented Information response with the same LCS correlation. LPPa
remains unsupported.

`internal/session` keys state by association identity plus Correlation-ID and
uses an exactly-once terminal operation. Association loss, reset, and shutdown
remove sessions. The current synchronous recovered provider does not leave
long-lived sessions, but the manager provides the bounded lifecycle boundary
needed by later real positioning work.

## Positioning protocol boundary

`internal/lpp` is a transport-neutral UPER codec and
`internal/lpp/procedure` is transport-neutral transaction orchestration. The
currently implemented method branch is bounded root ECID request/provide
reporting, including typed measurement results; neither package selects ECID,
executes a measurement, or calculates a position. Future method selection must
live above this boundary and take requested QoS, response time, UE/eNB/network
capabilities, assistance availability, permitted methods, operator policy and
prior method outcomes as inputs. It may select ECID, OTDOA, A-GNSS or another
supported method without encoding that policy into an ASN.1 codec.

SLs/LCS-AP remains the carrier boundary. It dispatches payload type 0 APDUs to
the LPP procedure, returns generated actions such as acknowledgements, and
creates a positioning job for a recovered Location Request. The current request
model has only correlation, location type, and serving ECGI, so it cannot claim
QoS-driven method selection. It can initiate ECID only through explicit local
policy followed by a matching UE capability result; otherwise it returns the
existing LCS failure response.

`internal/positioning` owns that next-layer job lifecycle. An incoming LCS
Location Request supplies only the recovered correlation, location type and
serving ECGI. If local operator policy explicitly enables ECID, the job starts
LPP ECID capability discovery; only an actual UE ECID capability result that
satisfies the configured requested-measurement bitmap can start ECID
RequestLocationInformation. Without an enabled eligible method, the server
returns the existing LCS failure response. No capability, QoS, or method is
invented from configuration.

Completed ECID measurements are retained as raw method results, not location
estimates. They pass through a job-scoped estimator boundary: with no
authoritative estimator the job returns a correlated LCS failure; an explicitly
enabled local simulation estimator may exercise successful LCS estimate
delivery and is always labelled as simulation.

When configured with a fresh operator-maintained serving-cell catalog, the
ECID estimator can instead return the surveyed serving-cell reference point
with its catalog-provided conservative GAD uncertainty. It does not use RSRP,
RSRQ, or Rx-Tx as ranging inputs.

The LCS request boundary preserves priority, root LCS-QoS, and the UE LPP
capability when supplied. LPP capability false excludes ECID before initiation;
horizontal QoS and vertical requirements are checked against the estimate only
when the corresponding metadata exists. The LCS response boundary maps typed
terminal outcomes to the supported root LCS-Cause subset without exposing
internal errors.
