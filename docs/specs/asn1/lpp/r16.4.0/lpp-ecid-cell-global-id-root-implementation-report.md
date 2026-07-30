# Root-only CellGlobalIdEUTRA-AndUTRA implementation

Implemented in `internal/lpp/location/result`: bounded PLMN digits, exact MCC,
length-preserving MNC, PLMN identity, fixed 28-bit E-UTRA and 32-bit UTRA
identities, an exclusive CHOICE, and the extensible wrapper's root form.

Encoding is `extension=false | MCC | MNC count+digits | choice | payload` with
no alignment. Decoding rejects extension=true. Fixtures reproduce the four wire
length classes (51, 55, 55, and 59 bits), including `001/01`, `310/260`, and
both identity alternatives. No measured-result aggregate, list, open type,
transport, or runtime tooling was added. Decision A: root-only CellGlobalId is
ready for a separately bounded root-only MeasuredResultsElement phase.
