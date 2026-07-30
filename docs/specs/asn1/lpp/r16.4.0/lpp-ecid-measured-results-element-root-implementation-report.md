# Root-only MeasuredResultsElement implementation

`internal/lpp/location/result` implements the normalized root sequence only:
extension-present=false, five presence bits in declaration order, then
`physCellId`, optional CellGlobalId, `arfcnEUTRA`, and the remaining present
optionals. It reuses the completed scalar and CellGlobalId codecs without
alignment. Decode rejects extension-present=true.

The immutable options model distinguishes nil/absent from present zero values.
Tests exhaust all 32 optional maps, including present-zero scalar fields, and
cover extension rejection and every-prefix truncation of a root sample. No list,
signal-information, provide envelope, extension map, open type, procedure, or
transport behavior was introduced. Decision A is contingent on the completed
repository validation recorded by this phase.
