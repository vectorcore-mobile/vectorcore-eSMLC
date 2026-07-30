# Procedure ECID capability integration

`internal/lpp/procedure` carries the existing `internal/lpp/capability` values
without reinterpreting the BIT STRING. `StartOptions.RequestECID` constructs the
R9 ECID request selector; `ProvideCapabilitiesResult` accepts a typed
`ProvideCapabilitiesR9IEs` value. Inbound request/provide events carry copied
typed values.

Pending application state is bounded and records whether ECID was requested.
The conservative procedure policy rejects an unsolicited ECID response; an
empty response remains valid for either request form. Abort, Error, completion,
and prune retain the existing pending-state cleanup behavior. The transaction
store continues to own sequences, acknowledgements, terminal state, and
duplicate fingerprints.

There is no method selection, capability persistence, location-information
payload, timer, transport, or positioning behavior. The maximum added pending
state is one boolean plus the typed response supplied only in an action/event;
the ECID bit string is bounded to eight bits.
